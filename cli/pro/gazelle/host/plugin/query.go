package plugin

import (
	"github.com/aspect-build/silo/cli/core/gazelle/common/treesitter"
)

type Grammar treesitter.LanguageGrammar

type NamedQueries map[string]QueryDefinition

type QueryDefinition struct {
	Grammar Grammar
	Filter  []string
	Query   string
}
