package github

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// GhRunner executes `gh` subcommands. The production implementation shells
// out; tests inject fixtures by implementing this interface.
type GhRunner interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
}

// ExecRunner runs `gh` via os/exec.
type ExecRunner struct {
	Binary string // defaults to "gh"
}

// NewExecRunner returns an ExecRunner using the gh binary on $PATH.
func NewExecRunner() *ExecRunner { return &ExecRunner{Binary: "gh"} }

// Run invokes `gh args...` and returns stdout. On failure the error message
// includes stderr from gh so the user can see authentication/permission issues.
func (r *ExecRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	bin := r.Binary
	if bin == "" {
		bin = "gh"
	}
	cmd := exec.CommandContext(ctx, bin, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := stderr.String()
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("gh %s: %s", argsPreview(args), msg)
	}
	return stdout.Bytes(), nil
}

func argsPreview(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}
