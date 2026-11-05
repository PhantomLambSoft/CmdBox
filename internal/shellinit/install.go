package shellinit

import (
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
)

type InstallResult struct {
	Path       string
	Changed    bool
	BackupPath string
}

type Initializer struct {
	Env Env
	FS  Filesystem
}

func New() *Initializer {
	return &Initializer{Env: SystemEnv(), FS: OSFS{}}
}

func (i *Initializer) DetectShell() string {
	return Detect(i.Env)
}

func (i *Initializer) Resolve(shell string) (Spec, error) {
	if strings.TrimSpace(shell) == "" {
		shell = i.DetectShell()
	}
	return Lookup(shell)
}

func (i *Initializer) Snippet(spec Spec) (string, error) {
	return LoadIntegration(spec.Integration)
}

func (i *Initializer) TargetPath(spec Spec, override string) (string, error) {
	if override != "" {
		return i.expandHome(override), nil
	}
	if spec.DefaultPath == nil {
		return "", &NoDefaultPathError{Shell: spec.Display}
	}
	return spec.DefaultPath(i.Env)
}

func (i *Initializer) expandHome(p string) string {
	if i.Env.Home == "" {
		return p
	}
	if p == "~" {
		return i.Env.Home
	}
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) {
		return filepath.Join(i.Env.Home, p[2:])
	}
	return p
}

// Install writes or updates the shell integration snippet at the resolved target path. It backs up existing files
// before changes and reports whether the target changed.
func (i *Initializer) Install(spec Spec, snippet, override string) (InstallResult, error) {
	target, err := i.TargetPath(spec, override)
	if err != nil {
		return InstallResult{}, err
	}
	res := InstallResult{Path: target}

	existing, existed, err := i.read(target)
	if err != nil {
		return res, err
	}

	var updated string
	switch spec.Mode {
	case ModeProfileBlock:
		updated, err = UpsertBlock(existing, snippet)
		if err != nil {
			return res, err
		}
	case ModeWriteFile:
		updated = snippet
	default:
		return res, errors.New("shell cannot be installed automatically")
	}

	if existed && updated == existing {
		res.Changed = false
	} else {
		if existed {
			backup := target + ".bak"
			if err := i.FS.CopyFile(target, backup); err != nil {
				return res, err
			}
			res.BackupPath = backup
		}
		if err := i.FS.WriteFile(target, []byte(updated), 0o644); err != nil {
			return res, err
		}
		res.Changed = true
	}

	return res, nil
}

func (i *Initializer) read(path string) (text string, existed bool, err error) {
	data, err := i.FS.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return string(data), true, nil
}
