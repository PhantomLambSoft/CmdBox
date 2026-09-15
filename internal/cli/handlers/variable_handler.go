package handlers

type AddVarArgs struct {
	Name        string
	Value       string
	HasValue    bool
	Interactive bool
	Profile     string
}

type UpdateVarArgs struct {
	CurrentName string
	NewName     string
	NewValue    string
	HasValue    bool
	SetPairs    string // raw "key=value,key=value" - parsed in handler function
	EditMode    bool
	EditFields  string
	Profile     string
}

type ListVarArgs struct {
	Tags     []string
	Order    string
	Fields   []string
	Limit    int
	HasLimit bool
	Page     *bool // nil = not specified - pull from settings, true/false = explicit
	Profile  string
}

type SearchVarArgs struct {
	Term     string
	SearchIn []string // fields to search in
	Fields   []string // fields to output
	Limit    int
	HasLimit bool
	Page     *bool
	Profile  string
}

type DeleteVarArgs struct {
	Name    string
	Profile string
}

type TagVarArgs struct {
	Name     string
	TagNames []string
	Profile  string
}

func RunAddVar(args AddVarArgs) error {
	return nil
}

func RunGetVar(name, profile string) error {
	return nil
}

func RunUpdateVar(args UpdateVarArgs) error {
	return nil
}

func RunListVar(args ListVarArgs) error {
	return nil
}

func RunSearchVar(args SearchVarArgs) error {
	return nil
}

func RunDeleteVar(args DeleteVarArgs) error {
	return nil
}

func RunTagVar(args TagVarArgs) error {
	return nil
}

func RunUntagVar(args TagVarArgs) error {
	return nil
}
