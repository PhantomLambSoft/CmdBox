package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWantsHelp(t *testing.T) {
	assert.True(t, wantsHelp([]string{"--help"}))
	assert.True(t, wantsHelp([]string{"-h"}))
	assert.True(t, wantsHelp([]string{"alias", "--help"}))
	assert.True(t, wantsHelp([]string{"alias", "-h", "--other"}))
	assert.False(t, wantsHelp([]string{}))
	assert.False(t, wantsHelp([]string{"alias"}))
	assert.False(t, wantsHelp([]string{"alias", "--verbose"}))
}

func TestTakeValue(t *testing.T) {
	v, next, err := takeValue([]string{"--cwd", "/tmp"}, 0)
	require.NoError(t, err)
	assert.Equal(t, "/tmp", v)
	assert.Equal(t, 2, next)
}

func TestTakeValue_MissingValueAtEnd(t *testing.T) {
	v, next, err := takeValue([]string{"--cwd"}, 0)
	assert.Error(t, err)
	assert.Equal(t, "--cwd requires a value", err.Error())
	assert.Equal(t, "", v)
	assert.Equal(t, 0, next)
}

func TestParseInvocation_AliasOnly(t *testing.T) {
	alias, ctx, profile, preview, vars, err := parseInvocation([]string{"myalias"})
	require.NoError(t, err)
	assert.Equal(t, "myalias", alias)
	assert.Equal(t, "", profile)
	assert.False(t, preview)
	assert.Empty(t, vars)
	assert.False(t, ctx.HasCwd)
	assert.False(t, ctx.HasShell)
	assert.Nil(t, ctx.Verbose)
}

func TestParseInvocation_MissingAlias(t *testing.T) {
	_, _, _, _, _, err := parseInvocation([]string{})
	assert.EqualError(t, err, "missing required argument: alias")
}

func TestParseInvocation_MissingAliasWithOnlyFlags(t *testing.T) {
	_, _, _, _, _, err := parseInvocation([]string{"--capture"})
	assert.EqualError(t, err, "missing required argument: alias")
}

func TestParseInvocation_PreviewFlag(t *testing.T) {
	alias, _, _, preview, _, err := parseInvocation([]string{"myalias", "--preview"})
	require.NoError(t, err)
	assert.Equal(t, "myalias", alias)
	assert.True(t, preview)

	alias, _, _, preview, _, err = parseInvocation([]string{"myalias", "-p"})
	require.NoError(t, err)
	assert.Equal(t, "myalias", alias)
	assert.True(t, preview)
}

func TestParseInvocation_CwdFlag(t *testing.T) {
	_, ctx, _, _, _, err := parseInvocation([]string{"myalias", "--cwd", "/srv"})
	require.NoError(t, err)
	assert.True(t, ctx.HasCwd)
	assert.Equal(t, "/srv", ctx.Cwd)

	_, ctx, _, _, _, err = parseInvocation([]string{"myalias", "-d", "/srv"})
	require.NoError(t, err)
	assert.True(t, ctx.HasCwd)
	assert.Equal(t, "/srv", ctx.Cwd)
}

func TestParseInvocation_CwdFlagMissingValue(t *testing.T) {
	_, _, _, _, _, err := parseInvocation([]string{"myalias", "--cwd"})
	assert.Error(t, err)
}

func TestParseInvocation_EnvFlagRepeatable(t *testing.T) {
	_, ctx, _, _, _, err := parseInvocation([]string{"myalias", "--env", "A=1", "--env", "B=2"})
	require.NoError(t, err)
	assert.Equal(t, []string{"A=1", "B=2"}, ctx.Env)
}

func TestParseInvocation_CaptureFlag(t *testing.T) {
	_, ctx, _, _, _, err := parseInvocation([]string{"myalias", "--capture"})
	require.NoError(t, err)
	assert.True(t, ctx.Capture)

	_, ctx, _, _, _, err = parseInvocation([]string{"myalias", "-c"})
	require.NoError(t, err)
	assert.True(t, ctx.Capture)
}

func TestParseInvocation_ShellFlag(t *testing.T) {
	_, ctx, _, _, _, err := parseInvocation([]string{"myalias", "--shell", "bash"})
	require.NoError(t, err)
	assert.True(t, ctx.HasShell)
	assert.Equal(t, "bash", ctx.Shell)

	_, ctx, _, _, _, err = parseInvocation([]string{"myalias", "-s", "zsh"})
	require.NoError(t, err)
	assert.Equal(t, "zsh", ctx.Shell)
}

func TestParseInvocation_TimeoutFlag(t *testing.T) {
	_, ctx, _, _, _, err := parseInvocation([]string{"myalias", "--timeout", "30"})
	require.NoError(t, err)
	assert.Equal(t, 30, ctx.Timeout)

	_, ctx, _, _, _, err = parseInvocation([]string{"myalias", "-t", "45"})
	require.NoError(t, err)
	assert.Equal(t, 45, ctx.Timeout)
}

func TestParseInvocation_TimeoutFlagNotANumber(t *testing.T) {
	_, _, _, _, _, err := parseInvocation([]string{"myalias", "--timeout", "notanumber"})
	assert.Error(t, err)
}

func TestParseInvocation_EmitFlag(t *testing.T) {
	_, ctx, _, _, _, err := parseInvocation([]string{"myalias", "--emit"})
	require.NoError(t, err)
	assert.True(t, ctx.Emit)

	_, ctx, _, _, _, err = parseInvocation([]string{"myalias", "-e"})
	require.NoError(t, err)
	assert.True(t, ctx.Emit)
}

func TestParseInvocation_VerboseFlags(t *testing.T) {
	_, ctx, _, _, _, err := parseInvocation([]string{"myalias", "--verbose"})
	require.NoError(t, err)
	if assert.NotNil(t, ctx.Verbose) {
		assert.True(t, *ctx.Verbose)
	}

	_, ctx, _, _, _, err = parseInvocation([]string{"myalias", "--no-verbose"})
	require.NoError(t, err)
	if assert.NotNil(t, ctx.Verbose) {
		assert.False(t, *ctx.Verbose)
	}
}

func TestParseInvocation_ProfileFlag(t *testing.T) {
	_, _, profile, _, _, err := parseInvocation([]string{"myalias", "--profile", "work"})
	require.NoError(t, err)
	assert.Equal(t, "work", profile)
}

func TestParseInvocation_ProfileFlagMissingValue(t *testing.T) {
	_, _, _, _, _, err := parseInvocation([]string{"myalias", "--profile"})
	assert.Error(t, err)
}

func TestParseInvocation_RuntimeVarFlagWithValue(t *testing.T) {
	_, _, _, _, vars, err := parseInvocation([]string{"myalias", "--name", "value"})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"name": "value"}, vars)
}

func TestParseInvocation_MultipleRuntimeVars(t *testing.T) {
	_, _, _, _, vars, err := parseInvocation([]string{"myalias", "--foo", "1", "--bar", "2"})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"foo": "1", "bar": "2"}, vars)
}

// A runtime-var flag immediately followed by another flag (rather than a plain value) has
// no value to consume, so it's dropped rather than swallowing the next flag as its value.
func TestParseInvocation_RuntimeVarFlagFollowedByAnotherFlagIsIgnored(t *testing.T) {
	_, _, _, _, vars, err := parseInvocation([]string{"myalias", "--foo", "--capture"})
	require.NoError(t, err)
	assert.Empty(t, vars)
}

func TestParseInvocation_RuntimeVarFlagAtEndOfArgsIsIgnored(t *testing.T) {
	_, _, _, _, vars, err := parseInvocation([]string{"myalias", "--foo"})
	require.NoError(t, err)
	assert.Empty(t, vars)
}

func TestParseInvocation_ExtraPositionalArgsIgnoredBeyondFirst(t *testing.T) {
	alias, _, _, _, _, err := parseInvocation([]string{"myalias", "second", "third"})
	require.NoError(t, err)
	assert.Equal(t, "myalias", alias)
}

func TestParseInvocation_MixedFlagsAndPositionals(t *testing.T) {
	alias, ctx, profile, preview, vars, err := parseInvocation([]string{
		"myalias", "--cwd", "/tmp", "--preview", "--env", "A=1", "--profile", "work", "--myvar", "42",
	})
	require.NoError(t, err)
	assert.Equal(t, "myalias", alias)
	assert.True(t, preview)
	assert.Equal(t, "work", profile)
	assert.Equal(t, "/tmp", ctx.Cwd)
	assert.Equal(t, []string{"A=1"}, ctx.Env)
	assert.Equal(t, map[string]string{"myvar": "42"}, vars)
}

func TestRunRun_HelpShortCircuitsBeforeParsing(t *testing.T) {
	// Even malformed args (missing alias) should not surface a parse error when --help is present.
	err := runRun(runCmd, []string{"--help"})
	assert.NoError(t, err)
}

func TestRunRun_ParseErrorPropagates(t *testing.T) {
	err := runRun(runCmd, []string{})
	assert.EqualError(t, err, "missing required argument: alias")
}

func TestRunRun_HappyPathDelegatesToHandler(t *testing.T) {
	// handlers.RunRunCommand/RunPreviewCommand are currently stubs that always return nil,
	// so this only proves runRun reaches the handler call without erroring.
	err := runRun(runCmd, []string{"myalias"})
	assert.NoError(t, err)

	err = runRun(runCmd, []string{"myalias", "--preview"})
	assert.NoError(t, err)
}

func TestRunPreview_HelpShortCircuitsBeforeParsing(t *testing.T) {
	err := runPreview(previewCmd, []string{"--help"})
	assert.NoError(t, err)
}

func TestRunPreview_ParseErrorPropagates(t *testing.T) {
	err := runPreview(previewCmd, []string{})
	assert.EqualError(t, err, "missing required argument: alias")
}

func TestRunPreview_HappyPathDelegatesToHandler(t *testing.T) {
	err := runPreview(previewCmd, []string{"myalias"})
	assert.NoError(t, err)
}
