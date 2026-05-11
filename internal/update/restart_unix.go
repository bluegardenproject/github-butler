//go:build !windows

package update

import (
	"os"
	"syscall"
)

// Restart replaces the current process with the installed github-butler binary.
func Restart(args []string) error {
	path, err := InstalledBinaryPath()
	if err != nil {
		return err
	}
	return syscall.Exec(path, append([]string{path}, args...), os.Environ())
}
