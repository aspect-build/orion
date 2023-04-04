package gazelle

import (
	"testing"

	"github.com/aspect-build/silo/cli/core/gazelle/js/parser/tests"
)

func TestTreesitterParser(t *testing.T) {
	tests.RunParserTests(t, NewParser(), true, "treesitter")
}
