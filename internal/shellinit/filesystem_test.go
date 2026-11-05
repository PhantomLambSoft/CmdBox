package shellinit

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOSFS_WriteAndReadFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile")

	var fsys OSFS
	require.NoError(t, fsys.WriteFile(path, []byte("hello"), 0o644))

	data, err := fsys.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "hello", string(data))
}

func TestOSFS_ReadFileMissing(t *testing.T) {
	dir := t.TempDir()
	var fsys OSFS
	_, err := fsys.ReadFile(filepath.Join(dir, "missing"))
	require.True(t, os.IsNotExist(err))
}

func TestOSFS_WriteFileOverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile")
	require.NoError(t, os.WriteFile(path, []byte("old"), 0o644))

	var fsys OSFS
	require.NoError(t, fsys.WriteFile(path, []byte("new"), 0o644))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "new", string(data))
}

func TestOSFS_CopyFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dest := filepath.Join(dir, "dest")
	require.NoError(t, os.WriteFile(src, []byte("content"), 0o644))

	var fsys OSFS
	require.NoError(t, fsys.CopyFile(src, dest))

	data, err := os.ReadFile(dest)
	require.NoError(t, err)
	require.Equal(t, "content", string(data))

	// src should be untouched
	srcData, err := os.ReadFile(src)
	require.NoError(t, err)
	require.Equal(t, "content", string(srcData))
}

func TestOSFS_CopyFileMissingSrc(t *testing.T) {
	dir := t.TempDir()
	var fsys OSFS
	err := fsys.CopyFile(filepath.Join(dir, "missing"), filepath.Join(dir, "dest"))
	require.Error(t, err)
}

func TestOSFS_Rename(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old")
	newPath := filepath.Join(dir, "new")
	require.NoError(t, os.WriteFile(oldPath, []byte("content"), 0o644))

	var fsys OSFS
	require.NoError(t, fsys.Rename(oldPath, newPath))

	_, err := os.Stat(oldPath)
	require.True(t, os.IsNotExist(err))

	data, err := os.ReadFile(newPath)
	require.NoError(t, err)
	require.Equal(t, "content", string(data))
}
