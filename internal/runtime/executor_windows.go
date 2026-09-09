//go:build windows

package runtime

import (
	"os/exec"
	"strconv"
)

func configureProcessGroup(cmd *exec.Cmd) {
	// No pre-start setup needed
}

func killProcessTree(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cmd.Process.Pid)).Run()
}
