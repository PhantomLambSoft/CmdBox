package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveArgs_EmptyArgs(t *testing.T) {
	assert.Equal(t, []string{}, resolveArgs([]string{}))
	assert.Nil(t, resolveArgs(nil))
}

func TestResolveArgs_FlagFirst(t *testing.T) {
	tests := [][]string{
		{"--help"},
		{"-h"},
		{"--version"},
		{"-"},
	}
	for _, args := range tests {
		got := resolveArgs(args)
		assert.Equal(t, args, got)
	}
}

func TestResolveArgs_TopLevelCommandNames(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"cmd", []string{"cmd", "list"}},
		{"run", []string{"run", "myalias"}},
		{"preview", []string{"preview", "myalias"}},
		{"var", []string{"var", "list"}},
		{"tag", []string{"tag", "list"}},
		{"history", []string{"history", "list"}},
		{"settings", []string{"settings", "show"}},
		{"profile", []string{"profile", "list"}},
		{"bangbang", []string{"!!"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveArgs(tc.args)
			assert.Equal(t, tc.args, got)
		})
	}
}

func TestResolveArgs_TopLevelCommandAliases(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"cmds", []string{"cmds", "list"}},
		{"vars", []string{"vars", "list"}},
		{"tags", []string{"tags", "list"}},
		{"hist", []string{"hist", "list"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveArgs(tc.args)
			assert.Equal(t, tc.args, got)
		})
	}
}

func TestResolveArgs_LazyHelpAndCompletionAreRecognized(t *testing.T) {
	// help/completion are only registered on rootCmd once resolveArgs forces them into
	// existence, so this also exercises that InitDefaultHelpCmd/InitDefaultCompletionCmd
	// are called before the top-level lookup.
	assert.Equal(t, []string{"help"}, resolveArgs([]string{"help"}))
	assert.Equal(t, []string{"help", "cmd"}, resolveArgs([]string{"help", "cmd"}))
	assert.Equal(t, []string{"completion", "bash"}, resolveArgs([]string{"completion", "bash"}))
}

func TestResolveArgs_CmdSubcommandNamesFallToCmd(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"add", []string{"add", "myalias", "echo hi"}, []string{"cmd", "add", "myalias", "echo hi"}},
		{"get", []string{"get", "myalias"}, []string{"cmd", "get", "myalias"}},
		{"update", []string{"update", "myalias"}, []string{"cmd", "update", "myalias"}},
		{"delete", []string{"delete", "myalias"}, []string{"cmd", "delete", "myalias"}},
		{"search", []string{"search", "term"}, []string{"cmd", "search", "term"}},
		{"untag", []string{"untag", "myalias", "sometag"}, []string{"cmd", "untag", "myalias", "sometag"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveArgs(tc.args)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestResolveArgs_CmdSubcommandAliasesFallToCmd(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"list", []string{"list"}, []string{"cmd", "list"}},
		{"ls", []string{"ls"}, []string{"cmd", "ls"}},
		{"find", []string{"find", "term"}, []string{"cmd", "find", "term"}},
		{"rm", []string{"rm", "myalias"}, []string{"cmd", "rm", "myalias"}},
		{"del", []string{"del", "myalias"}, []string{"cmd", "del", "myalias"}},
		{"remove", []string{"remove", "myalias"}, []string{"cmd", "remove", "myalias"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resolveArgs(tc.args)
			assert.Equal(t, tc.want, got)
		})
	}
}

// "edit" is an alias on cmdUpdateCmd, but it collides with the "edit" subcommand names on
// several other command groups (var, tag, profile, settings). Since isCmdSubcommand only
// consults cmdCmd's own children, "edit" resolves via the cmd fallback exactly like any
// other cmdCmd alias.
func TestResolveArgs_EditAliasFallsToCmd(t *testing.T) {
	got := resolveArgs([]string{"edit", "myalias"})
	assert.Equal(t, []string{"cmd", "edit", "myalias"}, got)
}

// "tag" is both a top-level command (managing tags) and a cmd subcommand (tagging a
// command). Top-level resolution must win: "cb tag ..." should never be rewritten to
// "cb cmd tag ...".
func TestResolveArgs_AmbiguousNamePrefersTopLevel(t *testing.T) {
	got := resolveArgs([]string{"tag", "list"})
	assert.Equal(t, []string{"tag", "list"}, got)
}

func TestResolveArgs_UnknownCommandFallsBackToRun(t *testing.T) {
	tests := [][]string{
		{"my-alias"},
		{"my-alias", "--flag", "value"},
		{"deploy-prod"},
	}
	for _, args := range tests {
		got := resolveArgs(args)
		want := append([]string{"run"}, args...)
		assert.Equal(t, want, got)
	}
}

func TestResolveArgs_EmptyStringFirstArgFallsBackToRun(t *testing.T) {
	got := resolveArgs([]string{""})
	assert.Equal(t, []string{"run", ""}, got)
}

func TestResolveArgs_PreservesRemainingArgsUntouched(t *testing.T) {
	args := []string{"my-alias", "--cwd", "/tmp", "positional"}
	got := resolveArgs(args)
	assert.Equal(t, append([]string{"run"}, args...), got)
}

func TestIsTopLevelCommand(t *testing.T) {
	assert.True(t, isTopLevelCommand("cmd"))
	assert.True(t, isTopLevelCommand("cmds"))
	assert.True(t, isTopLevelCommand("run"))
	assert.True(t, isTopLevelCommand("!!"))
	assert.False(t, isTopLevelCommand("add"))
	assert.False(t, isTopLevelCommand("nonexistent"))
	assert.False(t, isTopLevelCommand(""))
}

func TestIsCmdSubcommand(t *testing.T) {
	assert.True(t, isCmdSubcommand("add"))
	assert.True(t, isCmdSubcommand("ls"))
	assert.True(t, isCmdSubcommand("edit"))
	assert.False(t, isCmdSubcommand("cmd"))
	assert.False(t, isCmdSubcommand("run"))
	assert.False(t, isCmdSubcommand("nonexistent"))
	assert.False(t, isCmdSubcommand(""))
}
