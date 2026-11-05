package shellinit

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestShellNames_ReturnsExpectedOrderAndCopy(t *testing.T) {
	names := ShellNames()
	require.Equal(t, []string{"bash", "zsh", "fish", "pwsh", "powershell", "cmd"}, names)

	names[0] = "mutated"
	require.Equal(t, "bash", ShellNames()[0], "ShellNames should return a fresh copy each call")
}

func TestLookup_KnownShells(t *testing.T) {
	for _, name := range ShellNames() {
		spec, err := Lookup(name)
		require.NoError(t, err)
		require.Equal(t, name, spec.Name)
	}
}

func TestLookup_CaseInsensitiveAndTrimsExeAndWhitespace(t *testing.T) {
	spec, err := Lookup("  Bash.EXE  ")
	require.NoError(t, err)
	require.Equal(t, "bash", spec.Name)
}

func TestLookup_UnknownShell(t *testing.T) {
	_, err := Lookup("tcsh")
	require.Error(t, err)
	var unknown *UnknownShellError
	require.ErrorAs(t, err, &unknown)
	require.Equal(t, "tcsh", unknown.Name)
	require.Equal(t, ShellNames(), unknown.Valid)
}

func TestBashProfile_Darwin(t *testing.T) {
	path, err := bashProfile(Env{GOOS: "darwin", Home: `/Users/alice`})
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/Users/alice", ".bash_profile"), path)
}

func TestBashProfile_Linux(t *testing.T) {
	path, err := bashProfile(Env{GOOS: "linux", Home: "/home/alice"})
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/home/alice", ".bashrc"), path)
}

func TestBashProfile_NoHome(t *testing.T) {
	_, err := bashProfile(Env{GOOS: "linux", Home: ""})
	var noPath *NoDefaultPathError
	require.ErrorAs(t, err, &noPath)
	require.Equal(t, "bash", noPath.Shell)
}

func TestZshProfile_UsesZDOTDIR(t *testing.T) {
	env := Env{Home: "/home/alice", GetEnv: envWith(map[string]string{"ZDOTDIR": "/custom/zdotdir"})}
	path, err := zshProfile(env)
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/custom/zdotdir", ".zshrc"), path)
}

func TestZshProfile_FallsBackToHome(t *testing.T) {
	env := Env{Home: "/home/alice", GetEnv: envWith(map[string]string{})}
	path, err := zshProfile(env)
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/home/alice", ".zshrc"), path)
}

func TestZshProfile_NoHomeNoZDOTDIR(t *testing.T) {
	env := Env{Home: "", GetEnv: envWith(map[string]string{})}
	_, err := zshProfile(env)
	var noPath *NoDefaultPathError
	require.ErrorAs(t, err, &noPath)
}

func TestConfigHome_UsesXDGConfigHome(t *testing.T) {
	env := Env{Home: "/home/alice", GetEnv: envWith(map[string]string{"XDG_CONFIG_HOME": "/custom/config"})}
	path, err := configHome(env)
	require.NoError(t, err)
	require.Equal(t, "/custom/config", path)
}

func TestConfigHome_FallsBackToHomeConfig(t *testing.T) {
	env := Env{Home: "/home/alice", GetEnv: envWith(map[string]string{})}
	path, err := configHome(env)
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/home/alice", ".config"), path)
}

func TestConfigHome_NoHome(t *testing.T) {
	env := Env{Home: "", GetEnv: envWith(map[string]string{})}
	_, err := configHome(env)
	require.Error(t, err)
}

func TestFishFunction_UsesConfigHome(t *testing.T) {
	env := Env{Home: "/home/alice", GetEnv: envWith(map[string]string{"XDG_CONFIG_HOME": "/custom/config"})}
	path, err := fishFunction(env)
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/custom/config", "fish", "functions", "cbe.fish"), path)
}

func TestFishFunction_NoHomeReturnsFishSpecificError(t *testing.T) {
	env := Env{Home: "", GetEnv: envWith(map[string]string{})}
	_, err := fishFunction(env)
	var noPath *NoDefaultPathError
	require.ErrorAs(t, err, &noPath)
	require.Equal(t, "fish", noPath.Shell)
}

func TestAskPowerShell_NilOutputReturnsFalse(t *testing.T) {
	_, ok := askPowerShell(Env{Output: nil}, "pwsh")
	require.False(t, ok)
}

func TestAskPowerShell_OutputErrorReturnsFalse(t *testing.T) {
	env := Env{Output: func(name string, args ...string) (string, error) {
		return "", errors.New("not found")
	}}
	_, ok := askPowerShell(env, "pwsh")
	require.False(t, ok)
}

func TestAskPowerShell_EmptyOutputReturnsFalse(t *testing.T) {
	env := Env{Output: func(name string, args ...string) (string, error) {
		return "", nil
	}}
	_, ok := askPowerShell(env, "pwsh")
	require.False(t, ok)
}

func TestAskPowerShell_RelativePathReturnsFalse(t *testing.T) {
	env := Env{Output: func(name string, args ...string) (string, error) {
		return "relative/path/profile.ps1", nil
	}}
	_, ok := askPowerShell(env, "pwsh")
	require.False(t, ok)
}

func TestAskPowerShell_AbsolutePathReturnsTrue(t *testing.T) {
	abs := mustAbs(t)
	env := Env{Output: func(name string, args ...string) (string, error) {
		require.Equal(t, "pwsh", name)
		require.Contains(t, args, "$PROFILE.CurrentUserCurrentHost")
		return abs, nil
	}}
	path, ok := askPowerShell(env, "pwsh")
	require.True(t, ok)
	require.Equal(t, abs, path)
}

func TestAskPowerShell_TrimsCRLFAndTakesLastLine(t *testing.T) {
	abs := mustAbs(t)
	env := Env{Output: func(name string, args ...string) (string, error) {
		return "some warning\r\n" + abs + "\r\n", nil
	}}
	path, ok := askPowerShell(env, "pwsh")
	require.True(t, ok)
	require.Equal(t, abs, path)
}

// mustAbs returns an absolute path appropriate for the host OS running the test.
func mustAbs(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("some", "profile.ps1"))
	require.NoError(t, err)
	require.True(t, filepath.IsAbs(abs))
	return abs
}

func TestPwshProfile_UsesAskPowerShellWhenAvailable(t *testing.T) {
	abs := mustAbs(t)
	env := Env{
		GOOS: "linux",
		Output: func(name string, args ...string) (string, error) {
			return abs, nil
		},
	}
	path, err := pwshProfile(env)
	require.NoError(t, err)
	require.Equal(t, abs, path)
}

func TestPwshProfile_WindowsFallback(t *testing.T) {
	env := Env{
		GOOS:   "windows",
		Home:   `C:\Users\alice`,
		Output: func(name string, args ...string) (string, error) { return "", errors.New("no pwsh") },
	}
	path, err := pwshProfile(env)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(`C:\Users\alice`, "Documents", "PowerShell", profileFile), path)
}

func TestPwshProfile_WindowsFallbackNoHome(t *testing.T) {
	env := Env{
		GOOS:   "windows",
		Home:   "",
		Output: nil,
	}
	_, err := pwshProfile(env)
	var noPath *NoDefaultPathError
	require.ErrorAs(t, err, &noPath)
	require.Equal(t, "pwsh", noPath.Shell)
}

func TestPwshProfile_NonWindowsFallbackUsesConfigHome(t *testing.T) {
	env := Env{
		GOOS:   "linux",
		Home:   "/home/alice",
		Output: nil,
		GetEnv: envWith(map[string]string{}),
	}
	path, err := pwshProfile(env)
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/home/alice", ".config", "powershell", profileFile), path)
}

func TestPwshProfile_NonWindowsFallbackNoHome(t *testing.T) {
	env := Env{
		GOOS:   "linux",
		Home:   "",
		Output: nil,
		GetEnv: envWith(map[string]string{}),
	}
	_, err := pwshProfile(env)
	var noPath *NoDefaultPathError
	require.ErrorAs(t, err, &noPath)
	require.Equal(t, "pwsh", noPath.Shell)
}

func TestWindowsPowerShellProfile_DelegatesOnNonWindows(t *testing.T) {
	env := Env{
		GOOS:   "linux",
		Home:   "/home/alice",
		Output: nil,
		GetEnv: envWith(map[string]string{}),
	}
	path, err := windowsPowerShellProfile(env)
	require.NoError(t, err)
	expected, err := pwshProfile(env)
	require.NoError(t, err)
	require.Equal(t, expected, path)
}

func TestWindowsPowerShellProfile_UsesAskPowerShellWhenAvailable(t *testing.T) {
	abs := mustAbs(t)
	env := Env{
		GOOS: "windows",
		Output: func(name string, args ...string) (string, error) {
			require.Equal(t, "powershell", name)
			return abs, nil
		},
	}
	path, err := windowsPowerShellProfile(env)
	require.NoError(t, err)
	require.Equal(t, abs, path)
}

func TestWindowsPowerShellProfile_FallsBackToHome(t *testing.T) {
	env := Env{
		GOOS:   "windows",
		Home:   `C:\Users\alice`,
		Output: func(name string, args ...string) (string, error) { return "", errors.New("no powershell") },
	}
	path, err := windowsPowerShellProfile(env)
	require.NoError(t, err)
	require.Equal(t, filepath.Join(`C:\Users\alice`, "Documents", "WindowsPowerShell", profileFile), path)
}

func TestWindowsPowerShellProfile_NoHome(t *testing.T) {
	env := Env{
		GOOS:   "windows",
		Home:   "",
		Output: nil,
	}
	_, err := windowsPowerShellProfile(env)
	var noPath *NoDefaultPathError
	require.ErrorAs(t, err, &noPath)
	require.Equal(t, "powershell", noPath.Shell)
}
