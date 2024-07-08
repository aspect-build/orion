package plugin

/**
 * Starlark wrappers/interfaces/implementations in order for aspect-configure starzelle
 * plugins to interact with the aspect-configure plugin host.
 */

import (
	"fmt"

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
func (ai *declareTargetActionsImpl) Attr(name string) (starlark.Value, error) {
	switch name {
	case "add":
		return declareTargetAdd.BindReceiver(ai), nil
	case "remove":
		return declareTargetRemove.BindReceiver(ai), nil
	}

	return nil, fmt.Errorf("no such attribute: %s", name)
}
func (*declareTargetActionsImpl) AttrNames() []string {
	return []string{"add", "remove"}
}

var declareTargetAdd = starlark.NewBuiltin("add", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
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

	var symbols []Symbol
	if starSymbols != nil {
		symbols = starUtils.ReadList(starSymbols, readSymbol)
	}

	ai := fn.Receiver().(*declareTargetActionsImpl)
	ai.Add(TargetDeclaration{
		Name:    starName.GoString(),
		Kind:    starKind.GoString(),
		Attrs:   attrs,
		Imports: imports,
		Symbols: symbols,
	})

	return starlark.None, nil
})
var declareTargetRemove = starlark.NewBuiltin("remove", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var t starlark.String
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &t); err != nil {
		return nil, err
	}

	ai := fn.Receiver().(*declareTargetActionsImpl)
	ai.Remove(t.GoString())
	return starlark.None, nil
})

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

// ---------------- AnalyzeContext

var _ starlark.Value = (*AnalyzeContext)(nil)
var _ starlark.HasAttrs = (*AnalyzeContext)(nil)

func (a *AnalyzeContext) Attr(name string) (starlark.Value, error) {
	switch name {
	case "source":
		return a.Source, nil
	case "add_symbol":
		return analyzeContextAddSymbol.BindReceiver(a), nil
	default:
		return nil, starlark.NoSuchAttrError(name)
	}
}

func (a *AnalyzeContext) AttrNames() []string {
	return []string{"source"}
}
func (a *AnalyzeContext) Freeze() {}
func (a *AnalyzeContext) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", a.Type())
}
func (a *AnalyzeContext) String() string {
	return fmt.Sprintf("AnalyzeContext{source: %v}", a.Source)
}
func (a *AnalyzeContext) Truth() starlark.Bool { return starlark.True }
func (a *AnalyzeContext) Type() string         { return "AnalyzeContext" }

var analyzeContextAddSymbol = starlark.NewBuiltin("add_symbol", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var id, provider_type string
	var label Label
	err := starlark.UnpackArgs(
		"add_symbol", args, kwargs,
		"id", &id,
		"provider_type", &provider_type,
		"label", &label,
	)
	if err != nil {
		return nil, err
	}

	ctx := fn.Receiver().(*AnalyzeContext)
	ctx.AddSymbol(label, Symbol{
		Id:       id,
		Provider: provider_type,
	})

	return starlark.None, nil
})

// ---------------- Gazelle Label

var _ starlark.Value = (*Label)(nil)
var _ starlark.HasAttrs = (*Label)(nil)

func (l Label) Attr(name string) (starlark.Value, error) {
	switch name {
	case "repo":
		return starlark.String(l.Repo), nil
	case "pkg":
		return starlark.String(l.Pkg), nil
	case "name":
		return starlark.String(l.Name), nil
	default:
		return nil, starlark.NoSuchAttrError(name)
	}
}

func (l Label) AttrNames() []string {
	return []string{"repo", "pkg", "name"}
}

func (l Label) String() string {
	return fmt.Sprintf("Label{repo: %q, pkg: %q, name: %q}", l.Repo, l.Pkg, l.Name)
}
func (l Label) Type() string         { return "Label" }
func (l Label) Freeze()              {}
func (l Label) Truth() starlark.Bool { return starlark.True }
func (l Label) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", l.Type())
}
