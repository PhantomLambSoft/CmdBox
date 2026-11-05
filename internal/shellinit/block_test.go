package shellinit

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpsertBlock_EmptyExisting(t *testing.T) {
	got, err := UpsertBlock("", "export FOO=bar")
	require.NoError(t, err)
	require.Equal(t, StartMark+"\nexport FOO=bar\n"+EndMark+"\n", got)
}

func TestUpsertBlock_AppendsAfterContentWithoutTrailingNewline(t *testing.T) {
	got, err := UpsertBlock("alias ll='ls -la'", "export FOO=bar")
	require.NoError(t, err)
	require.Equal(t, "alias ll='ls -la'\n\n"+StartMark+"\nexport FOO=bar\n"+EndMark+"\n", got)
}

func TestUpsertBlock_AppendsAfterContentWithTrailingNewline(t *testing.T) {
	got, err := UpsertBlock("alias ll='ls -la'\n", "export FOO=bar")
	require.NoError(t, err)
	require.Equal(t, "alias ll='ls -la'\n\n"+StartMark+"\nexport FOO=bar\n"+EndMark+"\n", got)
}

func TestUpsertBlock_NoExtraBlankLineWhenAlreadyPresent(t *testing.T) {
	got, err := UpsertBlock("alias ll='ls -la'\n\n", "export FOO=bar")
	require.NoError(t, err)
	require.Equal(t, "alias ll='ls -la'\n\n"+StartMark+"\nexport FOO=bar\n"+EndMark+"\n", got)
}

func TestUpsertBlock_ReplacesExistingBlock(t *testing.T) {
	existing := "before\n" + StartMark + "\nold content\n" + EndMark + "\nafter\n"
	got, err := UpsertBlock(existing, "new content")
	require.NoError(t, err)
	require.Equal(t, "before\n"+StartMark+"\nnew content\n"+EndMark+"\nafter\n", got)
}

func TestUpsertBlock_ReplacesExistingBlockNoTrailingNewlineAfterEnd(t *testing.T) {
	existing := "before\n" + StartMark + "\nold content\n" + EndMark
	got, err := UpsertBlock(existing, "new content")
	require.NoError(t, err)
	require.Equal(t, "before\n"+StartMark+"\nnew content\n"+EndMark+"\n", got)
}

func TestUpsertBlock_UnterminatedBlockReturnsError(t *testing.T) {
	existing := "before\n" + StartMark + "\nold content without end marker\n"
	_, err := UpsertBlock(existing, "new content")
	require.ErrorIs(t, err, ErrUnterminatedBlock)
}

func TestUpsertBlock_PreservesCRLFLineEndings(t *testing.T) {
	existing := "alias ll='ls -la'\r\n"
	got, err := UpsertBlock(existing, "export FOO=bar")
	require.NoError(t, err)
	require.True(t, strings.Contains(got, "\r\n"))
	require.False(t, strings.Contains(got, "\n\n")) // no bare LF should sneak in
	expected := "alias ll='ls -la'\r\n\r\n" + strings.ReplaceAll(StartMark, "\n", "\r\n") + "\r\nexport FOO=bar\r\n" + EndMark + "\r\n"
	require.Equal(t, expected, got)
}

func TestUpsertBlock_CRLFReplaceExistingBlock(t *testing.T) {
	existing := "before\r\n" + StartMark + "\r\nold content\r\n" + EndMark + "\r\nafter\r\n"
	got, err := UpsertBlock(existing, "new content")
	require.NoError(t, err)
	require.Equal(t, "before\r\n"+StartMark+"\r\nnew content\r\n"+EndMark+"\r\nafter\r\n", got)
}

func TestUpsertBlock_TrimsTrailingWhitespaceFromSnippet(t *testing.T) {
	got, err := UpsertBlock("", "export FOO=bar\n\n  \t\n")
	require.NoError(t, err)
	require.Equal(t, StartMark+"\nexport FOO=bar\n"+EndMark+"\n", got)
}

func TestUpsertBlock_NormalizesCRLFInSnippet(t *testing.T) {
	got, err := UpsertBlock("", "line1\r\nline2")
	require.NoError(t, err)
	require.Equal(t, StartMark+"\nline1\nline2\n"+EndMark+"\n", got)
}

func TestUpsertBlock_IdempotentOnSecondCall(t *testing.T) {
	first, err := UpsertBlock("before\n", "export FOO=bar")
	require.NoError(t, err)

	second, err := UpsertBlock(first, "export FOO=bar")
	require.NoError(t, err)
	require.Equal(t, first, second)
}
