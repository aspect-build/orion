package plugin

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"

	treeutils "github.com/aspect-build/silo/cli/core/gazelle/common/treesitter"
	BazelLog "github.com/aspect-build/silo/cli/core/pkg/logger"
	"github.com/itchyny/gojq"

	"github.com/aspect-build/silo/cli/core/gazelle/common/treesitter"
)

// A set of queries keyed by name.
type NamedQueries map[string]QueryDefinition

// A function to execute a set of queries on the content of a source file.
// A processor may execute queries in parallel or using shared/cached resources for
// the given file and content.
type QueryProcessor = func(fileName string, fileContent []byte, queries NamedQueries, queryResults *QueryResults) error

// A query to run on source files
type QueryDefinition struct {
	Filter    []string
	Processor QueryProcessor
	Params    interface{}
}

func (q QueryDefinition) Match(f string) bool {
	if len(q.Filter) == 0 {
		return true
	}

	for _, filter := range q.Filter {
		is_match, err := filepath.Match(filter, f)

		if err != nil {
			fmt.Println("Error matching filter: ", err)
		}
		if is_match {
			return true
		}
	}
	return false
}

// TODO: better naming?  QueryMapping?
type QueryResults map[string]interface{}

// Multiple matches
type QueryMatches struct {
	m []QueryMatch
}

// The captures of a single query match
type QueryCapture map[string]string

// A single match.
type QueryMatch struct {
	result   interface{}
	captures QueryCapture
}

func NewQueryMatch(captures QueryCapture, result interface{}) QueryMatch {
	return QueryMatch{captures: captures, result: result}
}

func NewQueryMatches(matches []QueryMatch) QueryMatches {
	return QueryMatches{m: matches}
}

// Builtin query processors.
var (
	RawQueryProcessor   QueryProcessor = runRawQueries
	ASTQueryProcessor                  = runPluginTreeQueries
	JsonQueryProcessor                 = runJsonQueries
	RegexQueryProcessor                = runRegexQueries
)

type AstQueryParams struct {
	Grammar treesitter.LanguageGrammar
	Query   string
}

type RegexQueryParams = *regexp.Regexp

type JsonQueryParams = string

func runPluginTreeQueries(fileName string, sourceCode []byte, queries NamedQueries, queryResults *QueryResults) error {
	ast, err := treeutils.ParseSourceCode(toQueryLanguage(fileName, queries), fileName, sourceCode)
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

	// TODO: parallelize - run queries concurrently
	for key, query := range queries {
		resultCh := ast.Query(query.Params.(AstQueryParams).Query)

		// TODO: delay collection from channel until first read?
		// Then it must be cached for later reads...
		match := make([]QueryMatch, 0, 1)
		for r := range resultCh {
			match = append(match, NewQueryMatch(r.Captures(), nil))
		}

		(*queryResults)[key] = NewQueryMatches(match)
	}

	return nil
}

func toQueryLanguage(fileName string, queries NamedQueries) treeutils.LanguageGrammar {
	// TODO: fail if queries on the same file use different languages?

	for _, q := range queries {
		grammar := q.Params.(AstQueryParams).Grammar
		if grammar != "" {
			return treeutils.LanguageGrammar(grammar)
		}
	}

	return treeutils.PathToLanguage(fileName)
}

func runRawQueries(fileName string, sourceCode []byte, queries NamedQueries, queryResults *QueryResults) error {
	for key, _ := range queries {
		(*queryResults)[key] = string(sourceCode)
	}
	return nil
}

func runRegexQueries(fileName string, sourceCode []byte, queries NamedQueries, queryResults *QueryResults) error {
	// TODO: parallelize - run queries concurrently
	for key, q := range queries {
		(*queryResults)[key] = runRegexQuery(string(sourceCode), q.Params.(RegexQueryParams))
	}
	return nil
}

func runRegexQuery(sourceCode string, re *regexp.Regexp) QueryMatches {
	reMatches := re.FindAllStringSubmatch(sourceCode, -1)
	if reMatches == nil {
		return NewQueryMatches(nil)
	}

	matches := make([]QueryMatch, 0, 1)

	for _, reMatch := range reMatches {
		captures := make(QueryCapture)
		for i, name := range re.SubexpNames() {
			if i > 0 && i <= len(reMatch) {
				captures[name] = reMatch[i]
			}
		}

		matches = append(matches, NewQueryMatch(captures, nil))
	}

	return NewQueryMatches(matches)
}

func runJsonQueries(fileName string, sourceCode []byte, queries NamedQueries, queryResults *QueryResults) error {
	var doc interface{}
	err := json.Unmarshal(sourceCode, &doc)
	if err != nil {
		return err
	}

	// TODO: parallelize - run queries concurrently
	for key, q := range queries {
		r, err := runJsonQuery(doc, q.Params.(JsonQueryParams))
		if err != nil {
			return err
		}

		(*queryResults)[key] = r
	}

	return nil
}

func runJsonQuery(doc interface{}, query string) (interface{}, error) {
	q, err := gojq.Parse(query)
	if err != nil {
		return nil, err
	}

	matches := make([]interface{}, 0)

	iter := q.Run(doc)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}

		// See error snippet and notes: https://pkg.go.dev/github.com/itchyny/gojq#readme-usage-as-a-library
		if err, ok := v.(error); ok {
			if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
				break
			}
			return nil, err
		}

		matches = append(matches, v)
	}

	return matches, nil
}
