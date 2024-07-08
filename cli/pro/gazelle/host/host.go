package gazelle

/**
 * A Gazelle language.Language implementation hosting and delegating to one
 * or more aspect-configure language implementations.
 */

import (
	"os"
	"path"
	"path/filepath"

	"github.com/aspect-build/silo/cli/core/gazelle/common/git"
	BazelLog "github.com/aspect-build/silo/cli/core/pkg/logger"
	plugin "github.com/aspect-build/silo/cli/pro/gazelle/host/plugin"
	starzelle "github.com/aspect-build/silo/cli/pro/gazelle/host/starzelle"
	"github.com/bazelbuild/bazel-gazelle/config"
	gazelleLanguage "github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/emirpasic/gods/sets/treeset"
)

const GazelleLanguageName = "AspectConfigure"

// A gazelle
type GazelleHost struct {
	database *plugin.Database

	// Hosted plugins
	// TODO: support enabling/disabling/adding in subdirs
	plugins map[string]plugin.Plugin

	// Metadata about rules being generated. May be pre-configured, potentially loaded from *.star etc
	kinds           map[string]plugin.RuleKind
	sourceRuleKinds *treeset.Set

	// Ignore configurations for the workspace.
	gitignore *git.GitIgnore

	// Lazy loaded from plugins
	gazelleDirectives []string
	gazelleLoadInfo   []rule.LoadInfo
	gazelleKindInfo   map[string]rule.KindInfo
}

var _ gazelleLanguage.Language = (*GazelleHost)(nil)
var _ plugin.PluginHost = (*GazelleHost)(nil)

// This is the entrypoint for the gazelle extension initialization.
func NewLanguage() gazelleLanguage.Language {
	return NewHost()
}

func NewHost() *GazelleHost {
	l := &GazelleHost{
		gitignore:       git.NewGitIgnore(),
		plugins:         make(map[string]plugin.Plugin),
		kinds:           make(map[string]plugin.RuleKind),
		sourceRuleKinds: treeset.NewWithStringComparator(),
		database:        &plugin.Database{},
	}

	// Initialize with builtin kinds. Plugins can add/overwrite these.
	for _, k := range builtinKinds {
		l.kinds[k.Name] = k
	}

	l.loadStarzellePlugins()

	return l
}

func (h *GazelleHost) loadStarzellePlugins() {
	// Add builtin languages
	builtinPluginDir := os.Getenv("STARZELLE_PLUGINS")
	if builtinPluginDir == "" {
		builtinPluginDir = path.Join(os.Getenv("RUNFILES_DIR"), "aspect_silo/cli/pro/gazelle/plugins")
	}

	builtinPlugins, err := filepath.Glob(path.Join(builtinPluginDir, "*.lang.star"))
	if err != nil {
		BazelLog.Fatalf("Failed to load builtin plugins: %v", err)
	}

	if len(builtinPlugins) == 0 {
		BazelLog.Tracef("No configure plugins found in %q", builtinPluginDir)
	} else {
		BazelLog.Infof("Loading configure plugins from %q: %v", builtinPluginDir, builtinPlugins)
	}

	for _, p := range builtinPlugins {
		h.LoadPlugin(p)
	}
}
func (h *GazelleHost) LoadPlugin(defPath string) {
	// Can not add new plugins after configuration/data-collection has started
	if h.gazelleKindInfo != nil || h.gazelleLoadInfo != nil {
		BazelLog.Fatalf("Cannot add plugin %q after configuration has started", defPath)
		return
	}

	err := starzelle.LoadProxy(h, defPath)
	if err != nil {
		BazelLog.Errorf("Failed to load plugin definition %q: %v", defPath, err)
		return
	}
}

func (h *GazelleHost) AddPlugin(plugin plugin.Plugin) {
	if _, exists := h.plugins[plugin.Name()]; exists {
		BazelLog.Errorf("Duplicate plugin %q", plugin.Name())
	}

	BazelLog.Infof("Loaded plugin definition %q", plugin.Name())
	h.plugins[plugin.Name()] = plugin
}

func (h *GazelleHost) AddKind(k plugin.RuleKind) {
	if _, exists := h.kinds[k.Name]; exists {
		BazelLog.Errorf("Duplicate rule kind %q", k.Name)
	}

	BazelLog.Infof("Loaded kind %q", k.Name)
	h.kinds[k.Name] = k

	// Clear cached plugin.RuleKind => gazelle mapping.
	h.gazelleKindInfo = nil
	h.gazelleLoadInfo = nil
}

func (h *GazelleHost) Kinds() map[string]rule.KindInfo {
	if h.gazelleKindInfo == nil {
		h.gazelleKindInfo = make(map[string]rule.KindInfo, len(h.kinds))

		// Configured by plugins, potentially overriding builtin
		for k, v := range h.kinds {
			h.gazelleKindInfo[k] = rule.KindInfo{
				MatchAny:        v.MatchAny,
				MatchAttrs:      v.MatchAttrs,
				NonEmptyAttrs:   toKeyTrueMap(v.NonEmptyAttrs),
				MergeableAttrs:  toKeyTrueMap(v.MergeableAttrs),
				ResolveAttrs:    toKeyTrueMap(v.ResolveAttrs),
				SubstituteAttrs: make(map[string]bool),
			}
			h.sourceRuleKinds.Add(k)
		}
	}

	return h.gazelleKindInfo
}

func toKeyTrueMap(keys []string) map[string]bool {
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return m
}

func (h *GazelleHost) Loads() []rule.LoadInfo {
	if h.gazelleLoadInfo == nil {
		h.gazelleLoadInfo = make([]rule.LoadInfo, 0, len(h.kinds))

		loads := make(map[string]*rule.LoadInfo)

		for name, r := range h.kinds {
			from := r.From

			if loads[from] == nil {
				loads[from] = &rule.LoadInfo{
					Name:    from,
					Symbols: make([]string, 0, 1),
					After:   make([]string, 0),
				}
			}
			loads[from].Symbols = append(loads[from].Symbols, name)
		}

		for _, load := range loads {
			h.gazelleLoadInfo = append(h.gazelleLoadInfo, *load)
		}
	}

	return h.gazelleLoadInfo
}

func (*GazelleHost) Fix(c *config.Config, f *rule.File) {
	// Unsupported
}
