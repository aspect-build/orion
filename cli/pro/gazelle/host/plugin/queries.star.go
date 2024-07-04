package plugin

import (
	"fmt"
	"strings"

	"go.starlark.net/starlark"
)

// ---------------- QueryCapture

var _ starlark.Mapping = (*QueryCapture)(nil)

func (q *QueryCapture) Get(k starlark.Value) (v starlark.Value, found bool, err error) {
	if k.Type() != "string" {
		return nil, false, fmt.Errorf("invalid key type, expected string")
	}
	key := k.(starlark.String).GoString()
	r, found := (*q)[key]

	if !found {
		return nil, false, fmt.Errorf("no capture named: %s", key)
	}
	return starlark.String(r), true, nil
}

func (q *QueryCapture) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", q.Type())
}

func (q *QueryCapture) Freeze()              {}
func (q *QueryCapture) String() string       { return q.Type() }
func (q *QueryCapture) Truth() starlark.Bool { return starlark.True }
func (q *QueryCapture) Type() string         { return "QueryCapture" }

// ---------------- QueryMatch

var _ starlark.HasAttrs = (*QueryMatch)(nil)

func (q *QueryMatch) Attr(name string) (starlark.Value, error) {
	switch name {
	case "captures":
		return &q.captures, nil
	default:
		return nil, starlark.NoSuchAttrError(name)
	}
}
func (q *QueryMatch) AttrNames() []string {
	return []string{"captures"}
}

func (q *QueryMatch) String() string {
	return fmt.Sprintf("QueryMatch(captures: %v)", q.captures)
}
func (q *QueryMatch) Type() string {
	return "QueryMatch"
}
func (q *QueryMatch) Freeze()              {}
func (q *QueryMatch) Truth() starlark.Bool { return starlark.True }
func (q *QueryMatch) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", q.Type())
}

// ---------------- queryMatchIterator

type queryMatchIterator struct {
	m      *[]QueryMatch
	cursor int
}

var _ starlark.Iterator = (*queryMatchIterator)(nil)

func (q *queryMatchIterator) Done() {
	q.cursor = 0
}

func (q *queryMatchIterator) Next(p *starlark.Value) bool {
	if q.cursor+1 > len(*q.m) {
		return false
	}
	match := (*q.m)[q.cursor]
	*p = &match
	q.cursor++
	return true
}

// ---------------- QueryMatches

var _ starlark.Iterable = (*QueryMatches)(nil)
var _ starlark.Indexable = (*QueryMatches)(nil)

func (q *QueryMatches) Index(i int) starlark.Value {
	return &(*q.m)[i]
}

func (q *QueryMatches) Len() int {
	return len(*q.m)
}

func (q *QueryMatches) Freeze() {}

func (q *QueryMatches) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", q.Type())
}

func (q *QueryMatches) Iterate() starlark.Iterator {
	return &queryMatchIterator{m: q.m, cursor: 0}
}

func (q *QueryMatches) String() string {
	return q.Type()
}

func (q *QueryMatches) Truth() starlark.Bool {
	return starlark.True
}

func (q *QueryMatches) Type() string {
	return "QueryMatches"
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

var _ starlark.Value = (*NamedQueries)(nil)

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
