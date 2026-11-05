package shellinit

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Env struct {
	GOOS   string
	Home   string
	Ppid   int
	GetEnv func(string) string
	Output func(name string, args ...string) (string, error)
	Parent func(pid int) (ppid int, name string, err error)
}

func SystemEnv() Env {
	home, _ := os.UserHomeDir()
	return Env{
		GOOS:   runtime.GOOS,
		Home:   home,
		Ppid:   os.Getppid(),
		GetEnv: os.Getenv,
		Output: runOutput,
		Parent: lookupProcess,
	}
}

func (e Env) env(key string) string {
	if e.GetEnv == nil {
		return ""
	}
	return e.GetEnv(key)
}

func runOutput(name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
