package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/bluegardenproject/github-butler/internal/config"
	"github.com/bluegardenproject/github-butler/internal/github"
	"github.com/charmbracelet/lipgloss"
)

func TestUseCompactTable(t *testing.T) {
	m := Model{cfg: config.Config{DashboardView: config.DashboardViewAuto}, width: 120}
	if !m.useCompactTable() {
		t.Fatal("auto view should use compact table below the width threshold")
	}

	m.width = 180
	if m.useCompactTable() {
		t.Fatal("auto view should use full table above the width threshold")
	}

	m.cfg.DashboardView = config.DashboardViewCompact
	m.width = 220
	if !m.useCompactTable() {
		t.Fatal("compact view should always use compact table")
	}

	m.cfg.DashboardView = config.DashboardViewFull
	m.width = 80
	if m.useCompactTable() {
		t.Fatal("full view should never use compact table")
	}
}

func TestRenderCompactRowUsesSingleDenseStatusLine(t *testing.T) {
	m := Model{}
	pr := github.PR{
		Repo:            "ledger/github-butler",
		Number:          42,
		Title:           "Add compact dashboard layout for split terminal panes",
		HeadRef:         "feature/compact-dashboard",
		CreatedAt:       time.Now().Add(-48 * time.Hour),
		UpdatedAt:       time.Now().Add(-2 * time.Hour),
		ReviewDecision:  "REVIEW_REQUIRED",
		TotalChecks:     3,
		PendingChecks:   1,
		UnresolvedCount: 2,
		RequiredReviews: []github.RequiredReview{{State: github.ReviewStatePending, RequiredByCodeOwner: true}},
	}

	row := m.renderCompactRow(pr, true, 96)
	if !strings.Contains(row, "#42") || !strings.Contains(row, "Add compact dashboard") {
		t.Fatalf("row missing PR identity: %q", row)
	}
	if lipgloss.Width(row) > 96 {
		t.Fatalf("row width = %d, want <= 96", lipgloss.Width(row))
	}
	for _, token := range []string{"RUN×1", "REQ 0/1", "UNRES 2"} {
		if !strings.Contains(row, token) {
			t.Fatalf("row missing status token %s: %q", token, row)
		}
	}
}

func TestCompactStatus(t *testing.T) {
	cases := []struct {
		name string
		pr   github.PR
		want []string
	}{
		{
			name: "passing approved",
			pr: github.PR{
				TotalChecks:     1,
				ReviewDecision:  "APPROVED",
				RequiredReviews: []github.RequiredReview{{State: github.ReviewStateApproved, RequiredByCodeOwner: true}},
			},
			want: []string{"PASS", "OK"},
		},
		{
			name: "failing changes unresolved",
			pr: github.PR{
				TotalChecks:     2,
				FailingChecks:   []github.Check{{Name: "test"}},
				UnresolvedCount: 3,
				RequiredReviews: []github.RequiredReview{{State: github.ReviewStateChangesRequested, RequiredByCodeOwner: true}},
			},
			want: []string{"FAIL×1", "CHG", "UNRES 3"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := compactStatus(tc.pr)
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Fatalf("compactStatus() = %q, want token %q", got, want)
				}
			}
		})
	}
}
