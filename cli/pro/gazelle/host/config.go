package gazelle

import (
	"fmt"

	common "github.com/aspect-build/silo/cli/core/gazelle/common"
	plugin "github.com/aspect-build/silo/cli/pro/gazelle/host/plugin"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/emirpasic/gods/maps/linkedhashmap"
)

type BUILDConfig struct {
	// Shared across all
	repoName string

	// This config
	rel    string
	parent *BUILDConfig

	// All directives of this BUILD
	directiveRawValues map[string][]string

	// General global config
	generationMode common.GenerationModeType

	// Custom/overridden resolutions
	resolves *linkedhashmap.Map

	// Plugin specific config
	pluginPrepareResults map[plugin.PluginId]pluginConfig
}

func NewRootConfig(repoName string) *BUILDConfig {
	return &BUILDConfig{
		repoName:           repoName,
		rel:                "",
		directiveRawValues: make(map[string][]string),

		generationMode: common.GenerationModeCreate,
		resolves:       linkedhashmap.New(),

		pluginPrepareResults: make(map[string]pluginConfig),
	}
}

func (c *BUILDConfig) NewChildConfig(rel string) *BUILDConfig {
	// TODO: freeze the parent config now that a child has copied/inherited it.

	cCopy := *c

	// Child specific
	cCopy.rel = rel
	cCopy.parent = c
	cCopy.directiveRawValues = make(map[string][]string)

	// Non-inherited that require cloning
	// TODO: verify these should not be inherited
	cCopy.pluginPrepareResults = make(map[string]pluginConfig)

	// Inherited that must be cloned
	cCopy.resolves = linkedhashmap.New()
	c.resolves.Each(cCopy.resolves.Put)

	return &cCopy
}

func (p *BUILDConfig) appendDirectiveValue(key, value string) {
	values, valueExists := p.directiveRawValues[key]
	if !valueExists {
		p.directiveRawValues[key] = []string{value}
	} else {
		p.directiveRawValues[key] = append(values, value)
	}
}

// GenerationMode returns whether coarse-grained targets should be
// generated or not.
func (c *BUILDConfig) GenerationMode() common.GenerationModeType {
	return c.generationMode
}

// SetGenerationMode sets the generation mode.
func (c *BUILDConfig) SetGenerationMode(mode common.GenerationModeType) {
	c.generationMode = mode
}

func (c *BUILDConfig) IsPluginEnabled(pluginId plugin.PluginId) bool {
	val, exists := c.directiveRawValues[pluginId]
	if exists {
		return val[len(val)-1] == "enabled"
	}

	if c.parent == nil {
		return true
	}

	return c.parent.IsPluginEnabled(pluginId)
}

func (c *BUILDConfig) GetResolution(imprt string) *label.Label {
	config := c
	for config != nil {
		for _, glob := range config.resolves.Keys() {
			m, e := doublestar.Match(glob.(string), imprt)
			if e != nil {
				fmt.Println("Resolve import glob error: ", e)
				return nil
			}

			if m {
				resolveLabel, _ := config.resolves.Get(glob)
				return resolveLabel.(*label.Label)
			}
		}
		config = config.parent
	}

	return nil
}

// An extension of PrepareContext+Result to add internal utils
type pluginConfig struct {
	plugin.PrepareContext
	plugin.PrepareResult
}

func (c *pluginConfig) GetQueriesForFile(f string) plugin.NamedQueries {
	fileQueries := make(plugin.NamedQueries)

	for n, query := range c.PrepareResult.Queries {
		if query.Match(f) {
			fileQueries[n] = query
			continue
		}
	}

	return fileQueries
}
