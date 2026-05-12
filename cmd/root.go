// Package cmd wires the config, GitHub client, and UI into a single
// runnable program. Keeping this separate from main.go makes it trivial
// to add subcommands later without touching the UI layer.
package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/bluegardenproject/github-butler/internal/config"
	"github.com/bluegardenproject/github-butler/internal/github"
	"github.com/bluegardenproject/github-butler/internal/ui"
	"github.com/bluegardenproject/github-butler/internal/ui/theme"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

// SetVersion stores the build-time version metadata so the `--version`
// flag and any future `version` subcommand can render it consistently.
// Called from main() before Run().
func SetVersion(v, bt string) {
	if v != "" {
		version = v
	}
	if bt != "" {
		buildTime = bt
	}
}

// Run parses flags, loads config, and starts the Bubble Tea program.
func Run(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("github-butler", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "path to config file (default: ~/.config/github-butler/config.yaml)")
	noColor := fs.Bool("no-color", false, "disable all color output")
	showVersion := fs.Bool("version", false, "print version information and exit")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	if *showVersion {
		fmt.Printf("github-butler %s\n", version)
		fmt.Printf("  Built:    %s\n", buildTime)
		fmt.Printf("  Go:       %s\n", runtime.Version())
		fmt.Printf("  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		return nil
	}

	if *noColor {
		_ = os.Setenv("NO_COLOR", "1")
	}

	if updated, err := maybePromptForUpdate(ctx); err != nil {
		return fmt.Errorf("update github-butler: %w", err)
	} else if updated {
		return nil
	}

	if _, err := exec.LookPath("gh"); err != nil {
		return fmt.Errorf("the GitHub CLI (`gh`) was not found in PATH; install it and run `gh auth login` first")
	}

	path := *cfgPath
	if path == "" {
		var err error
		path, err = config.DefaultPath()
		if err != nil {
			return fmt.Errorf("resolving default config path: %w", err)
		}
	}

	cfg, err := config.Load(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = config.Default()
			cfg.Path = path
		} else {
			return fmt.Errorf("loading config: %w", err)
		}
	}
	cfg.Path = path
	cfg.Theme.Directory = cfg.ThemeDirectory()

	if err := theme.SeedExamples(cfg.Theme.Directory); err != nil {
		return fmt.Errorf("preparing themes: %w", err)
	}
	if err := theme.Activate(theme.OptionsFromConfig(cfg.Theme)); err != nil {
		return fmt.Errorf("loading theme: %w", err)
	}

	client := github.NewClient()
	model := ui.NewModel(cfg, client, theme.Choices(cfg.Theme.Directory))

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithContext(ctx))
	_, err = p.Run()
	return err
}
