package search

type Result struct {
	Results []SearchResult
	Matches int
	Err     error
}

type SearchResult struct {
	File       string
	LineNumber int
	Line       string
	isMatch    bool
}
