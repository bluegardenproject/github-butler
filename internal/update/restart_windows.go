//go:build windows

package update

import "errors"

// Restart is unsupported on Windows because the bundled installer is Unix-only.
func Restart(args []string) error {
	return errors.New("self-update restart is not supported on Windows")
}
