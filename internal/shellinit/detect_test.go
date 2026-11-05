package shellinit

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func envWith(getEnv map[string]string) func(string) string {
	return func(key string) string {
		return getEnv[key]
	}
}

func TestClassify(t *testing.T) {
	cases := map[string]string{
		"bash":            "bash",
		"/usr/bin/bash":   "bash",
		`C:\Windows\bash`: "bash",
		"-zsh":            "zsh",
		"zsh":             "zsh",
		"fish":            "fish",
		"pwsh.exe":        "pwsh",
		"PowerShell.exe":  "powershell",
		"cmd.exe":         "cmd",
		"CMD":             "cmd",
		"git-bash.exe":    "bash",
		"msys2":           "bash",
		"explorer.exe":    "",
		"":                "",
		"  bash  ":        "bash",
	}
	for input, want := range cases {
		require.Equal(t, want, classify(input), "classify(%q)", input)
	}
}

func TestDetectFromProcessTree_NilParentReturnsEmpty(t *testing.T) {
	env := Env{Ppid: 100}
	require.Equal(t, "", detectFromProcessTree(env))
}

func TestDetectFromProcessTree_ImmediateMatch(t *testing.T) {
	env := Env{
		Ppid: 100,
		Parent: func(pid int) (int, string, error) {
			require.Equal(t, 100, pid)
			return 1, "bash", nil
		},
	}
	require.Equal(t, "bash", detectFromProcessTree(env))
}

func TestDetectFromProcessTree_ClimbsUntilMatch(t *testing.T) {
	// 300 -> 200 (unrecognized "explorer.exe") -> 100 (recognized "zsh")
	calls := map[int]struct {
		ppid int
		name string
	}{
		300: {200, "explorer.exe"},
		200: {100, "zsh"},
	}
	env := Env{
		Ppid: 300,
		Parent: func(pid int) (int, string, error) {
			c, ok := calls[pid]
			require.True(t, ok, "unexpected pid %d", pid)
			return c.ppid, c.name, nil
		},
	}
	require.Equal(t, "zsh", detectFromProcessTree(env))
}

func TestDetectFromProcessTree_StopsAtPidOne(t *testing.T) {
	env := Env{
		Ppid: 2,
		Parent: func(pid int) (int, string, error) {
			return 1, "explorer.exe", nil
		},
	}
	require.Equal(t, "", detectFromProcessTree(env))
}

func TestDetectFromProcessTree_StopsAtMaxDepth(t *testing.T) {
	calls := 0
	env := Env{
		Ppid: 1000,
		Parent: func(pid int) (int, string, error) {
			calls++
			return pid - 1, "explorer.exe", nil
		},
	}
	require.Equal(t, "", detectFromProcessTree(env))
	require.Equal(t, maxDepth, calls)
}

func TestDetectFromProcessTree_ErrorStopsTraversal(t *testing.T) {
	calls := 0
	env := Env{
		Ppid: 100,
		Parent: func(pid int) (int, string, error) {
			calls++
			return 0, "", errors.New("no such process")
		},
	}
	require.Equal(t, "", detectFromProcessTree(env))
	require.Equal(t, 1, calls)
}

func TestDetectFromEnv_PowerShellDistributionChannelWins(t *testing.T) {
	env := Env{
		GOOS: "linux",
		GetEnv: envWith(map[string]string{
			"POWERSHELL_DISTRIBUTION_CHANNEL": "MSIX",
			"SHELL":                           "/bin/bash",
		}),
	}
	require.Equal(t, "pwsh", detectFromEnv(env))
}

func TestDetectFromEnv_UsesShellVar(t *testing.T) {
	env := Env{
		GOOS:   "linux",
		GetEnv: envWith(map[string]string{"SHELL": "/usr/bin/zsh"}),
	}
	require.Equal(t, "zsh", detectFromEnv(env))
}

func TestDetectFromEnv_IgnoresCmdAndPowershellFromShellVar(t *testing.T) {
	env := Env{
		GOOS:   "linux",
		GetEnv: envWith(map[string]string{"SHELL": "cmd"}),
	}
	require.Equal(t, "bash", detectFromEnv(env))

	env2 := Env{
		GOOS:   "linux",
		GetEnv: envWith(map[string]string{"SHELL": "powershell"}),
	}
	require.Equal(t, "bash", detectFromEnv(env2))
}

func TestDetectFromEnv_WindowsGitBash(t *testing.T) {
	env := Env{
		GOOS:   "windows",
		GetEnv: envWith(map[string]string{"MSYSTEM": "MINGW64"}),
	}
	require.Equal(t, "bash", detectFromEnv(env))
}

func TestDetectFromEnv_WindowsMinGWPrefix(t *testing.T) {
	env := Env{
		GOOS:   "windows",
		GetEnv: envWith(map[string]string{"MINGW_PREFIX": `C:\msys64\mingw64\`}),
	}
	require.Equal(t, "bash", detectFromEnv(env))
}

func TestDetectFromEnv_WindowsDefaultsToCmd(t *testing.T) {
	env := Env{
		GOOS:   "windows",
		GetEnv: envWith(map[string]string{}),
	}
	require.Equal(t, "cmd", detectFromEnv(env))
}

func TestDetectFromEnv_NonWindowsDefaultsToBash(t *testing.T) {
	env := Env{
		GOOS:   "linux",
		GetEnv: envWith(map[string]string{}),
	}
	require.Equal(t, "bash", detectFromEnv(env))

	env2 := Env{
		GOOS:   "darwin",
		GetEnv: envWith(map[string]string{}),
	}
	require.Equal(t, "bash", detectFromEnv(env2))
}

func TestDetect_PrefersProcessTreeOverEnv(t *testing.T) {
	env := Env{
		GOOS: "linux",
		Ppid: 100,
		Parent: func(pid int) (int, string, error) {
			return 1, "fish", nil
		},
		GetEnv: envWith(map[string]string{"SHELL": "/bin/bash"}),
	}
	require.Equal(t, "fish", Detect(env))
}

func TestDetect_FallsBackToEnvWhenProcessTreeUnresolved(t *testing.T) {
	env := Env{
		GOOS: "linux",
		Ppid: 100,
		Parent: func(pid int) (int, string, error) {
			return 1, "explorer.exe", nil
		},
		GetEnv: envWith(map[string]string{"SHELL": "/bin/zsh"}),
	}
	require.Equal(t, "zsh", Detect(env))
}
