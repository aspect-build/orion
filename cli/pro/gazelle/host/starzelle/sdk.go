package starzelle

/**
 * Starlark utility libraries for starzelle plugins.
 *
 * See cli/core/gazelle/common/starlark/stdlib for standard non-starzelle starlark libraries.
 */

import (
	starUtils "github.com/aspect-build/silo/cli/core/gazelle/common/starlark/utils"
	BazelLog "github.com/aspect-build/silo/cli/core/pkg/logger"
	"github.com/aspect-build/silo/cli/pro/gazelle/host/plugin"
	"go.starlark.net/starlark"
)

func addPlugin(t *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var pluginId starlark.String
	var properties *starlark.Dict
	var prepare, analyze, declare *starlark.Function

	err := starlark.UnpackArgs(
		"add_plugin",
		args,
		kwargs,
		"id", &pluginId,
		"properties?", &properties,
		"prepare?", &prepare,
		"analyze?", &analyze,
		"declare?", &declare,
	)
	if err != nil {
		return nil, err
	}

	t.Local(proxyStateKey).(*starzelleState).AddPlugin(
		pluginId,
		properties,
		prepare,
		analyze,
		declare,
	)

	return starlark.None, nil
}

func addKind(t *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var kind starlark.String
	var attributes *starlark.Dict

	err := starlark.UnpackArgs(
		"add_kind",
		args,
		kwargs,
		"name", &kind,
		"attributes?", &attributes,
	)
	if err != nil {
		return nil, err
	}

	t.Local(proxyStateKey).(*starzelleState).AddKind(kind, attributes)
	return starlark.None, nil
}

func newQueryDefinition(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

func newSourceExtensions(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return plugin.SourceExtensionsFilter{
		Extensions: starUtils.ReadStringTuple(args),
	}, nil
}

func newSourceGlobs(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return plugin.SourceGlobFilter{
		Globs: starUtils.ReadStringTuple(args),
	}, nil
}

func newSourceFiles(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return plugin.SourceFileFilter{
		Files: starUtils.ReadStringTuple(args),
	}, nil
}

func newPrepareResult(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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
		iter := queriesValue.Iterate()
		defer iter.Done()

		var k starlark.Value
		for iter.Next(&k) {
			v, _, _ := queriesValue.Get(k)

			qd, isQd := v.(plugin.QueryDefinition)
			if !isQd {
				BazelLog.Fatalf("'queries' %v is not a QueryDefinition", qd)
			}

			queries[k.(starlark.String).GoString()] = qd
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
	f, isF := v.(plugin.SourceFilter)

	if !isF {
		BazelLog.Fatalf("'sources' %v is not a SourceFilter", f)
	}

	return f
}

func newImport(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

func newSymbol(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var id, provider starlark.String

	starlark.UnpackArgs(
		"NewSymbol",
		args,
		kwargs,
		"id", &id,
		"provider", &provider,
	)

	return plugin.Symbol{
		Id:       id.GoString(),
		Provider: provider.GoString(),
	}, nil
}

func newProperty(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

var starzelleModule = starUtils.CreateModule(
	"starzelle",
	map[string]starUtils.ModuleFunction{
		"add_plugin":       addPlugin,
		"add_kind":         addKind,
		"Query":            newQueryDefinition,
		"PrepareResult":    newPrepareResult,
		"Import":           newImport,
		"Symbol":           newSymbol,
		"Property":         newProperty,
		"SourceExtensions": newSourceExtensions,
		"SourceGlobs":      newSourceGlobs,
		"SourceFiles":      newSourceFiles,
	},
	map[string]starlark.Value{},
)
