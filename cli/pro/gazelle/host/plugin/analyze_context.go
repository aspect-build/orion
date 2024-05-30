package plugin

import (
	"go.starlark.net/starlark"
)

var AnalyzeContextAttrs = []string{"source"}

func newAddSymbolFunc(source *TargetSource, database *Database) *starlark.Builtin {
	return starlark.NewBuiltin("add_symbol", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		var id, provider_type, label string
		err := starlark.UnpackArgs(
			"add_symbol", args, kwargs,
			"id", &id,
			"provider_type", &provider_type,
			"label", &label,
		)
		if err != nil {
			return nil, err
		}

		database.AddSymbol(id, provider_type, label, source.Path)
		return starlark.None, nil
	})
}

func NewAnalyzeContext(source *TargetSource, database *Database) AnalyzeContext {
	return AnalyzeContext{
		source:     source,
		add_symbol: newAddSymbolFunc(source, database),
	}
}

type AnalyzeContext struct {
	source     *TargetSource
	add_symbol *starlark.Builtin
}

var _ starlark.Value = (*AnalyzeContext)(nil)
var _ starlark.HasAttrs = (*AnalyzeContext)(nil)

func (a *AnalyzeContext) Attr(name string) (starlark.Value, error) {
	switch name {
	case "source":
		return a.source, nil
	case "add_symbol":
		return a.add_symbol, nil
	default:
		return nil, starlark.NoSuchAttrError(name)
	}
}

func (a *AnalyzeContext) AttrNames() []string {
	return AnalyzeContextAttrs
}

func (a *AnalyzeContext) Freeze() {
	panic("unfreezeable: AnalyzeContext")
}

func (a *AnalyzeContext) Hash() (uint32, error) {
	panic("unhashable: AnalyzeContext")
}

func (a *AnalyzeContext) String() string {
	return a.Type()
}

func (a *AnalyzeContext) Truth() starlark.Bool {
	return starlark.True
}

func (a *AnalyzeContext) Type() string {
	return "AnalyzeContext"
}
