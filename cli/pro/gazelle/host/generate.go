package gazelle

import (
	"crypto"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	common "github.com/aspect-build/silo/cli/core/gazelle/common"
	"github.com/aspect-build/silo/cli/core/gazelle/common/cache"
	BazelLog "github.com/aspect-build/silo/cli/core/pkg/logger"
	"github.com/aspect-build/silo/cli/pro/gazelle/host/plugin"
	gazelleLabel "github.com/bazelbuild/bazel-gazelle/label"
	gazelleLanguage "github.com/bazelbuild/bazel-gazelle/language"
	gazelleRule "github.com/bazelbuild/bazel-gazelle/rule"
	"golang.org/x/sync/errgroup"
)

const (
	// TODO: move to common
	MaxWorkerCount = 12
)

const (
	targetAttrImports    = "__target_attr_imports"
	targetAttrValues     = "__target_attr_values"
	targetDeclarationKey = "__target_declaration"
	targetPluginKey      = "__target_plugin"
)

// Gazelle GenerateRules phase - declare:
//   - which rules to delete (GenerateResult.Empty)
//   - which rules to create (or merge with existing) and their associated metadata (GenerateResult.Gen + GenerateResult.Imports)
func (host *GazelleHost) GenerateRules(args gazelleLanguage.GenerateArgs) gazelleLanguage.GenerateResult {
	BazelLog.Tracef("GenerateRules(%s): %s", GazelleLanguageName, args.Rel)

	cfg := args.Config.Exts[GazelleLanguageName].(*BUILDConfig)

	queryCache := cache.Get[plugin.QueryResults](args.Config)

	// TODO: normally would...
	//   1. collect "source files"
	//   2. generate rules for groups of "source rules"
	// Now or maybe later...
	//   3. parse "source files"
	//   4. persist source file imports + symbols

	// Collect source files grouped by plugins consuming them.
	// Recurse if subdirectories will not generate their own BUILD files
	sourceFilesByPlugin, sourceFilePlugins := host.collectSourceFilesByPlugin(cfg, args)

	// Run queries on source files and collect results
	eg := errgroup.Group{}
	eg.SetLimit(100)

	sourceFileQueryResults := make(map[string]plugin.QueryResults)
	sourceFileQueryResultsLock := sync.Mutex{}

	// Parse and query source files
	for sourceFile, pluginIds := range sourceFilePlugins {
		// Collect all queries for this source file from all plugins
		queries := make(plugin.NamedQueries)
		for _, pluginId := range pluginIds {
			prep := cfg.pluginPrepareResults[pluginId]
			for queryId, query := range prep.GetQueriesForFile(sourceFile) {
				queries[fmt.Sprintf("%s|%s", pluginId, queryId)] = query
			}
		}

		if len(queries) == 0 {
			continue
		}

		eg.Go(func() error {
			queryResults, err := host.runSourceQueries(queryCache, queries, args.Dir, sourceFile)
			if err != nil {
				msg := fmt.Sprintf("Querying source file %q: %v", path.Join(args.Rel, sourceFile), err)
				fmt.Printf("%s\n", msg)
				BazelLog.Error(msg)
				return nil
			}

			sourceFileQueryResultsLock.Lock()
			defer sourceFileQueryResultsLock.Unlock()
			sourceFileQueryResults[sourceFile] = queryResults
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		BazelLog.Errorf("Collect plugin sources error: %v", err)
	}

	pluginTargetActions := make(map[plugin.PluginId][]plugin.TargetAction)
	pluginTargetsLock := sync.Mutex{}

	// Loop over all plugins
	for pluginId, prep := range cfg.pluginPrepareResults {
		pluginSrcs := sourceFilesByPlugin[pluginId]

		queryPrefix := fmt.Sprintf("%s|", pluginId)

		// Collect the query results for this plugin's source files
		targetSources := make([]plugin.TargetSource, 0, len(pluginSrcs))
		for _, f := range pluginSrcs {
			queryResults := make(plugin.QueryResults)
			for queryId, results := range sourceFileQueryResults[f] {
				if strings.HasPrefix(queryId, queryPrefix) {
					queryResults[queryId[len(queryPrefix):]] = results
				}
			}

			targetSources = append(targetSources, plugin.TargetSource{
				Path:         f,
				QueryResults: queryResults,
			})
		}

		eg.Go(func() error {
			// Analyze the source file metadata for this plugin.
			host.analyzePluginTargetSources(pluginId, prep, targetSources)

			// Use the collected sources and analysis to generate rules
			actions := host.generateTargets(pluginId, prep, targetSources)

			// Lock for the assignment into the cross-thread pluginTargets
			pluginTargetsLock.Lock()
			defer pluginTargetsLock.Unlock()
			pluginTargetActions[pluginId] = actions

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		BazelLog.Errorf("Unknown GenerateRules(%s) error: %v", GazelleLanguageName, err)
	}

	return host.convertPlugActionsToGenerateResult(pluginTargetActions, args)
}

func applyRemoveAction(args gazelleLanguage.GenerateArgs, result *gazelleLanguage.GenerateResult, rm plugin.RemoveTargetAction) *gazelleRule.Rule {
	if args.File == nil {
		return nil
	}

	for _, r := range args.File.Rules {
		if r.Name() == rm.Name {
			kind := rm.Kind
			if rm.Kind == "" {
				kind = r.Kind() // TODO: need to reverse map_kind?
			}
			result.Empty = append(result.Empty, gazelleRule.NewRule(kind, r.Name()))
			return r
		}
	}
	return nil
}

func (host *GazelleHost) convertPlugActionsToGenerateResult(pluginActions map[string][]plugin.TargetAction, args gazelleLanguage.GenerateArgs) gazelleLanguage.GenerateResult {
	var result gazelleLanguage.GenerateResult

	// Iterate over the pluginIds[] in a deterministic order
	// instead of iterating over the plugins[] or pluginActions[pluginId] map
	for _, pluginId := range host.pluginIds {
		for _, action := range pluginActions[pluginId] {
			host.applyPluginAction(args, pluginId, action, &result)
		}
	}

	return result
}

func (host *GazelleHost) applyPluginAction(args gazelleLanguage.GenerateArgs, pluginId plugin.PluginId, action plugin.TargetAction, result *gazelleLanguage.GenerateResult) {
	switch action.(type) {
	case plugin.RemoveTargetAction:
		// If marked for removal simply add to the empty list and continue
		if removed := applyRemoveAction(args, result, action.(plugin.RemoveTargetAction)); removed != nil {
			BazelLog.Debugf("GenerateRules remove target: %s %s(%q)", args.Rel, removed.Kind(), removed.Name())
		}
	case plugin.AddTargetAction:
		// Check for name-collisions with the rule being generated.
		target := action.(plugin.AddTargetAction).TargetDeclaration
		colError := common.CheckCollisionErrors(target.Name, target.Kind, host.sourceRuleKinds, args)
		if colError != nil {
			fmt.Fprintf(os.Stderr, "Source rule generation error: %v\n", colError)
			os.Exit(1)
		}

		// Generate the gazelle Rule to be added/merged into the BUILD file.
		rule := convertPluginTargetDeclaration(args, pluginId, target)

		result.Gen = append(result.Gen, rule)
		result.Imports = append(result.Imports, rule.PrivateAttr(targetAttrImports))

		BazelLog.Tracef("GenerateRules(%s) add target: %s %s(%q)", GazelleLanguageName, args.Rel, target.Kind, target.Name)
	default:
		BazelLog.Fatalf("Unknown plugin action type: %T", action)
	}
}

func convertPluginTargetDeclaration(args gazelleLanguage.GenerateArgs, pluginId plugin.PluginId, target plugin.TargetDeclaration) *gazelleRule.Rule {
	targetRule := gazelleRule.NewRule(target.Kind, target.Name)

	ruleImports := make(map[string][]plugin.TargetImport, 0)
	ruleAttrs := make(map[string]interface{}, 0)

	targetRule.SetPrivateAttr(targetPluginKey, pluginId)
	targetRule.SetPrivateAttr(targetDeclarationKey, target)
	targetRule.SetPrivateAttr(targetAttrImports, ruleImports)
	targetRule.SetPrivateAttr(targetAttrValues, ruleAttrs)

	for attr, val := range target.Attrs {
		attrValue, attrImports := convertPluginAttribute(args, val)

		// Record imports assigned to this attribute
		if attrImports != nil && len(attrImports) > 0 {
			// TODO: verify 'attr' is resolveable if len(attrImports) > 0
			ruleImports[attr] = attrImports
		}

		// Record and set values assigned to this attribute (of any type).
		// This may be merged with imports during the resolution stage.
		if attrValue != nil {
			ruleAttrs[attr] = attrValue
			targetRule.SetAttr(attr, attrValue)
		}
	}

	return targetRule
}

func convertPluginAttribute(args gazelleLanguage.GenerateArgs, val interface{}) (interface{}, []plugin.TargetImport) {
	if a, isArray := val.([]interface{}); isArray {
		r := make([]interface{}, 0, len(a))
		i := make([]plugin.TargetImport, 0)
		for _, v := range a {
			newR, newI := convertPluginAttribute(args, v)
			if newR != nil {
				r = append(r, newR)
			}
			if newI != nil {
				i = append(i, newI...)
			}
		}
		if len(r) == 0 {
			return nil, i
		}
		return r, i
	}

	if targetImport, isImport := val.(plugin.TargetImport); isImport {
		return nil, []plugin.TargetImport{targetImport}
	}

	// Convert plugin.Label to a gazelle Label
	if l, isLabel := val.(plugin.Label); isLabel {
		val = gazelleLabel.New(l.Repo, l.Pkg, l.Name)
	}

	// Normalize gazelle labels to be relative to the BUILD file
	if l, isLabel := val.(gazelleLabel.Label); isLabel {
		// TODO: also convert the `args.Config.RepoName` repo to relative?
		return l.Rel("", args.Rel), nil
	}

	return val, nil
}

func init() {
	// Ensure types used in cache key computation are known to the gob encoder
	gob.Register(plugin.NamedQueries{})
	gob.Register(plugin.QueryDefinition{})
	gob.Register(plugin.QueryType(""))
	gob.Register(plugin.AstQueryParams{})
	gob.Register(plugin.RegexQueryParams(""))
	gob.Register(plugin.JsonQueryParams(""))
}

func computeQueriesCacheKey(sourceCode []byte, queries plugin.NamedQueries) (string, bool) {
	cacheDigest := crypto.MD5.New()
	cacheDigest.Write(sourceCode)

	e := gob.NewEncoder(cacheDigest)
	if err := e.Encode(queries); err != nil {
		return "", false
	}

	return hex.EncodeToString(cacheDigest.Sum(nil)), true
}

func (host *GazelleHost) runSourceQueries(queryCache cache.Cache, queries plugin.NamedQueries, baseDir, f string) (plugin.QueryResults, error) {
	// Read the file content
	sourceCode, err := os.ReadFile(path.Join(baseDir, f))
	if err != nil {
		return nil, err
	}

	queryCacheKey, queryingCacheable := computeQueriesCacheKey(sourceCode, queries)
	if queryCache != nil && queryingCacheable {
		if cachedResults, found := queryCache.Load(queryCacheKey); found {
			return cachedResults.(plugin.QueryResults), nil
		}
	}

	// Split queries by type to invoke in batches
	queriesByType := make(map[plugin.QueryType]plugin.NamedQueries)
	for key, query := range queries {
		if queriesByType[query.QueryType] == nil {
			queriesByType[query.QueryType] = make(plugin.NamedQueries)
		}
		queriesByType[query.QueryType][key] = query
	}

	queryResultsChan := make(chan *plugin.QueryProcessorResult)
	wg := sync.WaitGroup{}

	for queryType, queries := range queriesByType {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := plugin.RunQueries(queryType, f, sourceCode, queries, queryResultsChan); err != nil {
				msg := fmt.Sprintf("Error running queries for %q: %v", f, err)
				fmt.Printf("%s\n", msg)
				BazelLog.Error(msg)
			}
		}()
	}

	go func() {
		wg.Wait()
		close(queryResultsChan)
	}()

	// Read the result channel and collect the results
	queryResults := make(plugin.QueryResults)
	for result := range queryResultsChan {
		queryResults[result.Key] = result.Result
	}

	if queryCache != nil && queryingCacheable {
		queryCache.Store(queryCacheKey, queryResults)
	}

	return queryResults, nil
}

// Collect source files managed by this BUILD and batch them by plugins interested in them.
func (host *GazelleHost) collectSourceFilesByPlugin(cfg *BUILDConfig, args gazelleLanguage.GenerateArgs) (map[plugin.PluginId][]string, map[string][]plugin.PluginId) {
	sourceFilesByPlugin := make(map[plugin.PluginId][]string)
	sourceFilePlugins := make(map[string][]plugin.PluginId)

	// Collect source files managed by this BUILD for each plugin.
	common.GazelleWalkDir(args, func(f string) error {
		for pluginId, p := range cfg.pluginPrepareResults {
			for _, s := range p.Sources {
				if s.Match(f) {
					if sourceFilesByPlugin[pluginId] == nil {
						sourceFilesByPlugin[pluginId] = make([]string, 0, 1)
					}
					sourceFilesByPlugin[pluginId] = append(sourceFilesByPlugin[pluginId], f)
					sourceFilePlugins[f] = append(sourceFilePlugins[f], pluginId)
					break
				}
			}
		}

		return nil
	})

	return sourceFilesByPlugin, sourceFilePlugins
}

// Let plugins analyze sources and declare their outputs
func (host *GazelleHost) analyzePluginTargetSources(pluginId plugin.PluginId, prep pluginConfig, sources []plugin.TargetSource) {
	eg := errgroup.Group{}
	eg.SetLimit(100)

	for _, src := range sources {
		eg.Go(func() error {
			actx := plugin.NewAnalyzeContext(prep.PrepareContext, &src, host.database)

			err := host.plugins[pluginId].Analyze(actx)
			if err != nil {
				// TODO:
				fmt.Println(fmt.Errorf("analyze failed for %s: %w", pluginId, err))
			}
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		BazelLog.Errorf("Analyze plugin error: %v", err)
	}
}

// Let plugins declare any targets they want to generate for the target sources.
func (host *GazelleHost) generateTargets(pluginId plugin.PluginId, prep pluginConfig, targetSources []plugin.TargetSource) []plugin.TargetAction {
	ctx := plugin.NewDeclareTargetsContext(
		prep.PrepareContext,
		targetSources,
		plugin.NewDeclareTargetActions(),
		host.database,
	)

	return host.plugins[pluginId].DeclareTargets(ctx).Actions
}
