package services

type TransferCommand struct {
	Alias       string            `json:"alias"`
	Template    string            `json:"template"`
	Description *string           `json:"description"`
	Tags        []string          `json:"tags"`
	Cwd         *string           `json:"cwd"`
	Shell       *string           `json:"shell"`
	Env         map[string]string `json:"env"`
	Timeout     *int              `json:"timeout"`
}

type TransferVariable struct {
	Name  string   `json:"name"`
	Value string   `json:"value"`
	Tags  []string `json:"tags"`
}

type TransferDocument struct {
	Version         string             `json:"version"`
	Type            string             `json:"type"`
	ExportedAt      string             `json:"exported_at"`
	CommandProfile  string             `json:"command_profile"`
	VariableProfile string             `json:"variable_profile"`
	Commands        []TransferCommand  `json:"commands"`
	Variables       []TransferVariable `json:"variables"`
}
