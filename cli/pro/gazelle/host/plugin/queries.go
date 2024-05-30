package plugin

// TODO: better naming?  QueryMapping?
type QueryResults map[string]QueryMatches

// a struct representing all matches for a given query.
type QueryMatches struct {
	m *[]QueryMatch
}

// struct prepresenting a single matches' captures.
type QueryCapture map[string]string

// struct prepresenting a single match.
type QueryMatch struct {
	captures QueryCapture
}

func NewQueryMatch(captures QueryCapture) QueryMatch {
	return QueryMatch{captures: captures}
}

func NewQueryMatches(matches *[]QueryMatch) QueryMatches {
	return QueryMatches{m: matches}
}
