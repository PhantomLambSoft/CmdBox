package handlers

type RunExportCmdArgs struct {
	Aliases    []string
	Tag        string
	Flatten    bool
	Output     string
	CmdProfile string
	VarProfile string
}

type RunExportVarArgs struct {
	Names      []string
	Tag        string
	Flatten    bool
	Output     string
	VarProfile string
}

type RunExportAllArgs struct {
	Flatten    bool
	Output     string
	CmdProfile string
	VarProfile string
}

func RunExportCmds(args RunExportCmdArgs) error {
	return nil
}

func RunExportVars(args RunExportVarArgs) error {
	return nil
}

func RunExportAll(args RunExportAllArgs) error {
	return nil
}
