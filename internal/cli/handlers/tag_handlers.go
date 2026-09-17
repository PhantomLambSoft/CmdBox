package handlers

type AddTagArgs struct {
	Name           string
	Description    string
	HasDescription bool
	Interactive    bool
}

type UpdateTagArgs struct {
	CurrentName    string
	NewName        string
	Description    string
	HasDescription bool
	SetPairs       string
	EditMode       bool
	EditFields     string
}

type ListTagArgs struct {
	ListParams
}

type SearchTagArgs struct {
	SearchParams
}

type DeleteTagArgs struct {
	Name string
}

func RunAddTag(args AddTagArgs) error {
	return nil
}

func RunGetTag(name string) error {
	return nil
}

func RunUpdateTag(args UpdateTagArgs) error {
	return nil
}

func RunListTag(args ListTagArgs) error {
	return nil
}

func RunSearchTag(args SearchTagArgs) error {
	return nil
}

func RunDeleteTag(args DeleteTagArgs) error {
	return nil
}
