package github

import (
	"testing"
)

func TestMergeRequiredReviews_CodeOwnerPendingTeam(t *testing.T) {
	requests := []rawReviewRequest{
		{AsCodeOwner: true, RequestedReviewer: struct {
			Typename     string `json:"__typename"`
			Login        string `json:"login"`
			Name         string `json:"name"`
			CombinedSlug string `json:"combinedSlug"`
		}{Typename: "Team", CombinedSlug: "acme/backend"}},
	}
	got := mergeRequiredReviews(requests, nil)
	if len(got) != 1 {
		t.Fatalf("want 1 review, got %d", len(got))
	}
	r := got[0]
	if r.Name != "@acme/backend" || r.Kind != ReviewerKindTeam || !r.RequiredByCodeOwner || r.State != ReviewStatePending {
		t.Fatalf("unexpected: %#v", r)
	}
}

func TestMergeRequiredReviews_ApprovedOnBehalfOfTeam(t *testing.T) {
	reviews := []rawLatestReview{
		{
			State: "APPROVED",
			Author: struct {
				Login string `json:"login"`
			}{Login: "alice"},
			OnBehalfOf: struct {
				Nodes []struct {
					Name         string `json:"name"`
					CombinedSlug string `json:"combinedSlug"`
				} `json:"nodes"`
			}{Nodes: []struct {
				Name         string `json:"name"`
				CombinedSlug string `json:"combinedSlug"`
			}{{CombinedSlug: "acme/frontend"}}},
		},
	}
	got := mergeRequiredReviews(nil, reviews)
	if len(got) != 2 {
		t.Fatalf("want 2 entries (user + team), got %d", len(got))
	}

	var team *RequiredReview
	for i := range got {
		if got[i].Kind == ReviewerKindTeam {
			team = &got[i]
		}
	}
	if team == nil {
		t.Fatal("team entry missing")
	}
	if team.State != ReviewStateApproved {
		t.Fatalf("team state = %s, want APPROVED", team.State)
	}
	if team.ApprovedByLogin != "alice" {
		t.Fatalf("approved-by = %q, want alice", team.ApprovedByLogin)
	}
}

func TestMergeRequiredReviews_ChangesRequestedWins(t *testing.T) {
	reviews := []rawLatestReview{
		{State: "COMMENTED", Author: struct {
			Login string `json:"login"`
		}{Login: "bob"}},
		{State: "CHANGES_REQUESTED", Author: struct {
			Login string `json:"login"`
		}{Login: "bob"}},
	}
	got := mergeRequiredReviews(nil, reviews)
	if len(got) != 1 {
		t.Fatalf("want 1 entry, got %d", len(got))
	}
	if got[0].State != ReviewStateChangesRequested {
		t.Fatalf("state = %s, want CHANGES_REQUESTED", got[0].State)
	}
}

func TestParseRepoRef(t *testing.T) {
	cases := map[string]string{
		"acme/web":                         "acme/web",
		"  acme/web  ":                     "acme/web",
		"https://github.com/acme/web":      "acme/web",
		"https://github.com/acme/web.git":  "acme/web",
		"https://github.com/acme/web/":     "acme/web",
		"http://github.com/acme/web":       "acme/web",
		"git@github.com:acme/web.git":      "acme/web",
		"git@github.com:acme/web":          "acme/web",
		"https://github.com/a.c/web-thing": "a.c/web-thing",
	}
	for in, want := range cases {
		got, err := ParseRepoRef(in)
		if err != nil {
			t.Errorf("ParseRepoRef(%q) error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseRepoRef(%q) = %q, want %q", in, got, want)
		}
	}

	bad := []string{"", "not-a-repo", "https://gitlab.com/a/b", "owner/", "/repo"}
	for _, in := range bad {
		if _, err := ParseRepoRef(in); err == nil {
			t.Errorf("ParseRepoRef(%q) want error, got nil", in)
		}
	}
}

func TestMapCheckRunState(t *testing.T) {
	cases := []struct {
		status, conclusion string
		want               CheckState
	}{
		{"COMPLETED", "SUCCESS", CheckStateSuccess},
		{"COMPLETED", "FAILURE", CheckStateFailure},
		{"COMPLETED", "TIMED_OUT", CheckStateFailure},
		{"IN_PROGRESS", "", CheckStatePending},
		{"QUEUED", "", CheckStatePending},
		{"COMPLETED", "NEUTRAL", CheckStateNeutral},
		{"COMPLETED", "SKIPPED", CheckStateSkipped},
	}
	for _, tc := range cases {
		if got := mapCheckRunState(tc.status, tc.conclusion); got != tc.want {
			t.Errorf("mapCheckRunState(%q,%q) = %s, want %s", tc.status, tc.conclusion, got, tc.want)
		}
	}
}
