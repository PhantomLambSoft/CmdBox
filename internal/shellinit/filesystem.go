package shellinit

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/PhantomLambSoft/CmdBox/internal/atomicfile"
)

type Filesystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm fs.FileMode) error
	CopyFile(src, dest string) error
	Rename(oldPath, newPath string) error
}

type OSFS struct{}

func (OSFS) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (OSFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	target := path
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		target = resolved
	}
	mode := perm
	if st, err := os.Stat(target); err == nil {
		mode = st.Mode().Perm()
	}
	if err := atomicfile.WriteFile(target, string(data)); err != nil {
		return err
	}
	_ = os.Chmod(target, mode) // best effort - only toggles read-only on Windows
	return nil
}

func (o OSFS) CopyFile(src, dest string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	mode := fs.FileMode(0o600)
	if st, err := os.Stat(src); err == nil {
		mode = st.Mode().Perm()
	}
	return o.WriteFile(dest, data, mode)
}

func (OSFS) Rename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}
