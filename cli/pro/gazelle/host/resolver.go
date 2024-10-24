package gazelle

import (
	"fmt"
	"os"
	"strings"
	"time"

	common "github.com/aspect-build/silo/cli/core/gazelle/common"
	BazelLog "github.com/aspect-build/silo/cli/core/pkg/logger"
	plugin "github.com/aspect-build/silo/cli/pro/gazelle/host/plugin"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/repo"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

var _ resolve.Resolver = (*GazelleHost)(nil)

type ResolutionType = int

const (
	Resolution_Error ResolutionType = iota
	Resolution_None
	Resolution_NotFound
	Resolution_Label
	Resolution_Native
	Resolution_Override
)

func (*GazelleHost) Name() string {
	return GazelleLanguageName
}

func symbolToImportSpec(symbol plugin.Symbol) resolve.ImportSpec {
	return resolve.ImportSpec{
		Lang: symbol.Provider,
		Imp:  symbol.Id,
	}
}

// Determine what rule (r) outputs which can be imported.
func (re *GazelleHost) Imports(c *config.Config, r *rule.Rule, f *rule.File) []resolve.ImportSpec {
	BazelLog.Debugf("Imports(%s): //%s:%s", GazelleLanguageName, f.Pkg, r.Name())

	targetDeclarationAttr := r.PrivateAttr(targetDeclarationKey)
	if targetDeclarationAttr == nil {
		return nil
	}

	targetDeclaration := targetDeclarationAttr.(plugin.TargetDeclaration)

	res := make([]resolve.ImportSpec, 0, len(targetDeclaration.Symbols))
	for _, s := range targetDeclaration.Symbols {
		res = append(res, symbolToImportSpec(s))
	}

	return res
}

// Extra targets embedded within rules.
func (re *GazelleHost) Embeds(r *rule.Rule, f label.Label) []label.Label {
	return []label.Label{}
}

// Resolve the dependencies of a rule and apply them to the necessary rule attributes.
func (re *GazelleHost) Resolve(c *config.Config, ix *resolve.RuleIndex, rc *repo.RemoteCache, r *rule.Rule, importData interface{}, from label.Label) {
	start := time.Now()
	BazelLog.Infof("Resolve(%s): //%s:%s", GazelleLanguageName, from.Pkg, r.Name())

	pluginIdAttr := r.PrivateAttr(targetPluginKey)
	if pluginIdAttr == nil {
		return
	}

	pluginId := pluginIdAttr.(plugin.PluginId)

	// The import data is the imports per attribute
	attrImports := importData.(map[string][]plugin.TargetImport)
	attrValues := r.PrivateAttr(targetAttrValues).(map[string]interface{})

	for attr, imports := range attrImports {
		importLabels, err := re.resolveImports(c, ix, pluginId, imports, from)
		if err != nil {
			BazelLog.Fatalf("Resolution Error: %v", err)
			os.Exit(1)
		}

		if !importLabels.Empty() {
			attrLabels := importLabels.Labels()
			attrValue := make([]interface{}, 0, len(attrLabels))

			// The resolved labels
			for _, l := range attrLabels {
				attrValue = append(attrValue, l.String())
			}

			// The attribute may have had some explicit values set by the plugin in addition to the imports.
			if attrConstValue, hasAttrValue := attrValues[attr]; hasAttrValue {
				for _, val := range attrConstValue.([]interface{}) {
					attrValue = append(attrValue, val)
				}
			}

			// NOTE: the attribute might have additional values added via # keep which gazelle will maintain
			// despite doing SetAttr.

			r.SetAttr(attr, attrValue)
		}
	}

	BazelLog.Infof("Resolve(%s): //%s:%s DONE in %s", GazelleLanguageName, from.Pkg, r.Name(), time.Since(start).String())
}

func (re *GazelleHost) resolveImports(
	c *config.Config,
	ix *resolve.RuleIndex,
	pluginId plugin.PluginId,
	imports []plugin.TargetImport,
	from label.Label,
) (*common.LabelSet, error) {
	deps := common.NewLabelSet(from)

	for _, imp := range imports {
		resolutionType, dep, err := re.resolveImport(c, ix, pluginId, imp, from)
		if err != nil {
			return nil, err
		}

		if resolutionType == Resolution_NotFound {
			BazelLog.Debugf("import '%s' for target '%s' not found", imp.Id, from.String())

			if !imp.Optional {
				notFound := fmt.Errorf(
					"Import %[1]q from %[2]q is an unknown dependency. Possible solutions:\n"+
						"\t1. Instruct Gazelle to resolve to a known dependency using a directive:\n"+
						"\t\t# aspect:resolve [src-lang] %[3]s import-string label\n",
					imp.Id, imp.From, pluginId,
				)

				fmt.Printf("Resolution error %v\n", notFound)
				continue
			}
		}

		if resolutionType == Resolution_Native || resolutionType == Resolution_None {
			continue
		}

		if dep != nil {
			deps.Add(dep)
		}
	}

	return deps, nil
}

func (host *GazelleHost) resolveImport(
	c *config.Config,
	ix *resolve.RuleIndex,
	pluginId plugin.PluginId,
	impt plugin.TargetImport,
	from label.Label,
) (ResolutionType, *label.Label, error) {
	// Convert to gazelle resolve.ImportSpec api
	importSpec := symbolToImportSpec(impt.Symbol)

	// Gazelle overrides
	// TODO: generalize into gazelle/common
	if override, ok := resolve.FindRuleWithOverride(c, importSpec, GazelleLanguageName); ok {
		return Resolution_Label, &override, nil
	}

	// TODO: Aspect Overrides
	// if res := c.GetResolution(impt.Name); res != nil {
	// 	return Resolution_Override, res, nil
	// }

	// Match imports generated by the starzelle gazelle plugin.
	// TODO: generalize into gazelle/common
	if matches := ix.FindRulesByImportWithConfig(c, importSpec, GazelleLanguageName); len(matches) > 0 {
		filteredMatches := make([]label.Label, 0, len(matches))
		for _, match := range matches {
			// Prevent from adding itself as a dependency.
			if !match.IsSelfImport(from) {
				filteredMatches = append(filteredMatches, match.Label)
			}
		}

		// Too many results, don't know which is correct
		// TODO: resolution conflicts must be solved by plugins
		if len(filteredMatches) > 1 {
			return Resolution_Error, nil, fmt.Errorf(
				"Import %q from %q resolved to multiple targets (%s) - this must be fixed using the \"aspect:resolve\" directive",
				impt.Id, impt.From, targetListFromResults(matches))
		}

		// The matches were self imports, no dependency is needed
		if len(filteredMatches) == 0 {
			return Resolution_None, nil, nil
		}

		match := filteredMatches[0]

		return Resolution_Label, &match, nil
	}

	// TODO: "native" imports
	// if IsNativeImport(impt.Imp) {
	// 	return Resolution_Native, nil, nil
	// }

	// Lookup symbols across plugins in the symbol db
	for _, symbol := range host.database.Symbols {
		// TODO: only match correct "providers"

		if importSpec.Imp == symbol.Symbol.Id {
			l := label.Label{
				Repo:     symbol.Label.Repo,
				Pkg:      symbol.Label.Pkg,
				Name:     symbol.Label.Name,
				Relative: false,
			}
			return Resolution_Label, &l, nil
		}
	}

	return Resolution_NotFound, nil, nil
}

var _ resolve.CrossResolver = (*GazelleHost)(nil)

// Support imports from other gazelle extensions resolving to symbols provided by starzelle plugins.
func (ts *GazelleHost) CrossResolve(c *config.Config, ix *resolve.RuleIndex, imp resolve.ImportSpec, lang string) []resolve.FindResult {
	// Skip resolves from within this gazelle language, only support resolving from other languages
	if lang == GazelleLanguageName {
		return nil
	}

	// Search for results within this gazelle plugin
	return ix.FindRulesByImportWithConfig(c, imp, GazelleLanguageName)
}

// targetListFromResults returns a string with the human-readable list of
// targets contained in the given results.
// TODO: move to gazelle/common
func targetListFromResults(results []resolve.FindResult) string {
	list := make([]string, len(results))
	for i, result := range results {
		list[i] = result.Label.String()
	}
	return strings.Join(list, ", ")
}
