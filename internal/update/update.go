// Package update checks GitHub releases and reruns the installer when the
// user accepts a self-update prompt.
package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Repo is the GitHub repo queried for releases. Tests may point it at a stub.
var Repo = "bluegardenproject/github-butler"

// InstallScriptURL is the public one-line installer target.
const InstallScriptURL = "https://raw.githubusercontent.com/bluegardenproject/github-butler/main/scripts/install.sh"

const (
	installDirName = ".github-butler"
	binaryName     = "github-butler"
)

// Release is the subset of the GitHub releases API payload we need.
type Release struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	HTMLURL string `json:"html_url"`
}

// LatestRelease fetches the latest published release.
func LatestRelease(ctx context.Context) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", Repo)

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.New("no published releases found")
	}
	if resp.StatusCode/100 != 2 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("github api %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var rel Release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("decode release: %w", err)
	}
	return &rel, nil
}

// Compare returns -1, 0, or +1 like strings.Compare, treating both versions as
// dot-separated numeric versions. A leading "v" is ignored.
func Compare(a, b string) int {
	if a == b {
		return 0
	}

	aDev := isDev(a)
	bDev := isDev(b)
	switch {
	case aDev && bDev:
		return 0
	case aDev:
		return -1
	case bDev:
		return 1
	}

	aParts := strings.Split(strings.TrimPrefix(a, "v"), ".")
	bParts := strings.Split(strings.TrimPrefix(b, "v"), ".")

	n := len(aParts)
	if len(bParts) > n {
		n = len(bParts)
	}
	for len(aParts) < n {
		aParts = append(aParts, "0")
	}
	for len(bParts) < n {
		bParts = append(bParts, "0")
	}

	for i := 0; i < n; i++ {
		ai, aErr := strconv.Atoi(aParts[i])
		bi, bErr := strconv.Atoi(bParts[i])
		if aErr == nil && bErr == nil {
			switch {
			case ai < bi:
				return -1
			case ai > bi:
				return 1
			}
			continue
		}
		if c := strings.Compare(aParts[i], bParts[i]); c != 0 {
			return c
		}
	}
	return 0
}

func isDev(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "" || v == "dev" || v == "unknown"
}

// Supported reports whether the bundled installer can replace the current
// platform's binary.
func Supported() bool {
	return runtime.GOOS != "windows"
}

// InstalledBinaryPath returns the location written by scripts/install.sh.
func InstalledBinaryPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, installDirName, binaryName), nil
}

// Run executes the install script for the current platform.
func Run(ctx context.Context) error {
	if !Supported() {
		return errors.New("self-update is not supported on Windows")
	}

	cmd := exec.CommandContext(ctx, "bash", "-c",
		fmt.Sprintf("curl -fsSL %s | bash", InstallScriptURL))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
