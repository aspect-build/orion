package gazelle

import (
	"fmt"
	"os"
	"path"
	"sync"

	common "github.com/aspect-build/silo/cli/core/gazelle/common"
	BazelLog "github.com/aspect-build/silo/cli/core/pkg/logger"
	plugin "github.com/aspect-build/silo/cli/pro/gazelle/host/plugin"
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
	cfg := args.Config.Exts[GazelleLanguageName].(*BUILDConfig).GetConfig(args.Rel)

	// Generating new BUILDs may disabled.
	if cfg.GenerationMode() == common.GenerationModeUpdate && args.File == nil {
		BazelLog.Tracef("GenerateRules(%s) BUILD creation disabled: %s", GazelleLanguageName, args.Rel)
		return gazelleLanguage.GenerateResult{}
	}

	BazelLog.Tracef("GenerateRules(%s): %s", GazelleLanguageName, args.Rel)

	// TODO: normally would...
	//   1. collect "source files"
	//   2. generate rules for groups of "source rules"
	// Now or maybe later...
	//   3. parse "source files"
	//   4. persist source file imports + symbols

	// Collect source files grouped by plugins consuming them.
	// Recurse if subdirectories will not generate their own BUILD files
	sourceFilesByPlugin := host.collectSourceFilesByPlugin(cfg, args)

	pluginTargets := make(map[string][]plugin.TargetDeclaration)
	pluginTargetsLock := sync.Mutex{}

	// Parse and query source files
	// TODO: only parse source files once, not once per plugin (most likely they only belong to one plugin though)
	eg := errgroup.Group{}
	eg.SetLimit(10)
	for pluginId, files := range sourceFilesByPlugin {
		prep := cfg.pluginPrepareResults[pluginId]

		eg.Go(func() error {
			// Collect source files and plugin metadata for those sources.
			targetSources := host.collectPluginTargetSources(pluginId, prep, args.Dir, files)

			// Analyze the source file metadata for this plugin.
			host.analyzePluginTargetSources(pluginId, prep, targetSources)

			// Use the collected sources and analysis to generate rules
			targets := host.generateTargets(pluginId, prep, targetSources)

			// Lock for the assignment into the cross-thread pluginTargets
			pluginTargetsLock.Lock()
			defer pluginTargetsLock.Unlock()
			pluginTargets[pluginId] = targets

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		BazelLog.Errorf("Unknown GenerateRules(%s) error: %v", GazelleLanguageName, err)
	}

	return host.convertPlugTargetsToGenerateResult(pluginTargets, args)
}

func (host *GazelleHost) convertPlugTargetsToGenerateResult(pluginTargets map[string][]plugin.TargetDeclaration, args gazelleLanguage.GenerateArgs) gazelleLanguage.GenerateResult {
	var result gazelleLanguage.GenerateResult

	for pluginId, declareResults := range pluginTargets {
		for _, target := range declareResults {
			// If marked for removal simply add to the empty list and continue
			if target.Remove {
				BazelLog.Debugf("GenerateRules remove target: %s %s(%q)", args.Rel, target.Kind, target.Name)
				result.Empty = append(result.Empty, gazelleRule.NewRule(target.Kind, target.Name))
				continue
			}

			// Check for name-collisions with the rule being generated.
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
		}
	}

	return result
}

func convertPluginTargetDeclaration(args gazelleLanguage.GenerateArgs, pluginId string, target plugin.TargetDeclaration) *gazelleRule.Rule {
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

func (host *GazelleHost) collectPluginTargetSources(pluginId string, prep pluginConfig, baseDir string, pluginSrcs []string) []plugin.TargetSource {
	targetSources := make([]plugin.TargetSource, len(pluginSrcs))

	eg := errgroup.Group{}
	eg.SetLimit(100)

	for i, f := range pluginSrcs {
		eg.Go(func() error {
			queryResults, err := runPluginQueries(prep, baseDir, f)
			if err != nil {
				msg := fmt.Sprintf("Error querying source file %q: %v", f, err)
				fmt.Printf("%s\n", msg)
				BazelLog.Errorf(msg)
			}

			src := plugin.TargetSource{
				Path:         f,
				QueryResults: queryResults,
			}

			// Assignment into a pre-allocated slice is safe across goroutines
			// and maintains order of the source files.
			targetSources[i] = src

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		BazelLog.Errorf("Collect plugin sources error: %v", err)
	}

	return targetSources
}

func runPluginQueries(prep pluginConfig, baseDir, f string) (plugin.QueryResults, error) {
	queries := prep.GetQueriesForFile(f)
	if len(queries) == 0 {
		return nil, nil
	}

	sourceCode, err := os.ReadFile(path.Join(baseDir, f))
	if err != nil {
		return nil, err
	}

	// Split queries by type to invoke in batches
	queriesByType := make(map[*plugin.QueryProcessor]plugin.NamedQueries)
	for key, query := range queries {
		if queriesByType[&query.Processor] == nil {
			queriesByType[&query.Processor] = make(plugin.NamedQueries)
		}
		queriesByType[&query.Processor][key] = query
	}

	queryResultsChan := make(chan *plugin.QueryProcessorResult)
	wg := sync.WaitGroup{}

	for processor, queries := range queriesByType {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if err := (*processor)(f, sourceCode, queries, queryResultsChan); err != nil {
				msg := fmt.Sprintf("Error running queries for %q: %v", f, err)
				fmt.Printf("%s\n", msg)
				BazelLog.Errorf(msg)
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

	return queryResults, nil
}

// Collect source files managed by this BUILD and batch them by plugins interested in them.
func (host *GazelleHost) collectSourceFilesByPlugin(cfg *BUILDConfig, args gazelleLanguage.GenerateArgs) map[string][]string {
	sourceFilesByPlugin := make(map[string][]string)

	// Collect source files managed by this BUILD for each plugin.
	common.GazelleWalkDir(args, host.gitignore.Matches, func(f string) error {
		for pluginId, p := range cfg.pluginPrepareResults {
			for _, s := range p.Sources {
				if s.Match(f) {
					if sourceFilesByPlugin[pluginId] == nil {
						sourceFilesByPlugin[pluginId] = make([]string, 0, 1)
					}
					sourceFilesByPlugin[pluginId] = append(sourceFilesByPlugin[pluginId], f)
					break
				}
			}
		}

		return nil
	})

	return sourceFilesByPlugin
}

// Let plugins analyze sources and declare their outputs
func (host *GazelleHost) analyzePluginTargetSources(pluginId string, prep pluginConfig, sources []plugin.TargetSource) {
	eg := errgroup.Group{}
	eg.SetLimit(100)

	for _, src := range sources {
		eg.Go(func() error {
			actx := plugin.NewAnalyzeContext(&src, host.database)

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
func (host *GazelleHost) generateTargets(pluginId string, prep pluginConfig, targetSources []plugin.TargetSource) []plugin.TargetDeclaration {
	ctx := plugin.DeclareTargetsContext{
		PrepareContext: prep.PrepareContext,
		Sources:        targetSources,
		Targets:        plugin.NewDeclareTargetActions(),
	}

	host.plugins[pluginId].DeclareTargets(ctx)

	return ctx.Targets.Targets()
}
