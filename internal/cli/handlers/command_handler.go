package handlers

type AddCmdArgs struct {
	Alias          string
	Template       string
	HasTemplate    bool
	Description    string
	HasDescription bool
	Interactive    bool
	Cwd            string
	Shell          string
	Env            []string // raw "key=value,key=value" entries, to be parsed in handler
	Timeout        int
	HasTimeout     bool
	Profile        string
}

type UpdateCmdArgs struct {
	CurrentAlias   string
	NewAlias       string
	Template       string
	HasTemplate    bool
	Description    string
	HasDescription bool
	Cwd            string
	HasCwd         bool
	ClearCwd       bool
	Shell          string
	HasShell       bool
	ClearShell     bool
	Env            []string
	HasEnv         bool
	ClearEnv       bool
	Timeout        int
	HasTimeout     bool
	ClearTimeout   bool
	SetPairs       string
	EditMode       bool
	EditFields     string
	Profile        string
}

type ListCmdArgs struct {
	ListParams
	Tags []string
}

type SearchCmdArgs struct {
	SearchParams
}

type DeleteCmdArgs struct {
	Alias   string
	Profile string
}

type TagCmdArgs struct {
	TagParams
	Alias string
}

func RunAddCmd(args AddCmdArgs) error {
	return nil
}

func RunGetCmd(name, profile string) error {
	return nil
}

func RunUpdateCmd(args UpdateCmdArgs) error {
	return nil
}

func RunListCmd(args ListCmdArgs) error {
	return nil
}

func RunSearchCmd(args SearchCmdArgs) error {
	return nil
}

func RunDeleteCmd(args DeleteCmdArgs) error {
	return nil
}

func RunTagCmd(args TagCmdArgs) error {
	return nil
}

func RunUntagCmd(args TagCmdArgs) error {
	return nil
}
