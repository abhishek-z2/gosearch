package search

type Options struct {
	CaseInsensitive bool
	InvertMatch     bool
	Recursive       bool
	CountOnly       bool
	Extensions      []string
}
