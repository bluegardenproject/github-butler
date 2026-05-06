package github

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ErrRepoNotFound is returned when Validate fails because the repository
// doesn't exist or the authenticated user can't access it.
var ErrRepoNotFound = errors.New("repository not found or inaccessible")

var (
	httpsRepoRe = regexp.MustCompile(`^https?://github\.com/([^/\s]+)/([^/\s]+?)(?:\.git)?/?$`)
	sshRepoRe   = regexp.MustCompile(`^git@github\.com:([^/\s]+)/([^/\s]+?)(?:\.git)?$`)
	slugRepoRe  = regexp.MustCompile(`^([A-Za-z0-9][A-Za-z0-9._-]{0,99})/([A-Za-z0-9._-]{1,100})$`)
)

// ParseRepoRef normalizes any of the supported input formats (HTTPS URL,
// SSH URL, owner/name slug) to a canonical "owner/name" string.
func ParseRepoRef(input string) (string, error) {
	s := strings.TrimSpace(input)
	if s == "" {
		return "", fmt.Errorf("empty repo reference")
	}
	for _, re := range []*regexp.Regexp{httpsRepoRe, sshRepoRe, slugRepoRe} {
		if m := re.FindStringSubmatch(s); m != nil {
			return m[1] + "/" + m[2], nil
		}
	}
	return "", fmt.Errorf("not a recognised GitHub repo reference: %q", input)
}

// Validate confirms the repo exists and the gh-authenticated user can
// access it by calling `gh api repos/:owner/:name`. Returns ErrRepoNotFound
// on 404/403.
func (c *Client) Validate(ctx context.Context, ownerRepo string) error {
	if _, err := ParseRepoRef(ownerRepo); err != nil {
		return err
	}
	_, err := c.Runner.Run(ctx, "api", "repos/"+ownerRepo, "--silent")
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "404") || strings.Contains(msg, "Not Found") ||
		strings.Contains(msg, "403") || strings.Contains(msg, "Forbidden") {
		return ErrRepoNotFound
	}
	return err
}
