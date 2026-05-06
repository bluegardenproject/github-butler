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
	return nil
}

// IsValidRepoSlug reports whether s is a valid owner/repo slug.
// Exposed so the UI can validate manual input before mutating the config.
func IsValidRepoSlug(s string) bool {
	return repoSlugPattern.MatchString(strings.TrimSpace(s))
}
