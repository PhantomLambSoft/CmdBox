package handlers

// ListParams defines parameters for filtering, ordering, and paginating lists in an API or data fetching context.
type ListParams struct {
	Order    string
	Fields   []string
	Limit    int
	HasLimit bool
	Page     *bool // nil = not specified - pull from settings, true/false = explicit
	Profile  string
	Tags     []string
}

type SearchParams struct {
	Term     string
	SearchIn []string // fields to search in
	Fields   []string // fields to output
	Limit    int
	HasLimit bool
	Page     *bool
	Profile  string
}
