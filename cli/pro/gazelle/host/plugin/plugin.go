package plugin

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bazelbuild/bazel-gazelle/rule"
)

// TODO: change the interface into a factory method (at least in starzelle)
type Plugin interface {
	// Static plugin metadata
	Name() string
	Properties() map[string]Property

	// Prepare for generating targets
	Prepare(ctx PrepareContext) PrepareResult
	Analyze(ctx AnalyzeContext) error
	DeclareTargets(ctx DeclareTargetsContext) DeclareTargetsResult
}

type PropertyType = string

const (
	PropertyType_String  PropertyType = "string"
	PropertyType_Strings PropertyType = "[]string"
	PropertyType_Bool    PropertyType = "bool"
	PropertyType_Number  PropertyType = "number"
)

type RuleKind struct {
	rule.KindInfo
	Name string
	From string
}

// Properties an extension can be configured
type Property struct {
	Name         string // TODO: drop because it's always specified in a map[Name]?
	PropertyType PropertyType
	Default      interface{}
}

// The context for an extension to prepare for generating targets.
type PrepareContext struct {
	RepoName   string
	Rel        string
	Properties map[string]interface{}
}

// The result of an extension preparing for generating targets.
//
// Queries are mapped by file extension and will be executed against all
// matching extensions.
//
// Example:
//
//	 PrepareResult {
//			Extensions: [".java"],
//			Queries: {
//				"imports": {
//					"Type": "string|strings|exists",
//					"Extensions": ["*.java"],
//					"Query": "(import_list)",
//				},
//			},
//	 }
type PrepareResult struct {
	Sources []SourceFilter
	Queries NamedQueries
}

type SourceFilter interface {
	Match(p string) bool
}

var _ SourceFilter = (*SourceGlobFilter)(nil)

type SourceGlobFilter struct {
	Globs []string
}

func (f SourceGlobFilter) Match(p string) bool {
	for _, glob := range f.Globs {
		m, err := filepath.Match(glob, p)
		if err != nil {
			fmt.Printf("Error matching glob: %v", err)
			return false
		}
		if m {
			return true
		}
	}
	return false
}

var _ SourceFilter = (*SourceExtensionsFilter)(nil)

type SourceExtensionsFilter struct {
	Extensions []string
}

func (f SourceExtensionsFilter) Match(p string) bool {
	for _, ext := range f.Extensions {
		if strings.HasSuffix(p, ext) {
			return true
		}
	}
	return false
}

var _ SourceFilter = (*SourceFileFilter)(nil)

type SourceFileFilter struct {
	Files []string
}

func (sf SourceFileFilter) Match(p string) bool {
	for _, f := range sf.Files {
		if p == f {
			return true
		}
	}
	return false
}

// The context for an extension to generate targets.
//
// Queries results are mapped by file extension, each containing a map of
// query name to result.
type DeclareTargetsContext struct {
	PrepareContext
	Sources []TargetSource

	Targets DeclareTargetActions
}

type DeclareTargetActions interface {
	Add(target TargetDeclaration)
	Remove(target string)
	Targets() []TargetDeclaration
}

var _ DeclareTargetActions = (*declareTargetActionsImpl)(nil)

type declareTargetActionsImpl struct {
	targets []TargetDeclaration
}

func NewDeclareTargetActions() DeclareTargetActions {
	return &declareTargetActionsImpl{
		targets: make([]TargetDeclaration, 0),
	}
}

func (ctx *declareTargetActionsImpl) Targets() []TargetDeclaration {
	return ctx.targets
}
func (ctx *declareTargetActionsImpl) Add(t TargetDeclaration) {
	ctx.targets = append(ctx.targets, t)
}
func (ctx *declareTargetActionsImpl) Remove(t string) {
	for i, target := range ctx.targets {
		if target.Name == t {
			ctx.targets = append(ctx.targets[:i], ctx.targets[i+1:]...)
		}
	}
}

// The result of declaring targets
type DeclareTargetsResult struct {
}

type TargetSource struct {
	Path         string
	QueryResults QueryResults
}
