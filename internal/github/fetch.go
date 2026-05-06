package github

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"
)

// Client is the high-level entry point for fetching PR data. It depends only
// on a GhRunner, which keeps it testable.
type Client struct {
	Runner GhRunner
}

// NewClient returns a Client backed by an ExecRunner using the `gh` binary
// on $PATH.
func NewClient() *Client { return &Client{Runner: NewExecRunner()} }

// rawResponse mirrors the shape returned by `gh api graphql`.
type rawResponse struct {
	Data struct {
		Search struct {
			Nodes []rawNode `json:"nodes"`
		} `json:"search"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type rawNode struct {
	Typename         string    `json:"__typename"`
	Number           int       `json:"number"`
	Title            string    `json:"title"`
	URL              string    `json:"url"`
	HeadRefName      string    `json:"headRefName"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
	IsDraft          bool      `json:"isDraft"`
	ReviewDecision   string    `json:"reviewDecision"`
	Mergeable        string    `json:"mergeable"`
	MergeStateStatus string    `json:"mergeStateStatus"`
	Repository       struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	Commits struct {
		Nodes []struct {
			Commit struct {
				StatusCheckRollup *struct {
					State    string `json:"state"`
					Contexts struct {
						Nodes []rawContext `json:"nodes"`
					} `json:"contexts"`
				} `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
	ReviewThreads struct {
		Nodes []struct {
			IsResolved bool `json:"isResolved"`
		} `json:"nodes"`
	} `json:"reviewThreads"`
	ReviewRequests struct {
		Nodes []rawReviewRequest `json:"nodes"`
	} `json:"reviewRequests"`
	LatestReviews struct {
		Nodes []rawLatestReview `json:"nodes"`
	} `json:"latestReviews"`
}

type rawContext struct {
	Typename string `json:"__typename"`
	// CheckRun fields
	Name       string `json:"name"`
	Conclusion string `json:"conclusion"`
	Status     string `json:"status"`
	DetailsURL string `json:"detailsUrl"`
	// StatusContext fields
	Context   string `json:"context"`
	State     string `json:"state"`
	TargetURL string `json:"targetUrl"`
}

type rawReviewRequest struct {
	AsCodeOwner       bool `json:"asCodeOwner"`
	RequestedReviewer struct {
		Typename     string `json:"__typename"`
		Login        string `json:"login"`
		Name         string `json:"name"`
		CombinedSlug string `json:"combinedSlug"`
	} `json:"requestedReviewer"`
}

type rawLatestReview struct {
	State  string `json:"state"`
	Author struct {
		Login string `json:"login"`
	} `json:"author"`
	OnBehalfOf struct {
		Nodes []struct {
			Name         string `json:"name"`
			CombinedSlug string `json:"combinedSlug"`
		} `json:"nodes"`
	} `json:"onBehalfOf"`
}

// FetchPRs runs the embedded search query against the given repos and
// returns fully-derived PRs. If repos is empty, returns an empty slice
// without calling the API.
func (c *Client) FetchPRs(ctx context.Context, repos []string) ([]PR, error) {
	if len(repos) == 0 {
		return []PR{}, nil
	}

	q := BuildSearchString(repos)
	out, err := c.Runner.Run(ctx,
		"api", "graphql",
		"-f", "query="+SearchQuery(),
		"-F", "q="+q,
	)
	if err != nil {
		return nil, err
	}

	var resp rawResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("parsing gh response: %w", err)
	}
	if len(resp.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", resp.Errors[0].Message)
	}

	prs := make([]PR, 0, len(resp.Data.Search.Nodes))
	for _, n := range resp.Data.Search.Nodes {
		if n.Typename != "PullRequest" {
			continue
		}
		prs = append(prs, toPR(n))
	}
	sort.SliceStable(prs, func(i, j int) bool {
		return prs[i].UpdatedAt.After(prs[j].UpdatedAt)
	})
	return prs, nil
}
