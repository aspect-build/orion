package gazelle

import (
	"testing"

	"github.com/aspect-build/silo/cli/core/gazelle/js/parser/tests"
)

func TestEsbuildParser(t *testing.T) {
	tests.RunParserTests(t, NewParser(), false, "esbuild")
}
