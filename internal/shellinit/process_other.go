//go:build !linux && !windows

package shellinit

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func lookupProcess(pid int) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ps", "-o", "ppid=,comm=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return 0, "", err
	}

	line := strings.TrimSpace(string(out))
	ppidStr, name, ok := strings.Cut(line, " ")
	if !ok {
		return 0, "", fmt.Errorf("unexpected ps output for pid %d", pid)
	}

	ppid, err := strconv.Atoi(ppidStr)
	if err != nil {
		return 0, "", err
	}

	return ppid, strings.TrimSpace(name), nil
}
