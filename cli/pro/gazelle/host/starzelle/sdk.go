package starzelle

/**
 * Starlark utility libraries for starzelle plugins.
 *
 * See cli/core/gazelle/common/starlark/stdlib for standard non-starzelle starlark libraries.
 */

import (
	starUtils "github.com/aspect-build/silo/cli/core/gazelle/common/starlark/utils"
	"github.com/aspect-build/silo/cli/pro/gazelle/host/plugin"
	"go.starlark.net/starlark"
)

func AddLanguagePlugin(t *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var pluginId starlark.String
	var properties *starlark.Dict
	var rules *starlark.Dict
	var prepare, analyze, declare *starlark.Function

	err := starlark.UnpackArgs(
		"AddLanguagePlugin",
		args,
		kwargs,
		"id", &pluginId,
		"properties?", &properties,
		"rules?", &rules,
		"prepare?", &prepare,
		"analyze?", &analyze,
		"declare?", &declare,
	)
	if err != nil {
		return nil, err
	}

	t.Local(proxyStateKey).(*starzelleState).AddLanguagePlugin(
		pluginId,
		properties,
		rules,
		prepare,
		analyze,
		declare,
	)

	return starlark.None, nil
}

func NewQueryDefinition(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var queryType, query starlark.String
	var filterValue starlark.Value
	var grammarValue starlark.String

	starlark.UnpackArgs(
		"NewQueryDefinition",
		args,
		kwargs,
		"type", &queryType,
		"query", &query,
		"grammar", &grammarValue,
		"filter??", &filterValue,
	)

	var filters []string
	if filterValue != nil {
		if filterString, ok := filterValue.(starlark.String); ok {
			filters = []string{filterString.GoString()}
		} else {
			filters = starUtils.ReadStringList(filterValue)
		}
	}

	return plugin.QueryDefinition{
		Grammar: plugin.Grammar(grammarValue.GoString()),
		Filter:  filters,
		Query:   query.GoString(),
	}, nil
}

func NewSourceExtensions(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return plugin.SourceExtensionsFilter{
		Extensions: starUtils.ReadStringTuple(args),
	}, nil
}

func NewSourceGlobs(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return plugin.SourceGlobFilter{
		Globs: starUtils.ReadStringTuple(args),
	}, nil
}

func NewSourceFiles(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return plugin.SourceFileFilter{
		Files: starUtils.ReadStringTuple(args),
	}, nil
}

func NewPrepareResult(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var queriesValue *starlark.Dict
	var sourcesValue *starlark.List

	starlark.UnpackArgs(
		"NewPrepareResult",
		args,
		kwargs,
		"sources", &sourcesValue,
		"queries??", &queriesValue,
	)

	queries := make(plugin.NamedQueries)
	if queriesValue != nil {
		for _, k := range queriesValue.Keys() {
			v, _, _ := queriesValue.Get(k)
			q := v.(plugin.QueryDefinition)
			queries[k.(starlark.String).GoString()] = q
		}
	}

	sources := []plugin.SourceFilter{}
	if sourcesValue != nil {
		sources = starUtils.ReadList(sourcesValue, readSourceFilter)
	}

	return plugin.PrepareResult{
		Sources: sources,
		Queries: queries,
	}, nil
}

func readSourceFilter(v starlark.Value) plugin.SourceFilter {
	return v.(plugin.SourceFilter)
}

func NewImport(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var id, provider, from starlark.String

	starlark.UnpackArgs(
		"NewImport",
		args,
		kwargs,
		"id", &id,
		"provider", &provider,
		"src", &from,
	)

	return plugin.TargetImport{
		Symbol: plugin.Symbol{
			Id:       id.GoString(),
			Provider: provider.GoString(),
		},
		From: from.GoString(),
	}, nil
}

func NewSymbol(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var id, provider, label starlark.String

	starlark.UnpackArgs(
		"NewSymbol",
		args,
		kwargs,
		"id", &id,
		"provider", &provider,
		"label", &label,
	)

	return plugin.TargetSymbol{
		Symbol: plugin.Symbol{
			Id:       id.GoString(),
			Provider: provider.GoString(),
		},
		Label: label.GoString(),
	}, nil
}

func NewProperty(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var propType, propDefault starlark.String

	starlark.UnpackArgs(
		"NewProperty",
		args,
		kwargs,
		"type", &propType,
		"default", &propDefault,
	)

	return plugin.Property{
		PropertyType: propType.GoString(),
		Default:      propDefault.GoString(),
	}, nil
}

var starzelleModule = map[string]starlark.Value{
	"starzelle": starUtils.CreateModule(
		"starzelle",
		map[string]starUtils.ModuleFunction{
			"AddLanguagePlugin": AddLanguagePlugin,
			"Query":             NewQueryDefinition,
			"PrepareResult":     NewPrepareResult,
			"Import":            NewImport,
			"Symbol":            NewSymbol,
			"Property":          NewProperty,
			"SourceExtensions":  NewSourceExtensions,
			"SourceGlobs":       NewSourceGlobs,
			"SourceFiles":       NewSourceFiles,
		},
		map[string]starlark.Value{},
	),
}
