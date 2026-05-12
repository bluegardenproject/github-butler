// Package config owns the on-disk YAML configuration: defaults, loading,
// atomic saving, and validation. The UI and GitHub packages only see an
// already-validated Config value.
package config

import (
	"os"
	"path/filepath"
	"time"
)

const (
	DefaultPollInterval    = 20 * time.Second
	MinPollIntervalSeconds = 2
	MaxPollIntervalSeconds = 3600
)

// Config is the user-facing configuration persisted as YAML.
type Config struct {
	Repos               []string `yaml:"repos"`
	PollIntervalSeconds int      `yaml:"poll_interval_seconds"`
	GroupByRepo         bool     `yaml:"group_by_repo"`
	Theme               Theme    `yaml:"theme"`

	// Path is the file this config was loaded from (or should be saved to).
	// Not persisted to YAML.
	Path string `yaml:"-"`
}

// Theme is the user-facing theme configuration persisted as YAML.
type Theme struct {
	Selected  string            `yaml:"selected"`
	Directory string            `yaml:"directory"`
	Colors    map[string]string `yaml:"colors,omitempty"`
}

// PollInterval returns the poll interval as a time.Duration, falling back
// to the default if the stored value is zero or outside the allowed range.
func (c Config) PollInterval() time.Duration {
	s := c.PollIntervalSeconds
	if s < MinPollIntervalSeconds || s > MaxPollIntervalSeconds {
		return DefaultPollInterval
	}
	return time.Duration(s) * time.Second
}

// Default returns a Config with sensible defaults but no repos.
func Default() Config {
	themeDir, _ := DefaultThemeDirectory()
	return Config{
		Repos:               []string{},
		PollIntervalSeconds: int(DefaultPollInterval / time.Second),
		Theme: Theme{
			Selected:  "auto",
			Directory: themeDir,
		},
	}
}

// ThemeDirectory returns the directory where user theme files live. An empty
// config value falls back to the default XDG-ish theme directory.
func (c Config) ThemeDirectory() string {
	if c.Theme.Directory != "" {
		return c.Theme.Directory
	}
	dir, _ := DefaultThemeDirectory()
	return dir
}

// DefaultPath returns the standard XDG-ish config file path:
// $XDG_CONFIG_HOME/github-butler/config.yaml, falling back to
// $HOME/.config/github-butler/config.yaml.
func DefaultPath() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "github-butler", "config.yaml"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "github-butler", "config.yaml"), nil
}

// DefaultThemeDirectory returns the standard directory for user theme files.
func DefaultThemeDirectory() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "github-butler", "themes"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "github-butler", "themes"), nil
}
