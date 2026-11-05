package shellinit

import (
	"path/filepath"
	"strings"
)

type InstallMode int

const (
	// ModeProfileBlock - inserts into an existing file
	ModeProfileBlock InstallMode = iota

	// ModeWriteFile - writes to a new file (used for fish)
	ModeWriteFile

	// ModeWrapperHint - can't be installed, show instructions
	ModeWrapperHint
)

type Spec struct {
	Name        string
	Display     string
	Integration string
	Mode        InstallMode
	DefaultPath func(Env) (string, error)
}

var shellOrder = []string{"bash", "zsh", "fish", "pwsh", "powershell", "cmd"}

var specs = map[string]Spec{
	"bash": {
		Name: "bash", Display: "Bash", Integration: "bash.sh",
		Mode: ModeProfileBlock, DefaultPath: bashProfile,
	},
	"zsh": {
		Name: "zsh", Display: "Zsh", Integration: "zsh.sh",
		Mode: ModeProfileBlock, DefaultPath: zshProfile,
	},
	"fish": {
		Name: "fish", Display: "Fish", Integration: "fish.fish",
		Mode: ModeWriteFile, DefaultPath: fishFunction,
	},
	"pwsh": {
		Name: "pwsh", Display: "PowerShell (pwsh)", Integration: "powershell.ps1",
		Mode: ModeProfileBlock, DefaultPath: pwshProfile,
	},
	"powershell": {
		Name: "powershell", Display: "PowerShell", Integration: "powershell.ps1",
		Mode: ModeProfileBlock, DefaultPath: windowsPowerShellProfile,
	},
	"cmd": {
		Name: "cmd", Display: "CMD", Integration: "cmd.bat",
		Mode: ModeWrapperHint,
	},
}

func ShellNames() []string {
	return append([]string(nil), shellOrder...)
}

func Lookup(name string) (Spec, error) {
	key := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(name)), ".exe")
	spec, ok := specs[key]
	if !ok {
		return Spec{}, &UnknownShellError{Name: name, Valid: ShellNames()}
	}
	return spec, nil
}

func requireHome(env Env, shell string) (string, error) {
	if env.Home == "" {
		return "", &NoDefaultPathError{Shell: shell}
	}
	return env.Home, nil
}

func configHome(env Env) (string, error) {
	if xdg := env.env("XDG_CONFIG_HOME"); xdg != "" {
		return xdg, nil
	}
	home, err := requireHome(env, "config directory")
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config"), nil
}

func bashProfile(env Env) (string, error) {
	home, err := requireHome(env, "bash")
	if err != nil {
		return "", err
	}
	// macOS terminals read .bash_profile and not .bashrc.
	if env.GOOS == "darwin" {
		return filepath.Join(home, ".bash_profile"), nil
	}
	return filepath.Join(home, ".bashrc"), nil
}

func zshProfile(env Env) (string, error) {
	if dir := env.env("ZDOTDIR"); dir != "" {
		return filepath.Join(dir, ".zshrc"), nil
	}
	home, err := requireHome(env, "zsh")
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".zshrc"), nil
}

// fishFunction returns the cbe.fish path. Fish autoloads a function from a file with the same name, so the file
// must be cbe.fish for `cbe` to exist before any other cb function has been loaded.
func fishFunction(env Env) (string, error) {
	cfg, err := configHome(env)
	if err != nil {
		return "", &NoDefaultPathError{Shell: "fish"}
	}
	return filepath.Join(cfg, "fish", "functions", "cbe.fish"), nil
}

const profileFile = "Microsoft.PowerShell_profile.ps1"

func pwshProfile(env Env) (string, error) {
	if p, ok := askPowerShell(env, "pwsh"); ok {
		return p, nil
	}
	if env.GOOS == "windows" {
		home, err := requireHome(env, "pwsh")
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Documents", "PowerShell", profileFile), nil
	}
	cfg, err := configHome(env)
	if err != nil {
		return "", &NoDefaultPathError{Shell: "pwsh"}
	}
	return filepath.Join(cfg, "powershell", profileFile), nil
}

func windowsPowerShellProfile(env Env) (string, error) {
	if env.GOOS != "windows" {
		return pwshProfile(env)
	}
	if p, ok := askPowerShell(env, "powershell"); ok {
		return p, nil
	}
	home, err := requireHome(env, "powershell")
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Documents", "WindowsPowerShell", profileFile), nil
}

// askPowerShell - asks PowerShell directly for the default profile path.
func askPowerShell(env Env, exe string) (string, bool) {
	if env.Output == nil {
		return "", false
	}
	out, err := env.Output(exe, "-NoLogo", "-NoProfile", "-NonInteractive", "-Command", "$PROFILE.CurrentUserCurrentHost")
	if err != nil {
		return "", false
	}
	lines := strings.Split(strings.ReplaceAll(strings.TrimSpace(out), "\r\n", "\n"), "\n")
	line := strings.TrimSpace(lines[len(lines)-1])
	if line == "" || !filepath.IsAbs(line) {
		return "", false
	}
	return line, true
}
