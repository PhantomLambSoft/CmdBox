package shellinit

import "strings"

const maxDepth = 5

// Detect identifies the current shell by inspecting the process tree or environment variables.
func Detect(env Env) string {
	if shell := detectFromProcessTree(env); shell != "" {
		return shell
	}
	return detectFromEnv(env)
}

// detectFromProcessTree inspects the process tree to identify the current shell by traversing parent processes.
func detectFromProcessTree(env Env) string {
	if env.Parent == nil {
		return ""
	}
	pid := env.Ppid
	for depth := 0; depth < maxDepth && pid > 1; depth++ {
		ppid, name, err := env.Parent(pid)
		if err != nil {
			return ""
		}
		if shell := classify(name); shell != "" {
			return shell
		}
		pid = ppid
	}
	return ""
}

// detectFromEnv determines the user's shell based on environment variables and the operating system.
func detectFromEnv(env Env) string {
	// variable exported by pwsh
	if env.env("POWERSHELL_DISTRIBUTION_CHANNEL") != "" {
		return "pwsh"
	}
	if sh := classify(env.env("SHELL")); sh != "" && sh != "cmd" && sh != "powershell" {
		return sh
	}
	if env.GOOS == "windows" {
		// git-bash, msys2, MinGW
		if env.env("MSYSTEM") != "" || env.env("MINGW_PREFIX") != "" {
			return "bash"
		}
		// If none of the previous are true for Windows, default to cmd
		return "cmd"
	}
	return "bash"
}

// classify determines the shell type based on the provided process or path name. Returns an empty string if unrecognized.
func classify(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	if i := strings.LastIndexAny(n, `/\`); i >= 0 {
		n = n[i+1:]
	}
	n = strings.TrimPrefix(n, "-")
	n = strings.TrimSuffix(n, ".exe")
	switch n {
	case "bash", "zsh", "fish", "pwsh", "powershell", "cmd":
		return n
	case "git-bash", "msys2":
		return "bash"
	}
	return ""
}
