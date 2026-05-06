package github

import (
	_ "embed"
	"strings"
)

//go:embed query.graphql
var searchQuery string

// SearchQuery returns the embedded GraphQL query used by FetchPRs.
func SearchQuery() string {
	return searchQuery
}

// BuildSearchString builds the GitHub search qualifier string used as the
// GraphQL `$q` variable. It always filters to open PRs authored by the
// authenticated user and restricts to the given repos.
//
// Example output:
//
//	is:pr is:open author:@me repo:acme/web repo:acme/api
func BuildSearchString(repos []string) string {
	var b strings.Builder
	b.WriteString("is:pr is:open author:@me")
	for _, r := range repos {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		b.WriteString(" repo:")
		b.WriteString(r)
	}
	return b.String()
}
