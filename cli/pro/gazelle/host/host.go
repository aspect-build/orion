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
)

const GazelleLanguageName = "AspectConfigure"

// A gazelle
type GazelleHost struct {
	database *plugin.Database

	// Hosted plugins
	// TODO: support enabling/disabling/adding in subdirs
	plugins map[string]plugin.Plugin

	// Ignore configurations for the workspace.
	gitignore *git.GitIgnore

	// Lazy loaded from plugins
	gazelleDirectives []string
	gazelleLoadInfo   []rule.LoadInfo
	gazelleKindInfo   map[string]rule.KindInfo
}

// This is the entrypoint for the gazelle extension initialization.
func NewLanguage() gazelleLanguage.Language {
	host := NewHost()
	host.loadStarzellePlugins()
	return host
}

func NewHost() *GazelleHost {
	l := &GazelleHost{
		gitignore: git.NewGitIgnore(),
		plugins:   make(map[string]plugin.Plugin),
		database:  &plugin.Database{},
	}

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
		BazelLog.Warnf("No builtin plugins found in %q", builtinPluginDir)
	} else {
		BazelLog.Infof("Loading builtin plugins from %q: %v", builtinPluginDir, builtinPlugins)
	}

	for _, p := range builtinPlugins {
		h.addStarzellePlugin(p)
	}

	// TODO: collect plugins configured in BUILD files or aspect cli config
}
func (h *GazelleHost) addStarzellePlugin(defPath string) {
	// Can not add new plugins after configuration/data-collection has started
	if h.gazelleKindInfo != nil || h.gazelleLoadInfo != nil {
		BazelLog.Fatalf("Cannot add plugin %q after configuration has started", defPath)
		return
	}

	proxy, err := starzelle.LoadProxy(defPath)
	if err != nil {
		BazelLog.Errorf("Failed to load plugin definition %q: %v", defPath, err)
		return
	}

	for _, plugin := range proxy.Plugins() {
		BazelLog.Infof("Loaded plugin definition %q\n", plugin.Name())
		h.AddPlugin(plugin)
	}
}

func (h *GazelleHost) AddExtension(defPath string) {
	h.addStarzellePlugin(defPath)
}

func (h *GazelleHost) AddPlugin(plugin plugin.Plugin) {
	h.plugins[plugin.Name()] = plugin
}

func (h *GazelleHost) Kinds() map[string]rule.KindInfo {
	if h.gazelleKindInfo == nil {
		h.gazelleKindInfo = make(map[string]rule.KindInfo, 0)

		for _, plugin := range h.plugins {
			for k, v := range plugin.Rules() {
				if _, exists := h.gazelleKindInfo[k]; exists {
					BazelLog.Warnf("Duplicate rule kind %q from plugin %q", k, plugin.Name())
				}

				h.gazelleKindInfo[k] = v.KindInfo
			}
		}
	}

	return h.gazelleKindInfo
}

func (h *GazelleHost) Loads() []rule.LoadInfo {
	if h.gazelleLoadInfo == nil {
		h.gazelleLoadInfo = make([]rule.LoadInfo, 0)

		loads := make(map[string]*rule.LoadInfo)

		for _, plugin := range h.plugins {
			for name, r := range plugin.Rules() {
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
