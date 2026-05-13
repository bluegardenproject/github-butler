package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDefaultsThemeConfig(t *testing.T) {
	xdg := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", xdg)

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("repos: []\npoll_interval_seconds: 20\ngroup_by_repo: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Theme.Selected != "auto" {
		t.Fatalf("Theme.Selected = %q, want auto", cfg.Theme.Selected)
	}
	if cfg.DashboardView != DashboardViewAuto {
		t.Fatalf("DashboardView = %q, want auto", cfg.DashboardView)
	}
	wantDir := filepath.Join(xdg, "github-butler", "themes")
	if cfg.Theme.Directory != wantDir {
		t.Fatalf("Theme.Directory = %q, want %q", cfg.Theme.Directory, wantDir)
	}
}

func TestValidateDashboardView(t *testing.T) {
	cfg := Default()
	for _, view := range []string{DashboardViewAuto, DashboardViewFull, DashboardViewCompact} {
		cfg.DashboardView = view
		if err := Validate(cfg); err != nil {
			t.Fatalf("Validate(%q) error: %v", view, err)
		}
	}

	cfg.DashboardView = "tiny"
	err := Validate(cfg)
	if err == nil {
		t.Fatal("Validate succeeded, want error")
	}
	if !strings.Contains(err.Error(), "dashboard_view") {
		t.Fatalf("Validate error = %q, want dashboard_view path", err)
	}
}

func TestValidateThemeColors(t *testing.T) {
	cfg := Default()
	cfg.Theme.Colors = map[string]string{"accent": "not-a-color"}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("Validate succeeded, want error")
	}
	if !strings.Contains(err.Error(), "theme.colors.accent") {
		t.Fatalf("Validate error = %q, want theme color path", err)
	}
}
