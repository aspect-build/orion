package plugin

/**
 * Starlark wrappers/interfaces/implementations in order for aspect-configure starzelle
 * plugins to interact with the aspect-configure plugin host.
 */

import (
	"fmt"
	"strings"

	starUtils "github.com/aspect-build/silo/cli/core/gazelle/common/starlark/utils"
	"go.starlark.net/starlark"
	"golang.org/x/exp/maps"
)

// ---------------- PropertyValues
var _ starlark.Value = (*PropertyValues)(nil)
var _ starlark.HasAttrs = (*PropertyValues)(nil)
var _ starlark.Mapping = (*PropertyValues)(nil)

func (p PropertyValues) String() string {
	return fmt.Sprintf("PropertyValues{values: %v}", p.values)
}
func (p PropertyValues) Type() string         { return "PropertyValues" }
func (p PropertyValues) Freeze()              {}
func (p PropertyValues) Truth() starlark.Bool { return starlark.True }
func (p PropertyValues) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", p.Type())
}

func (p PropertyValues) Attr(name string) (starlark.Value, error) {
	if v, ok := p.values[name]; ok {
		return starUtils.Write(v), nil
	}

	return nil, nil

}
func (p PropertyValues) AttrNames() []string {
	return maps.Keys(p.values)
}
func (p PropertyValues) Get(k starlark.Value) (v starlark.Value, found bool, err error) {
	if k.Type() != "string" {
		return nil, false, fmt.Errorf("invalid key type, expected string")
	}
	key := k.(starlark.String).GoString()
	r, found := p.values[key]

	if !found {
		return nil, false, fmt.Errorf("no property named: %s", key)
	}

	return starUtils.Write(r), true, nil
}

// ---------------- PrepareContext

var _ starlark.Value = (*PrepareContext)(nil)
var _ starlark.HasAttrs = (*PrepareContext)(nil)

func (ctx PrepareContext) String() string {
	return fmt.Sprintf("PrepareContext{repo_name: %q, rel: %q, properties: %v}", ctx.RepoName, ctx.Rel, ctx.Properties)
}
func (ctx PrepareContext) Type() string         { return "PrepareContext" }
func (ctx PrepareContext) Freeze()              {}
func (ctx PrepareContext) Truth() starlark.Bool { return starlark.True }
func (ctx PrepareContext) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", ctx.Type())
}

func (ctx PrepareContext) Attr(name string) (starlark.Value, error) {
	switch name {
	case "repo_name":
		return starlark.String(ctx.RepoName), nil
	case "rel":
		return starlark.String(ctx.Rel), nil
	case "properties":
		return ctx.Properties, nil
	}

	return nil, fmt.Errorf("no such attribute: %s", name)
}
func (ctx PrepareContext) AttrNames() []string {
	return []string{"repo_name", "rel", "properties"}
}

// ---------------- DeclareTargetsContext

var _ starlark.Value = (*DeclareTargetsContext)(nil)
var _ starlark.HasAttrs = (*DeclareTargetsContext)(nil)

func (ctx DeclareTargetsContext) Attr(name string) (starlark.Value, error) {
	switch name {
	case "sources":
		// TODO: don't copy every time
		srcs := make([]starlark.Value, 0, len(ctx.Sources))
		for _, v := range ctx.Sources {
			srcs = append(srcs, v)
		}
		return starlark.NewList(srcs), nil
	case "targets":
		return ctx.Targets.(*declareTargetActionsImpl), nil
	}

	return ctx.PrepareContext.Attr(name)
}
func (ctx DeclareTargetsContext) String() string {
	return fmt.Sprintf("DeclareTargetsContext{PrepareContext: %v, sources: %v, targets: %v}", ctx.PrepareContext, ctx.Sources, ctx.Targets)
}
func (ctx DeclareTargetsContext) AttrNames() []string {
	return []string{"repo_name", "rel", "properties", "sources", "targets"}
}
func (ctx DeclareTargetsContext) Type() string { return "DeclareTargetsContext" }

// ---------------- declareTargetActionsImpl

var _ starlark.Value = (*declareTargetActionsImpl)(nil)
var _ starlark.HasAttrs = (*declareTargetActionsImpl)(nil)

func (a *declareTargetActionsImpl) String() string {
	return fmt.Sprintf("declareTargetActionsImpl{%v}", a.targets)
}
func (a *declareTargetActionsImpl) Type() string         { return "declareTargetActionsImpl" }
func (a *declareTargetActionsImpl) Freeze()              {}
func (a *declareTargetActionsImpl) Truth() starlark.Bool { return starlark.True }
func (a *declareTargetActionsImpl) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", a.Type())
}

func (ai *declareTargetActionsImpl) AddCallable(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var starName starlark.String
	var starKind starlark.String
	var starAttrs starlark.Mapping
	var starImports starlark.Value
	var starSymbols starlark.Value
	err := starlark.UnpackArgs(
		fn.Name(),
		args,
		kwargs,
		"name", &starName,
		"kind", &starKind,
		"attrs??", &starAttrs,
		"imports??", &starImports,
		"symbols??", &starSymbols,
	)
	if err != nil {
		return nil, err
	}

	// TODO: don't create new clones of map/arrays every time

	var attrs map[string]interface{}
	if starAttrs != nil {
		attrs = starUtils.ReadMap2(starAttrs, starUtils.Read)
	}

	var imports []TargetImport
	if starImports != nil {
		imports = starUtils.ReadList(starImports, readTargetImport)
	}

	var symbols []TargetSymbol
	if starSymbols != nil {
		symbols = starUtils.ReadList(starSymbols, readTargetSymbols)
	}

	ai.Add(TargetDeclaration{
		Name:    starName.GoString(),
		Kind:    starKind.GoString(),
		Attrs:   attrs,
		Imports: imports,
		Symbols: symbols,
	})

	return starlark.None, nil
}

func (ai *declareTargetActionsImpl) RemoveCallable(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var t starlark.String
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &t); err != nil {
		return nil, err
	}

	ai.Remove(t.GoString())
	return starlark.None, nil
}

func (ai *declareTargetActionsImpl) Attr(name string) (starlark.Value, error) {
	switch name {
	case "add":
		// TODO: don't create every time
		return starlark.NewBuiltin("add", ai.AddCallable), nil
	case "remove":
		// TODO: don't create every time
		return starlark.NewBuiltin("remove", ai.RemoveCallable), nil
	}

	return nil, fmt.Errorf("no such attribute: %s", name)
}
func (*declareTargetActionsImpl) AttrNames() []string {
	return []string{"add", "remove"}
}

// ---------------- TargetSource

var _ starlark.Value = (*TargetSource)(nil)
var _ starlark.HasAttrs = (*TargetSource)(nil)

func (ts TargetSource) String() string {
	return fmt.Sprintf("TargetSource{path: %q, query_results: %v}", ts.Path, ts.QueryResults)
}
func (TargetSource) Freeze() {}
func (TargetSource) Truth() starlark.Bool {
	return starlark.True
}
func (TargetSource) Type() string {
	return "TargetSource"
}
func (ts TargetSource) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", ts.Type())
}

func (ctx TargetSource) Attr(name string) (starlark.Value, error) {
	switch name {
	case "path":
		return starlark.String(ctx.Path), nil
	case "query_results":
		return &ctx.QueryResults, nil
	}

	return nil, fmt.Errorf("no such attribute: %s", name)
}
func (ctx TargetSource) AttrNames() []string {
	return []string{"path", "query_results"}
}

// ---------------- Property

var _ starlark.Value = (*Property)(nil)

func (p Property) String() string {
	return fmt.Sprintf("Property{name: %q, property_type: %q, default_value: %q}", p.Name, p.PropertyType, p.Default)
}
func (p Property) Type() string         { return "Property" }
func (p Property) Freeze()              {}
func (p Property) Truth() starlark.Bool { return starlark.True }
func (p Property) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", p.Type())
}

// ---------------- QueryDefinition

var _ starlark.Value = (*QueryDefinition)(nil)

func (qd QueryDefinition) String() string {
	return fmt.Sprintf("QueryDefinition{grammar: %q, filter: %v, query: %q}", qd.Grammar, qd.Filter, qd.Query)
}
func (qd QueryDefinition) Type() string         { return "QueryDefinition" }
func (qd QueryDefinition) Freeze()              {}
func (qd QueryDefinition) Truth() starlark.Bool { return starlark.True }
func (qd QueryDefinition) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", qd.Type())
}

// ---------------- NamedQueries

var _ starlark.Value = (NamedQueries)(nil)

func (nq NamedQueries) String() string {
	keys := make([]string, 0, len(nq))
	for k := range nq {
		keys = append(keys, k)
	}
	return fmt.Sprintf("NamedQueries(%v)", strings.Join(keys, ","))
}
func (nq NamedQueries) Type() string         { return "NamedQueries" }
func (nq NamedQueries) Freeze()              {}
func (nq NamedQueries) Truth() starlark.Bool { return starlark.True }
func (nq NamedQueries) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", nq.Type())
}

var _ starlark.Mapping = (*QueryResults)(nil)

func (qr *QueryResults) String() string       { return qr.Type() }
func (qr *QueryResults) Type() string         { return "QueryResults" }
func (qr *QueryResults) Freeze()              {}
func (qr *QueryResults) Truth() starlark.Bool { return starlark.True }
func (qr *QueryResults) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", qr.Type())
}

func (qr *QueryResults) Get(k starlark.Value) (v starlark.Value, found bool, err error) {
	if k.Type() != "string" {
		return nil, false, fmt.Errorf("invalid key type, expected string")
	}
	key := k.(starlark.String).GoString()
	r, found := (*qr)[key]

	if !found {
		return nil, false, fmt.Errorf("no query named: %s", key)
	}

	return &r, true, nil
}

// ---------------- PrepareResult

var _ starlark.Value = (*PrepareResult)(nil)

func (r PrepareResult) String() string {
	return fmt.Sprintf("PrepareResult{sources: %v, queries: %v}", r.Sources, r.Queries)
}
func (r PrepareResult) Type() string         { return "PrepareResult" }
func (r PrepareResult) Freeze()              {}
func (r PrepareResult) Truth() starlark.Bool { return starlark.True }
func (r PrepareResult) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", r.Type())
}

// ---------------- SourceExtensionsFilter

var _ starlark.Value = (*SourceExtensionsFilter)(nil)

func (r SourceExtensionsFilter) String() string {
	return fmt.Sprintf("SourceExtensionsFilter{Extensions: %v}", r.Extensions)
}
func (r SourceExtensionsFilter) Type() string         { return "SourceExtensionsFilter" }
func (r SourceExtensionsFilter) Freeze()              {}
func (r SourceExtensionsFilter) Truth() starlark.Bool { return starlark.True }
func (r SourceExtensionsFilter) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", r.Type())
}

// ---------------- SourceFileFilter

var _ starlark.Value = (*SourceGlobFilter)(nil)

func (r SourceGlobFilter) String() string {
	return fmt.Sprintf("SourceGlobFilter{Globs: %v}", r.Globs)
}
func (r SourceGlobFilter) Type() string         { return "SourceGlobFilter" }
func (r SourceGlobFilter) Freeze()              {}
func (r SourceGlobFilter) Truth() starlark.Bool { return starlark.True }
func (r SourceGlobFilter) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", r.Type())
}

// ---------------- SourceFileFilter

var _ starlark.Value = (*SourceFileFilter)(nil)

func (r SourceFileFilter) String() string {
	return fmt.Sprintf("SourceFileFilter{Files: %v}", r.Files)
}
func (r SourceFileFilter) Type() string         { return "SourceFileFilter" }
func (r SourceFileFilter) Freeze()              {}
func (r SourceFileFilter) Truth() starlark.Bool { return starlark.True }
func (r SourceFileFilter) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", r.Type())
}

// ---------------- TargetImport

var _ starlark.Value = (*TargetImport)(nil)
var _ starlark.HasAttrs = (*TargetImport)(nil)

func (ti TargetImport) String() string {
	return fmt.Sprintf("TargetImport{id: %q, provider: %q from: %q}", ti.Id, ti.Provider, ti.From)
}
func (ti TargetImport) Type() string         { return "TargetImport" }
func (ti TargetImport) Freeze()              {}
func (ti TargetImport) Truth() starlark.Bool { return starlark.True }
func (ti TargetImport) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", ti.Type())
}

func (ti TargetImport) Attr(name string) (starlark.Value, error) {
	switch name {
	case "id":
		return starlark.String(ti.Id), nil
	case "provider":
		return starlark.String(ti.Provider), nil
	case "from":
		return starlark.String(ti.From), nil
	}

	return nil, fmt.Errorf("no such attribute: %s", name)
}
func (ti TargetImport) AttrNames() []string {
	return []string{"id", "provider", "from"}
}

// ---------------- TargetSymbol

var _ starlark.Value = (*TargetSymbol)(nil)
var _ starlark.HasAttrs = (*TargetSymbol)(nil)

func (te TargetSymbol) String() string {
	return fmt.Sprintf("TargetSymbol{id: %q, provider: %q, label: %q}", te.Id, te.Provider, te.Label)
}
func (te TargetSymbol) Type() string         { return "TargetSymbol" }
func (te TargetSymbol) Freeze()              {}
func (te TargetSymbol) Truth() starlark.Bool { return starlark.True }
func (te TargetSymbol) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", te.Type())
}

func (te TargetSymbol) Attr(name string) (starlark.Value, error) {
	switch name {
	case "id":
		return starlark.String(te.Id), nil
	case "provider":
		return starlark.String(te.Provider), nil
	case "label":
		return starlark.String(te.Label), nil
	}

	return nil, fmt.Errorf("no such attribute: %s", name)
}
func (te TargetSymbol) AttrNames() []string {
	return []string{"id", "provider", "label"}
}

// ---------------- utils

func readTargetImport(v starlark.Value) TargetImport {
	return v.(TargetImport)
}

func readTargetSymbols(v starlark.Value) TargetSymbol {
	return v.(TargetSymbol)
}
