package gazelle

import (
	"fmt"
	"path/filepath"

	plugin "github.com/aspect-build/silo/cli/pro/gazelle/host/plugin"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/emirpasic/gods/maps/linkedhashmap"
)

type BUILDConfig struct {
	// Shared across all
	all      map[string]*BUILDConfig
	repoName string

	// This config
	rel    string
	parent *BUILDConfig

	// All directives of this BUILD
	directiveRawValues map[string][]string

	// General global config
	generationMode GenerationModeType

	// Custom/overridden resolutions
	resolves *linkedhashmap.Map

	// Plugin specific config
	pluginPrepareResults map[string]pluginConfig
}

// GenerationModeType represents one of the generation modes.
type GenerationModeType string

// Generation modes
const (
	// None: do not update or create any BUILD files
	GenerationModeNone GenerationModeType = "none"

	// Update: update and maintain existing BUILD files
	GenerationModeUpdate GenerationModeType = "update"

	// Create: create new and updating existing BUILD files
	GenerationModeCreate GenerationModeType = "create"
)

func NewRootConfig(repoName string) *BUILDConfig {
	r := &BUILDConfig{
		repoName:           repoName,
		rel:                "",
		all:                make(map[string]*BUILDConfig),
		directiveRawValues: make(map[string][]string),

		generationMode: GenerationModeCreate,
		resolves:       linkedhashmap.New(),

		pluginPrepareResults: make(map[string]pluginConfig),
	}
	r.all[r.rel] = r
	return r
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

func (p *BUILDConfig) GetConfig(rel string) *BUILDConfig {
	if p.all[rel] == nil {
		// TODO: find first parent with a config
		parentRel := filepath.Dir(rel)
		if parentRel == "." {
			parentRel = ""
		}

		p.all[rel] = p.all[parentRel].NewChildConfig(rel)
	}

	return p.all[rel]
}

// GenerationMode returns whether coarse-grained targets should be
// generated or not.
func (c *BUILDConfig) GenerationMode() GenerationModeType {
	return c.generationMode
}

// SetGenerationMode sets the generation mode.
func (c *BUILDConfig) SetGenerationMode(mode GenerationModeType) {
	c.generationMode = mode
}

func (c *BUILDConfig) IsPluginEnabled(pluginId string) bool {
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
		if len(query.Filter) == 0 {
			fileQueries[n] = query
			continue
		}

		for _, filter := range query.Filter {
			is_match, err := filepath.Match(filter, f)

			if err != nil {
				fmt.Println("Error matching filter: ", err)
			}
			if is_match {
				fileQueries[n] = query
				break
			}
		}
	}

	return fileQueries
}
