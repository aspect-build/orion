package queries

import (
	treeutils "github.com/aspect-build/silo/cli/core/gazelle/common/treesitter"
	BazelLog "github.com/aspect-build/silo/cli/core/pkg/logger"
	"github.com/aspect-build/silo/cli/pro/gazelle/host/plugin"
)

func runPluginTreeQueries(fileName string, sourceCode []byte, queries plugin.NamedQueries, queryResults chan *plugin.QueryProcessorResult) error {
	ast, err := treeutils.ParseSourceCode(toTreeGrammar(fileName, queries), fileName, sourceCode)
	if err != nil {
		return err
	}

	// Parse errors. Only log them due to many false positives.
	// TODO: what false positives? See js plugin where this is from
	if BazelLog.IsLevelEnabled(BazelLog.TraceLevel) {
		treeErrors := ast.QueryErrors()
		if treeErrors != nil {
			BazelLog.Tracef("TreeSitter query errors: %v", treeErrors)
		}
	}

	// TODO: look into running queries in parallel on the same AST
	for key, query := range queries {
		params := query.Params.(plugin.AstQueryParams)
		treeQuery := treeutils.GetQuery(treeutils.LanguageGrammar(params.Grammar), params.Query)
		resultCh := ast.Query(treeQuery)

		// TODO: delay collection from channel until first read?
		// Then it must be cached for later reads...
		match := make([]plugin.QueryMatch, 0, 1)
		for r := range resultCh {
			match = append(match, plugin.NewQueryMatch(r.Captures(), nil))
		}

		queryResults <- &plugin.QueryProcessorResult{
			Key:    key,
			Result: plugin.NewQueryMatches(match),
		}
	}

	return nil
}

func toTreeGrammar(fileName string, queries plugin.NamedQueries) treeutils.LanguageGrammar {
	// TODO: fail if queries on the same file use different languages?

	for _, q := range queries {
		grammar := q.Params.(plugin.AstQueryParams).Grammar
		if grammar != "" {
			return treeutils.LanguageGrammar(grammar)
		}
	}

	return treeutils.PathToLanguage(fileName)
}
