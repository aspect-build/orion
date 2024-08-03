package plugin

import (
	godsutils "github.com/emirpasic/gods/utils"
)

type Symbol struct {
	Id       string // The unique id of the symbol
	Provider string // The provider type of the symbol
}

type TargetImport struct {
	Symbol

	// Optional imports will not be treated as resolution errors when not found.
	Optional bool

	// Where the import is from such as file path, for debugging
	From string
}

type TargetSymbol struct {
	Symbol

	// The label producing the symbol
	Label Label
}

/**
 * A bazel target declaration describing the target name/type/attributes as
 * well as symbols representing imports and exports of the target.
 */
type TargetDeclaration struct {
	Name  string
	Kind  string
	Attrs map[string]interface{}

	// Whether to remove the target from the BUILD
	// TODO: do this somewhere else, it should not be in a flag within a "declaration".
	Remove bool

	// Names (possibly as paths) within this target that are imported/exported
	Imports []TargetImport
	Symbols []Symbol
}

func symbolComparator(a, b interface{}) int {
	nc := godsutils.StringComparator(a.(Symbol).Id, b.(Symbol).Id)
	if nc != 0 {
		return nc
	}

	return godsutils.StringComparator(a.(Symbol).Provider, b.(Symbol).Provider)
}

func TargetImportComparator(a, b interface{}) int {
	nc := symbolComparator(a, b)
	if nc != 0 {
		return nc
	}

	return godsutils.StringComparator(a.(TargetImport).From, b.(TargetImport).From)
}

func TargetExportComparator(a, b interface{}) int {
	nc := symbolComparator(a, b)
	if nc != 0 {
		return nc
	}

	return godsutils.StringComparator(a.(TargetSymbol).Label, b.(TargetSymbol).Label)
}
