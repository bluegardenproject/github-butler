package ui

import (
	"context"
	"os/exec"
	"runtime"
	"time"

	"github.com/bluegardenproject/github-butler/internal/config"
	"github.com/bluegardenproject/github-butler/internal/github"
	tea "github.com/charmbracelet/bubbletea"
)

// Factory functions returning tea.Cmd values. Keeping them here separates
// the "how to trigger side effects" from model state transitions.

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

// uiTickInterval is how often we repaint the countdown bar and relative
// timestamps. Fast enough to feel live, slow enough to be cheap.
const uiTickInterval = 250 * time.Millisecond

func uiTickCmd() tea.Cmd {
	return tea.Tick(uiTickInterval, func(t time.Time) tea.Msg { return uiTickMsg(t) })
}

func fetchPRsCmd(client *github.Client, repos []string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		prs, err := client.FetchPRs(ctx, repos)
		if err != nil {
			return errMsg{Err: err}
		}
		return prsFetchedMsg{PRs: prs, Fetched: time.Now()}
	}
}

func validateRepoCmd(client *github.Client, slug string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := client.Validate(ctx, slug)
		return repoValidatedMsg{Slug: slug, Err: err}
	}
}

func saveConfigCmd(cfg config.Config) tea.Cmd {
	return func() tea.Msg {
		if err := config.Save(cfg); err != nil {
			return errMsg{Err: err}
		}
		return configSavedMsg{}
	}
}

// openURLCmd opens a URL in the user's default browser using the OS-native
// open tool. Errors are surfaced as errMsg but don't disrupt navigation.
func openURLCmd(url string) tea.Cmd {
	return func() tea.Msg {
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}
		if err := cmd.Start(); err != nil {
			return errMsg{Err: err}
		}
		return nil
	}
}
