package plugin

import (
	"encoding/json"
	"log"
	"regexp"

	treeutils "github.com/aspect-build/silo/cli/core/gazelle/common/treesitter"
	BazelLog "github.com/aspect-build/silo/cli/core/pkg/logger"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/itchyny/gojq"
	"golang.org/x/sync/errgroup"

	common "github.com/aspect-build/silo/cli/core/gazelle/common"
	"github.com/aspect-build/silo/cli/core/gazelle/common/treesitter"
)

// A set of queries keyed by name.
type NamedQueries map[string]QueryDefinition

// Intermediate object to hold a query key+result in a single struct.
type QueryProcessorResult struct {
	Result interface{}
	Key    string
}

type QueryType = string

const (
	QueryTypeAst   QueryType = "ast"
	QueryTypeRegex           = "regex"
	QueryTypeJson            = "json"
	QueryTypeRaw             = "raw"
)

// A query to run on source files
type QueryDefinition struct {
	Filter    []string
	QueryType QueryType
	Params    interface{}
}

func (q QueryDefinition) Match(f string) bool {
	if len(q.Filter) == 0 {
		return true
	}

	for _, filter := range q.Filter {
		if doublestar.MatchUnvalidated(filter, f) {
			return true
		}
	}
	return false
}

// TODO: better naming?  QueryMapping?
type QueryResults map[string]interface{}

// Multiple matches
type QueryMatches struct {
	Matches []QueryMatch
}

// The captures of a single query match
type QueryCapture map[string]string

// A single match.
type QueryMatch struct {
	Result   interface{}
	Captures QueryCapture
}

func NewQueryMatch(captures QueryCapture, result interface{}) QueryMatch {
	return QueryMatch{Captures: captures, Result: result}
}

func NewQueryMatches(matches []QueryMatch) QueryMatches {
	return QueryMatches{Matches: matches}
}

type AstQueryParams struct {
	Grammar treesitter.LanguageGrammar
	Query   treesitter.TreeQuery
}

type RegexQueryParams = string

type JsonQueryParams = string

func RunQueries(queryType QueryType, fileName string, sourceCode []byte, queries NamedQueries, queryResults chan *QueryProcessorResult) error {
	switch queryType {
	case QueryTypeAst:
		return runPluginTreeQueries(fileName, sourceCode, queries, queryResults)
	case QueryTypeRegex:
		return runRegexQueries(fileName, sourceCode, queries, queryResults)
	case QueryTypeJson:
		return runJsonQueries(fileName, sourceCode, queries, queryResults)
	case QueryTypeRaw:
		return runRawQueries(fileName, sourceCode, queries, queryResults)
	default:
		log.Panicf("Unknown query type: %v", queryType)
		return nil
	}
}

func runPluginTreeQueries(fileName string, sourceCode []byte, queries NamedQueries, queryResults chan *QueryProcessorResult) error {
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

	// TODO: look into running queries in parallel on the same AST
	for key, query := range queries {
		resultCh := ast.Query(query.Params.(AstQueryParams).Query)

		// TODO: delay collection from channel until first read?
		// Then it must be cached for later reads...
		match := make([]QueryMatch, 0, 1)
		for r := range resultCh {
			match = append(match, NewQueryMatch(r.Captures(), nil))
		}

		queryResults <- &QueryProcessorResult{
			Key:    key,
			Result: NewQueryMatches(match),
		}
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

func runRawQueries(fileName string, sourceCode []byte, queries NamedQueries, queryResults chan *QueryProcessorResult) error {
	sourceCodeStr := string(sourceCode)
	for key, _ := range queries {
		queryResults <- &QueryProcessorResult{
			Key:    key,
			Result: sourceCodeStr,
		}
	}
	return nil
}

func runRegexQueries(fileName string, sourceCode []byte, queries NamedQueries, queryResults chan *QueryProcessorResult) error {
	eg := errgroup.Group{}
	eg.SetLimit(10)

	for key, q := range queries {
		eg.Go(func() error {
			queryResults <- &QueryProcessorResult{
				Key:    key,
				Result: runRegexQuery(sourceCode, common.ParseRegex(q.Params.(RegexQueryParams))),
			}
			return nil
		})
	}

	return eg.Wait()
}

func runRegexQuery(sourceCode []byte, re *regexp.Regexp) QueryMatches {
	reMatches := re.FindAllSubmatch(sourceCode, -1)
	if reMatches == nil {
		return NewQueryMatches(nil)
	}

	matches := make([]QueryMatch, 0, 1)

	for _, reMatch := range reMatches {
		captures := make(QueryCapture)
		for i, name := range re.SubexpNames() {
			if i > 0 && i <= len(reMatch) {
				captures[name] = string(reMatch[i])
			}
		}

		matches = append(matches, NewQueryMatch(captures, reMatch[0]))
	}

	return NewQueryMatches(matches)
}

func runJsonQueries(fileName string, sourceCode []byte, queries NamedQueries, queryResults chan *QueryProcessorResult) error {
	var doc interface{}
	err := json.Unmarshal(sourceCode, &doc)
	if err != nil {
		return err
	}

	eg := errgroup.Group{}
	eg.SetLimit(10)

	// TODO: parallelize, see https://github.com/itchyny/gojq/issues/236
	// for issue + potential workaround (patch).
	for key, q := range queries {
		r, err := runJsonQuery(doc, q.Params.(JsonQueryParams))
		if err != nil {
			return err
		}

		queryResults <- &QueryProcessorResult{
			Key:    key,
			Result: r,
		}
	}

	return nil
}

func runJsonQuery(doc interface{}, query string) (interface{}, error) {
	q, err := common.ParseJsonQuery(query)
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
