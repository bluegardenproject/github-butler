package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bluegardenproject/github-butler/internal/update"
)

func maybePromptForUpdate(ctx context.Context) (bool, error) {
	if version == "dev" || version == "unknown" {
		return false, nil
	}
	if os.Getenv("NO_UPDATE_NOTIFIER") != "" {
		return false, nil
	}
	if !update.Supported() {
		return false, nil
	}
	if !isTerminal(os.Stdin) || !isTerminal(os.Stdout) {
		return false, nil
	}

	rel, err := update.LatestRelease(ctx)
	if err != nil {
		return false, nil
	}
	if update.Compare(version, rel.TagName) >= 0 {
		return false, nil
	}

	fmt.Fprintf(os.Stdout, "There is a new version %s available. Update now Y/N ", rel.TagName)
	answer, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stdout)
		return false, nil
	}

	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
	default:
		return false, nil
	}

	fmt.Fprintln(os.Stdout, "Updating github-butler...")
	if err := update.Run(ctx); err != nil {
		return true, err
	}
	fmt.Fprintf(os.Stdout, "Update complete. Restarting github-butler %s...\n", rel.TagName)
	if err := update.Restart(os.Args[1:]); err != nil {
		return true, err
	}
	return true, nil
}

func isTerminal(file *os.File) bool {
	info, err := file.Stat()
	return err == nil && (info.Mode()&os.ModeCharDevice) != 0
}
