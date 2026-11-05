//go:build linux

package shellinit

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// lookupProcess retrieves the parent process ID (PPID) and name of a process given its process ID (PID).
func lookupProcess(pid int) (int, string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, "", err
	}

	s := string(data)
	open, closeIndex := strings.IndexByte(s, '('), strings.LastIndexByte(s, ')')
	if open < 0 || closeIndex < open {
		return 0, "", fmt.Errorf("unexpected /proc/%d/stat format", pid)
	}

	fields := strings.Fields(s[closeIndex+1:])
	if len(fields) < 2 {
		return 0, "", fmt.Errorf("unexpected /proc/%d/stat format", pid)
	}
	ppid, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, "", err
	}

	return ppid, s[open+1 : closeIndex], nil
}
