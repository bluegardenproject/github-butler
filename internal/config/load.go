package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Load reads the YAML file at path, applies defaults for missing fields,
// and validates the result. If the file does not exist, a default Config
// is returned (with the path set) so the user can populate it via the UI.
func Load(path string) (Config, error) {
	path, err := expandHome(path)
	if err != nil {
		return Config{}, err
	}

	cfg := Default()
	cfg.Path = path

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("reading %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	cfg.Path = path

	if cfg.PollIntervalSeconds == 0 {
		cfg.PollIntervalSeconds = int(DefaultPollInterval / 1e9)
	}
	if cfg.Repos == nil {
		cfg.Repos = []string{}
	}
	if cfg.DashboardView == "" {
		cfg.DashboardView = DashboardViewAuto
	}
	if cfg.Theme.Selected == "" {
		cfg.Theme.Selected = "auto"
	}
	if cfg.Theme.Directory == "" {
		dir, err := DefaultThemeDirectory()
		if err != nil {
			return Config{}, err
		}
		cfg.Theme.Directory = dir
	}
	if cfg.Theme.Colors == nil {
		cfg.Theme.Colors = map[string]string{}
	}

	if err := Validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func expandHome(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~")), nil
}
