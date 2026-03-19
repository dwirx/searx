package engine

type Result struct {
	Title   string
	URL     string
	Snippet string
}

type SearchEngine interface {
	Search(query string) ([]Result, error)
	Name() string
}
