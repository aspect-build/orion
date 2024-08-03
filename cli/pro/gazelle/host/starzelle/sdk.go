package starzelle

/**
 * Starlark utility libraries for starzelle plugins.
 *
 * See cli/core/gazelle/common/starlark/stdlib for standard non-starzelle starlark libraries.
 */

import (
	"fmt"
	"reflect"

	common "github.com/aspect-build/silo/cli/core/gazelle/common"
	starUtils "github.com/aspect-build/silo/cli/core/gazelle/common/starlark/utils"
	"github.com/aspect-build/silo/cli/core/gazelle/common/treesitter"
	BazelLog "github.com/aspect-build/silo/cli/core/pkg/logger"
	"github.com/aspect-build/silo/cli/pro/gazelle/host/plugin"
	"go.starlark.net/starlark"
)

func addPlugin(t *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var pluginId starlark.String
	var properties *starlark.Dict
	var prepare, analyze, declare *starlark.Function

	err := starlark.UnpackArgs(
		"register_configure_extension",
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
		"register_rule_kind",
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

func readQueryFilter(v starlark.Value) []string {
	if v == nil {
		return nil
	}

	if filterString, ok := v.(starlark.String); ok {
		return []string{filterString.GoString()}
	}

	return starUtils.ReadStringList(v)
}

func newAstQuery(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var query starlark.String
	var filterValue starlark.Value
	var grammarValue starlark.String

	starlark.UnpackArgs(
		"NewAstQuery",
		args,
		kwargs,
		"query", &query,
		"grammar", &grammarValue,
		"filter??", &filterValue,
	)

	return plugin.QueryDefinition{
		Filter:    readQueryFilter(filterValue),
		Processor: plugin.ASTQueryProcessor,
		Params: plugin.AstQueryParams{
			Grammar: treesitter.LanguageGrammar(grammarValue.GoString()),
			Query:   query.GoString(),
		},
	}, nil
}

func newRegexQuery(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var expression starlark.String
	var filterValue starlark.Value

	starlark.UnpackArgs(
		"NewRegexQuery",
		args,
		kwargs,
		"expression", &expression,
		"filter??", &filterValue,
	)

	re, err := common.ParseRegex(expression.GoString())
	if err != nil {
		return starlark.None, err
	}

	return plugin.QueryDefinition{
		Filter:    readQueryFilter(filterValue),
		Processor: plugin.RegexQueryProcessor,
		Params:    re,
	}, nil
}

func newRawQuery(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var filterValue starlark.Value

	starlark.UnpackArgs(
		"NewRawQuery",
		args,
		kwargs,
		"filter??", &filterValue,
	)

	return plugin.QueryDefinition{
		Filter:    readQueryFilter(filterValue),
		Processor: plugin.RawQueryProcessor,
	}, nil
}

func newJsonQuery(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var queryValue starlark.String
	var filterValue starlark.Value

	starlark.UnpackArgs(
		"NewJsonQuery",
		args,
		kwargs,
		"query?", &queryValue,
		"filter??", &filterValue,
	)

	return plugin.QueryDefinition{
		Filter:    readQueryFilter(filterValue),
		Processor: plugin.JsonQueryProcessor,
		Params:    queryValue.GoString(),
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
				BazelLog.Fatalf("'queries' %v (%s) is not a QueryDefinition", v, reflect.TypeOf(v))
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
	var optional starlark.Bool

	starlark.UnpackArgs(
		"NewImport",
		args,
		kwargs,
		"id", &id,
		"provider", &provider,
		"src", &from,
		"optional", &optional,
	)

	if id.GoString() == "" || provider.GoString() == "" {
		msg := "Import id and provider cannot be empty\n"
		fmt.Printf(msg)
		BazelLog.Fatalf(msg)
	}

	return plugin.TargetImport{
		Symbol: plugin.Symbol{
			Id:       id.GoString(),
			Provider: provider.GoString(),
		},
		Optional: bool(optional.Truth()),
		From:     from.GoString(),
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

func newLabel(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var repo, pkg, name starlark.String

	starlark.UnpackArgs(
		"NewLabel",
		args,
		kwargs,
		"repo", &repo,
		"pkg", &pkg,
		"name", &name,
	)

	return plugin.Label{
		Repo: repo.GoString(),
		Pkg:  pkg.GoString(),
		Name: name.GoString(),
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

var aspectModule = starUtils.CreateModule(
	"aspect",
	map[string]starUtils.ModuleFunction{
		"register_configure_extension": addPlugin,
		"register_rule_kind":           addKind,
		"AstQuery":                     newAstQuery,
		"RegexQuery":                   newRegexQuery,
		"RawQuery":                     newRawQuery,
		"JsonQuery":                    newJsonQuery,
		"PrepareResult":                newPrepareResult,
		"Import":                       newImport,
		"Symbol":                       newSymbol,
		"Label":                        newLabel,
		"Property":                     newProperty,
		"SourceExtensions":             newSourceExtensions,
		"SourceGlobs":                  newSourceGlobs,
		"SourceFiles":                  newSourceFiles,
	},
	map[string]starlark.Value{},
)
