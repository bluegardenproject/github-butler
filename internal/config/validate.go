package config

import (
	"fmt"
	"regexp"
	"strings"
)

// repoSlugPattern accepts GitHub-style owner/name pairs. Owners and repos
// may contain letters, digits, hyphens, underscores, and dots; both segments
// must be 1..100 characters.
var repoSlugPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}/[A-Za-z0-9._-]{1,100}$`)

var hexColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

// Validate checks the config for obvious mistakes. Repos are checked for
// the "owner/repo" slug format; poll interval is bounds-checked.
func Validate(cfg Config) error {
	if cfg.PollIntervalSeconds != 0 {
		if cfg.PollIntervalSeconds < MinPollIntervalSeconds || cfg.PollIntervalSeconds > MaxPollIntervalSeconds {
			return fmt.Errorf("poll_interval_seconds must be between %d and %d (got %d)",
				MinPollIntervalSeconds, MaxPollIntervalSeconds, cfg.PollIntervalSeconds)
		}
	}
	seen := make(map[string]struct{}, len(cfg.Repos))
	for _, r := range cfg.Repos {
		r = strings.TrimSpace(r)
		if !repoSlugPattern.MatchString(r) {
			return fmt.Errorf("invalid repo %q (expected owner/name)", r)
		}
		key := strings.ToLower(r)
		if _, dup := seen[key]; dup {
			return fmt.Errorf("duplicate repo %q", r)
		}
		seen[key] = struct{}{}
	}
	switch cfg.DashboardView {
	case "", DashboardViewAuto, DashboardViewFull, DashboardViewCompact:
	default:
		return fmt.Errorf("dashboard_view must be one of %q, %q, or %q (got %q)",
			DashboardViewAuto, DashboardViewFull, DashboardViewCompact, cfg.DashboardView)
	}
	if strings.TrimSpace(cfg.Theme.Selected) == "" {
		return fmt.Errorf("theme.selected must not be empty")
	}
	for key, value := range cfg.Theme.Colors {
		if !hexColorPattern.MatchString(strings.TrimSpace(value)) {
			return fmt.Errorf("theme.colors.%s must be a #RRGGBB hex color", key)
		}
	}
	return nil
}

// IsValidRepoSlug reports whether s is a valid owner/repo slug.
// Exposed so the UI can validate manual input before mutating the config.
func IsValidRepoSlug(s string) bool {
	return repoSlugPattern.MatchString(strings.TrimSpace(s))
}
