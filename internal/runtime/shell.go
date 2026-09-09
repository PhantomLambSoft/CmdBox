package runtime

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type candidate struct {
	exe  string
	args []string
}

// BuildShellCommand constructs a shell command for the given platform, using the preferred or available shell executables.
func BuildShellCommand(command string, preferredShell string) ([]string, error) {
	var candidates []candidate

	// Windows options
	if runtime.GOOS == "windows" {
		if preferredShell != "" {
			candidates = append(candidates, candidate{
				exe:  preferredShell,
				args: windowsShellArgs(preferredShell, command),
			})
		}

		envShell, exists := os.LookupEnv("CMDBOX_SHELL")
		if exists {
			candidates = append(candidates, candidate{
				exe:  envShell,
				args: windowsShellArgs(envShell, command),
			})
		}

		// Known good fallbacks
		candidates = append(candidates,
			candidate{exe: "pwsh", args: []string{"pwsh", "-NoProfile", "-Command", command}},
			candidate{exe: "powershell", args: []string{"powershell", "-NoProfile", "-Command", command}},
			candidate{exe: "cmd.exe", args: []string{"cmd.exe", "/C", command}},
		)

		for _, c := range candidates {
			if which(c.exe) {
				return c.args, nil
			}
		}

		// Last resort option
		return []string{"cmd.exe", "/C", command}, nil
	}

	if preferredShell != "" {
		candidates = append(candidates, candidate{
			exe:  preferredShell,
			args: []string{preferredShell, "-lc", command},
		})
	}

	if envShell := os.Getenv("SHELL"); envShell != "" {
		candidates = append(candidates, candidate{
			exe:  envShell,
			args: []string{envShell, "-lc", command},
		})
	}

	candidates = append(candidates,
		candidate{exe: "/bin/bash", args: []string{"/bin/bash", "-lc", command}},
		candidate{exe: "/bin/sh", args: []string{"/bin/sh", "-c", command}},
	)

	for _, c := range candidates {
		if filepath.IsAbs(c.exe) {
			if _, err := os.Stat(c.exe); err == nil {
				return c.args, nil
			}
		} else if which(c.exe) {
			return c.args, nil
		}
	}

	return nil, ErrNoUsableShell
}

// windowsShellArgs returns the argument list required to execute a command in the specified Windows shell.
func windowsShellArgs(shell string, command string) []string {
	shell = strings.ToLower(shell)
	if shell == "pwsh" || shell == "powershell" {
		return []string{shell, "-NoProfile", "-Command", command}
	}
	if shell == "cmd" || shell == "cmd.exe" {
		return []string{"cmd.exe", "/C", command}
	}
	return []string{shell, command} // Treat unknown shell as executable
}

// which checks if the given executable can be found, either as an absolute/relative
// path that exists or by searching PATH for its name.
func which(exe string) bool {
	_, err := exec.LookPath(exe)
	return err == nil
}
