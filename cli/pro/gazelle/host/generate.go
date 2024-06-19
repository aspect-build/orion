package gazelle

import (
	"fmt"
	"os"
	"path"

	gazelle "github.com/aspect-build/silo/cli/core/gazelle/common"
	treeutils "github.com/aspect-build/silo/cli/core/gazelle/common/treesitter"
	BazelLog "github.com/aspect-build/silo/cli/core/pkg/logger"
	plugin "github.com/aspect-build/silo/cli/pro/gazelle/host/plugin"
	gazelleLanguage "github.com/bazelbuild/bazel-gazelle/language"
	gazelleRule "github.com/bazelbuild/bazel-gazelle/rule"
)

const (
	// TODO: move to common
	MaxWorkerCount = 12
)

const (
	targetDeclarationKey = "__target_declaration"
	targetPluginKey      = "__target_plugin"
)

// Gazelle GenerateRules phase - declare:
//   - which rules to delete (GenerateResult.Empty)
//   - which rules to create (or merge with existing) and their associated metadata (GenerateResult.Gen + GenerateResult.Imports)
func (host *GazelleHost) GenerateRules(args gazelleLanguage.GenerateArgs) gazelleLanguage.GenerateResult {
	BazelLog.Tracef("GenerateRules: %q", args.Rel)

	cfg := args.Config.Exts[GazelleLanguageName].(*BUILDConfig).GetConfig(args.Rel)

	// All generation may disabled.
	if cfg.GenerationMode() == GenerationModeNone {
		BazelLog.Tracef("GenerateRules disabled: %q", args.Rel)
		return gazelleLanguage.GenerateResult{}
	}

	// Generating new BUILDs may disabled.
	if cfg.GenerationMode() == GenerationModeUpdate && args.File == nil {
		BazelLog.Tracef("GenerateRules BUILD creation disabled: %q", args.Rel)
		return gazelleLanguage.GenerateResult{}
	}

	// TODO: normally would...
	//   1. collect "source files"
	//   2. generate rules for groups of "source rules"
	// Now or maybe later...
	//   3. parse "source files"
	//   4. persist source file imports + symbols

	// Collect source files grouped by plugins consuming them.
	// Recurse if subdirectories will not generate their own BUILD files
	sourceFilesByPlugin := host.collectSourceFilesByPlugin(cfg, args, cfg.GenerationMode() == GenerationModeUpdate)

	pluginTargets := make(map[string][]plugin.TargetDeclaration)

	// Parse and query source files
	// TODO: only parse source files once, not once per plugin (most likely they only belong to one plugin though)
	// TODO: parallelize
	for pluginId, files := range sourceFilesByPlugin {
		prep := cfg.pluginPrepareResults[pluginId]

		// Collect source files and plugin metadata for those sources
		targetSources := host.collectPluginTargetSources(pluginId, prep, args.Dir, files)

		// Analyze the source file metadata
		host.analyzePluginTargetSources(pluginId, prep, targetSources)

		// Use the collected sources and analysis to generate rules
		pluginTargets[pluginId] = host.generateTargets(pluginId, prep, targetSources)
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
			colError := gazelle.CheckCollisionErrors(target.Name, target.Kind, host.sourceRuleKinds, args)
			if colError != nil {
				fmt.Fprintf(os.Stderr, "Source rule generation error: %v\n", colError)
				os.Exit(1)
			}

			// Generate the gazelle Rule to be added/merged into the BUILD file.
			targetRule := gazelleRule.NewRule(target.Kind, target.Name)
			targetRule.SetPrivateAttr(targetPluginKey, pluginId)
			targetRule.SetPrivateAttr(targetDeclarationKey, target)

			for attr, val := range target.Attrs {
				targetRule.SetAttr(attr, val)
			}

			BazelLog.Tracef("GenerateRules add target: %s %s(%q)", args.Rel, target.Kind, target.Name)

			result.Gen = append(result.Gen, targetRule)
			result.Imports = append(result.Imports, target.Imports)
		}
	}

	return result
}

func (host *GazelleHost) collectPluginTargetSources(pluginId string, prep pluginConfig, baseDir string, pluginSrcs []string) []plugin.TargetSource {
	targetSources := make([]plugin.TargetSource, 0, len(pluginSrcs))

	// TODO: parallelize
	for _, f := range pluginSrcs {
		queryResults, err := runPluginQueries(prep, baseDir, f)
		if err != nil {
			BazelLog.Errorf("Error querying source file %q: %v", f, err)
		}

		src := plugin.TargetSource{
			Path:         f,
			QueryResults: queryResults,
		}
		targetSources = append(targetSources, src)
	}

	return targetSources
}

func runPluginQueries(prep pluginConfig, baseDir, f string) (plugin.QueryResults, error) {
	queries := prep.GetQueriesForFile(f)
	if len(queries) == 0 {
		return nil, nil
	}

	lang := treeutils.PathToLanguage(f)

	sourceCode, err := os.ReadFile(path.Join(baseDir, f))
	if err != nil {
		return nil, err
	}

	ast, err := treeutils.ParseSourceCode(lang, f, sourceCode)
	if err != nil {
		return nil, err
	}

	// Parse errors. Only log them due to many false positives.
	// TODO: what false positives? See js plugin where this is from
	if BazelLog.IsLevelEnabled(BazelLog.TraceLevel) {
		treeErrors := ast.QueryErrors()
		if treeErrors != nil {
			BazelLog.Tracef("TreeSitter query errors: %v", treeErrors)
		}
	}

	queryResults := make(plugin.QueryResults)

	// TODO: parallelize - run queries concurrently
	for key, query := range queries {
		resultCh := ast.Query(query.Query)

		// TODO: delay collection from channel until first read?
		// Then it must be cached for later reads...
		match := make([]plugin.QueryMatch, 0, 1)
		for r := range resultCh {
			match = append(match, plugin.NewQueryMatch(r.Captures()))
		}

		queryResults[key] = plugin.NewQueryMatches(&match)
	}

	return queryResults, nil
}

// Collect source files managed by this BUILD and batch them by plugins interested in them.
func (host *GazelleHost) collectSourceFilesByPlugin(cfg *BUILDConfig, args gazelleLanguage.GenerateArgs, recurse bool) map[string][]string {
	sourceFilesByPlugin := make(map[string][]string)

	excludes, has_excludes := cfg.directiveRawValues["excludes"]
	if !has_excludes {
		excludes = []string{}
	}

	// Collect source files managed by this BUILD for each plugin.
	gazelle.GazelleWalkDir(args, host.gitignore.Matches, excludes, recurse, func(f string) error {
		for pluginId, p := range cfg.pluginPrepareResults {
			for _, s := range p.Sources {
				if s.Match(f) {
					if sourceFilesByPlugin[pluginId] == nil {
						sourceFilesByPlugin[pluginId] = make([]string, 0, 1)
					}
					sourceFilesByPlugin[pluginId] = append(sourceFilesByPlugin[pluginId], f)
				}
			}
		}

		return nil
	})

	return sourceFilesByPlugin
}

// Let plugins analyze sources and declare their outputs
func (host *GazelleHost) analyzePluginTargetSources(pluginId string, prep pluginConfig, sources []plugin.TargetSource) {
	// TODO: parallelize - analyze concurrently
	for _, src := range sources {

		actx := plugin.NewAnalyzeContext(&src, host.database)

		err := host.plugins[pluginId].Analyze(actx)
		if err != nil {
			// TODO:
			fmt.Println(fmt.Errorf("analyze failed for %s: %w", pluginId, err))
		}
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
