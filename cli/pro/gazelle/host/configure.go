package gazelle

import (
	"flag"
	"strconv"
	"strings"

	BazelLog "github.com/aspect-build/silo/cli/core/pkg/logger"
	"github.com/aspect-build/silo/cli/pro/gazelle/host/plugin"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
)

var _ config.Configurer = (*GazelleHost)(nil)

func (c *GazelleHost) KnownDirectives() []string {
	if c.gazelleDirectives == nil {
		c.gazelleDirectives = []string{
			// Core builtin directives
			Directive_GenerationMode,
		}

		// TODO: verify no collisions with other plugins/globals

		for _, plugin := range c.plugins {
			// A directive to enable/disable the plugin
			c.gazelleDirectives = append(c.gazelleDirectives, plugin.Name())

			// Directives defined by the plugin
			for _, dir := range plugin.Properties() {
				c.gazelleDirectives = append(c.gazelleDirectives, dir.Name)
			}
		}
	}

	return c.gazelleDirectives
}

func (configurer *GazelleHost) Configure(c *config.Config, rel string, f *rule.File) {
	BazelLog.Tracef("Configure %s", rel)

	// Collect the ignore files for this package
	configurer.gitignore.CollectIgnoreFiles(c, rel)

	// Generate hierarchical configuration.
	if _, exists := c.Exts[GazelleLanguageName]; !exists {
		c.Exts[GazelleLanguageName] = NewRootConfig(c.RepoName)
	}

	config := c.Exts[GazelleLanguageName].(*BUILDConfig).GetConfig(rel)

	// Record directives from the existing BUILD file.
	if f != nil {
		for _, d := range f.Directives {
			config.appendDirectiveValue(d.Key, d.Value)

			// Generic non-plugin specific directives
			switch d.Key {
			case Directive_GenerationMode:
				mode := GenerationModeType(strings.TrimSpace(d.Value))
				switch mode {
				case GenerationModeCreate:
					config.SetGenerationMode(mode)
				case GenerationModeUpdate:
					config.SetGenerationMode(mode)
				case GenerationModeNone:
					config.SetGenerationMode(mode)
				default:
					BazelLog.Fatalf("invalid value for directive %q: %s", Directive_GenerationMode, d.Value)
				}
			}
		}
	}

	// All generation may disabled.
	if config.GenerationMode() == GenerationModeNone {
		BazelLog.Tracef("Configure disabled: %q", rel)
		return
	}

	// Generating new BUILDs may disabled.
	if config.GenerationMode() == GenerationModeUpdate && f == nil {
		BazelLog.Tracef("Configure BUILD creation disabled: %q", rel)
		return
	}

	// Prepare the plugins for this configuration.
	// TODO: parallelize
	for k, p := range configurer.plugins {
		if !config.IsPluginEnabled(k) {
			continue
		}

		prepContext := configToPrepareContext(p, config)
		prepResult := p.Prepare(prepContext)

		// Index the plugins and their PrepareResult
		config.pluginPrepareResults[k] = pluginConfig{
			PrepareContext: prepContext,
			PrepareResult:  prepResult,
		}
	}
}

func configToPrepareContext(p plugin.Plugin, cfg *BUILDConfig) plugin.PrepareContext {
	ctx := plugin.PrepareContext{
		RepoName:   cfg.repoName,
		Rel:        cfg.rel,
		Properties: plugin.NewPropertyValues(),
	}

	for k, p := range p.Properties() {
		pValue := p.Default

		if v, found := cfg.directiveRawValues[p.Name]; found {
			parsedValue, parseErr := parsePropertyValue(p, v)
			if parseErr != nil {
				BazelLog.Warnf("Failed to parse property %q: %v", p.Name, parseErr)
			} else {
				pValue = parsedValue
			}
		}

		ctx.Properties.Add(k, pValue)
	}

	return ctx
}

func parsePropertyValue(p plugin.Property, values []string) (interface{}, error) {
	switch p.PropertyType {
	case plugin.PropertyType_String:
		return onlyValue(p, values), nil
	case plugin.PropertyType_Strings:
		return values, nil
	case plugin.PropertyType_Bool:
		return onlyValue(p, values) == "true", nil
	case plugin.PropertyType_Number:
		return strconv.ParseInt(onlyValue(p, values), 10, 0)
	}

	panic("unhandled property type: " + p.PropertyType)
}

func onlyValue(p plugin.Property, value []string) string {
	c := len(value)

	if c == 0 {
		BazelLog.Fatalf("expected exactly one value, got none")
		return ""
	} else if c > 1 {
		BazelLog.Warnf("expected exactly one value for %q, got %d", p.Name, c)
	}

	return value[c-1]
}

func (c *GazelleHost) RegisterFlags(fs *flag.FlagSet, cmd string, cfg *config.Config) {
}

func (c *GazelleHost) CheckFlags(fs *flag.FlagSet, cfg *config.Config) error {
	return nil
}
