//go:build windows

package shellinit

import (
	"fmt"
	"syscall"
	"unsafe"
)

// lookupProcess retrieves the parent process ID and executable name for a given process ID (pid).
// Returns the parent process ID, the process name, and an error if the process is not found or an error occurs.
func lookupProcess(pid int) (int, string, error) {
	snap, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, "", err
	}
	defer syscall.CloseHandle(snap)

	var entry syscall.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	if err := syscall.Process32First(snap, &entry); err != nil {
		return 0, "", err
	}
	for {
		if int(entry.ProcessID) == pid {
			return int(entry.ParentProcessID), syscall.UTF16ToString(entry.ExeFile[:]), nil
		}
		if err := syscall.Process32Next(snap, &entry); err != nil {
			break
		}
	}
	return 0, "", fmt.Errorf("process %d not found", pid)
}
