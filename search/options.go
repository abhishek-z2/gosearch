package search

type Options struct {
	CaseInsensitive bool
	InvertMatch     bool
	Recursive       bool
	CountOnly       bool
	AfterContext    int
	BeforeContext   int
	Color           bool
	Extensions      []string
	Regex           bool
}
