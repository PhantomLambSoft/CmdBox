package handlers

type AddProfileArgs struct {
	Name           string
	Description    string
	HasDescription bool
	Interactive    bool
}

type UpdateProfileArgs struct {
	CurrentName    string
	NewName        string
	Description    string
	HasDescription bool
	SetPairs       string
	EditMode       bool
	EditFields     string
}

type ListProfileArgs struct {
	ListParams
}

type DeleteProfileArgs struct {
	Name  string
	Force bool
}

type SwitchProfileArgs struct {
	Name     string
	Cmd      bool
	Var      bool
	Settings bool
}

func RunAddProfile(args AddProfileArgs) error {
	return nil
}

func RunGetProfile(name string) error {
	return nil
}

func RunUpdateProfile(args UpdateProfileArgs) error {
	return nil
}

func RunListProfile(args ListProfileArgs) error {
	return nil
}

func RunDeleteProfile(args DeleteProfileArgs) error {
	return nil
}

func RunProfileStatus() error {
	return nil
}

func RunSwitchProfile(args SwitchProfileArgs) error {
	return nil
}
