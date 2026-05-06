// Package github is a thin wrapper around the local `gh` CLI. It speaks
// GraphQL via `gh api graphql`, exposes typed PR data to the rest of the
// app, and knows nothing about the UI.
package github

import (
	"fmt"
	"time"
)

// StaleAfter is how long without activity (no pushes, comments, reviews,
// etc.) a PR is considered stale.
const StaleAfter = 28 * 24 * time.Hour

// ReviewState mirrors GitHub's PullRequestReviewState enum but is kept
// as a string so we can attach "PENDING" (not a real review state) for
// reviewers that were requested but haven't reviewed yet.
type ReviewState string

const (
	ReviewStatePending          ReviewState = "PENDING"
	ReviewStateApproved         ReviewState = "APPROVED"
	ReviewStateChangesRequested ReviewState = "CHANGES_REQUESTED"
	ReviewStateCommented        ReviewState = "COMMENTED"
	ReviewStateDismissed        ReviewState = "DISMISSED"
)

// ReviewerKind distinguishes user reviewers from team reviewers.
type ReviewerKind string

const (
	ReviewerKindUser ReviewerKind = "user"
	ReviewerKindTeam ReviewerKind = "team"
)

// CheckState is our simplified view of a CI check's outcome.
type CheckState string

const (
	CheckStateSuccess CheckState = "SUCCESS"
	CheckStateFailure CheckState = "FAILURE"
	CheckStatePending CheckState = "PENDING"
	CheckStateNeutral CheckState = "NEUTRAL"
	CheckStateSkipped CheckState = "SKIPPED"
	CheckStateUnknown CheckState = "UNKNOWN"
)

// Check is a single CI check attached to a PR.
type Check struct {
	Name       string
	State      CheckState
	DetailsURL string
}

// RequiredReview describes a reviewer whose review is (or was) expected,
// along with their current state.
type RequiredReview struct {
	// Name is the display handle: "@login" for users, "@org/team" for teams.
	Name                string
	Kind                ReviewerKind
	RequiredByCodeOwner bool
	State               ReviewState
	// ApprovedByLogin is set when State=APPROVED and the review was made
	// by a user on behalf of a team (gives us "approved by @alice" context).
	ApprovedByLogin string
}

// PR is a pull request with all the data the dashboard needs. It is the
// stable boundary between the github package and the UI.
type PR struct {
	Repo           string
	Number         int
	Title          string
	URL            string
	HeadRef        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	IsDraft        bool
	ReviewDecision string // APPROVED / CHANGES_REQUESTED / REVIEW_REQUIRED / ""

	// Mergeable is GitHub's MergeableState enum: MERGEABLE,
	// CONFLICTING, or UNKNOWN. Tells us whether the HEAD would merge
	// cleanly into the base if we tried.
	Mergeable string
	// MergeStateStatus is GitHub's MergeStateStatus enum with more
	// context than Mergeable alone: CLEAN, DIRTY, BLOCKED, BEHIND,
	// DRAFT, HAS_HOOKS, UNSTABLE, UNKNOWN. See MergeBlockReasons for
	// how we translate this into human-readable blockers.
	MergeStateStatus string

	Checks          []Check
	FailingChecks   []Check
	PendingChecks   int
	TotalChecks     int
	UnresolvedCount int
	RequiredReviews []RequiredReview
}

// IsStale reports whether the PR has had no activity for more than
// StaleAfter. "Activity" here is whatever GitHub bumps updatedAt for —
// commits, comments, reviews, label changes, etc.
func (p PR) IsStale() bool {
	return !p.UpdatedAt.IsZero() && time.Since(p.UpdatedAt) > StaleAfter
}

// ApprovalCount returns how many requested reviewers have approved.
// This counts across every entry in RequiredReviews (both codeowner-
// required and ordinarily-requested reviewers).
func (p PR) ApprovalCount() int {
	n := 0
	for _, r := range p.RequiredReviews {
		if r.State == ReviewStateApproved {
			n++
		}
	}
	return n
}

// HasChangesRequested reports whether any requested reviewer requested
// changes. Like ApprovalCount, this spans every entry.
func (p PR) HasChangesRequested() bool {
	for _, r := range p.RequiredReviews {
		if r.State == ReviewStateChangesRequested {
			return true
		}
	}
	return false
}

// CodeOwnerReviews returns only the reviewers that GitHub considers
// "required" — i.e. those pulled in via CODEOWNERS (asCodeOwner=true on
// the underlying review request). GitHub doesn't expose branch-
// protection minimum-approvals count without admin scope, so CODEOWNERS
// is the strongest permission-free signal we have for "required".
func (p PR) CodeOwnerReviews() []RequiredReview {
	out := make([]RequiredReview, 0, len(p.RequiredReviews))
	for _, r := range p.RequiredReviews {
		if r.RequiredByCodeOwner {
			out = append(out, r)
		}
	}
	return out
}

// OptionalReviews returns the reviewers that were requested on the PR
// but are not CODEOWNERS-required (e.g. manually added reviewers or
// team requests that don't gate the merge).
func (p PR) OptionalReviews() []RequiredReview {
	out := make([]RequiredReview, 0, len(p.RequiredReviews))
	for _, r := range p.RequiredReviews {
		if !r.RequiredByCodeOwner {
			out = append(out, r)
		}
	}
	return out
}

// CodeOwnerApprovalCount reports how many CODEOWNERS-required reviewers
// have approved.
func (p PR) CodeOwnerApprovalCount() int {
	n := 0
	for _, r := range p.RequiredReviews {
		if r.RequiredByCodeOwner && r.State == ReviewStateApproved {
			n++
		}
	}
	return n
}

// CodeOwnerHasChangesRequested reports whether any CODEOWNERS-required
// reviewer has requested changes.
func (p PR) CodeOwnerHasChangesRequested() bool {
	for _, r := range p.RequiredReviews {
		if r.RequiredByCodeOwner && r.State == ReviewStateChangesRequested {
			return true
		}
	}
	return false
}

// MergeBlockReasons returns a short list of human-readable reasons the
// PR can't be merged right now, in rough order of severity. An empty
// slice means the PR is mergeable (or GitHub hasn't reported any
// blockers).
//
// Sources, in priority order:
//  1. mergeStateStatus — the most specific signal GitHub gives us
//     (DIRTY/BEHIND/DRAFT/BLOCKED/UNSTABLE/UNKNOWN/CLEAN/HAS_HOOKS).
//  2. For BLOCKED — which just means "branch protection is blocking
//     the merge" without saying why — we synthesise likely reasons
//     from observable signals: the PR's reviewDecision, CODEOWNERS
//     approvals, failing status checks, and pending checks.
//  3. As a fallback, a CONFLICTING value on mergeable is surfaced as
//     a conflict. Older GitHub instances sometimes populate mergeable
//     before mergeStateStatus.
func (p PR) MergeBlockReasons() []string {
	switch p.MergeStateStatus {
	case "CLEAN", "HAS_HOOKS":
		return nil
	case "DIRTY":
		return []string{"merge conflicts with base branch"}
	case "BEHIND":
		return []string{"branch is behind base"}
	case "DRAFT":
		return []string{"PR is a draft"}
	case "UNSTABLE":
		if len(p.FailingChecks) > 0 {
			return []string{fmt.Sprintf("%d failing check(s) (non-blocking)", len(p.FailingChecks))}
		}
		return []string{"non-required checks failing"}
	case "BLOCKED":
		return p.blockedReasons()
	case "UNKNOWN":
		return []string{"mergeability still being computed by GitHub"}
	}

	// mergeStateStatus was empty (older API, preview permissions, or
	// an unknown value we haven't enumerated). Fall back to whatever
	// the legacy mergeable field tells us.
	if p.Mergeable == "CONFLICTING" {
		return []string{"merge conflicts with base branch"}
	}
	return nil
}

// blockedReasons synthesises likely reasons the PR is in MergeState
// BLOCKED. GitHub's API doesn't expose which branch-protection rule
// is failing, so we list every observable condition that typically
// blocks a merge: changes requested, missing required approvals,
// failing checks, and pending checks.
func (p PR) blockedReasons() []string {
	var r []string
	switch p.ReviewDecision {
	case "CHANGES_REQUESTED":
		r = append(r, "changes requested by reviewer")
	case "REVIEW_REQUIRED":
		r = append(r, "missing required approvals")
	}
	if len(p.FailingChecks) > 0 {
		r = append(r, fmt.Sprintf("%d failing status check(s)", len(p.FailingChecks)))
	}
	if p.PendingChecks > 0 {
		r = append(r, fmt.Sprintf("%d check(s) still running", p.PendingChecks))
	}
	if len(r) == 0 {
		// GitHub marked this as blocked but none of our heuristics
		// hit — most likely a branch-protection rule we can't see
		// (e.g. required-status-checks list, signed commits).
		r = append(r, "blocked by branch protection")
	}
	return r
}
