package handlers

// ListParams defines parameters for filtering, ordering, and paginating lists in an API or data fetching context.
type ListParams struct {
	Order    string
	Fields   []string
	Limit    int
	HasLimit bool
	Page     *bool // nil = not specified - pull from settings, true/false = explicit
	Profile  string
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

type TagParams struct {
	TagNames []string
	Profile  string
}

// stringOrNil returns a pointer to the given string if it is not empty - otherwise, it returns nil.
func stringOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
