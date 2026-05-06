package github

import "sort"

// toPR converts a raw GraphQL node into the PR type exposed to the rest
// of the app. All derivation (failing checks, unresolved count, required
// reviews merge) lives here as pure logic so it can be unit-tested
// without hitting the network.
func toPR(n rawNode) PR {
	p := PR{
		Repo:             n.Repository.NameWithOwner,
		Number:           n.Number,
		Title:            n.Title,
		URL:              n.URL,
		HeadRef:          n.HeadRefName,
		CreatedAt:        n.CreatedAt,
		UpdatedAt:        n.UpdatedAt,
		IsDraft:          n.IsDraft,
		ReviewDecision:   n.ReviewDecision,
		Mergeable:        n.Mergeable,
		MergeStateStatus: n.MergeStateStatus,
	}

	p.Checks = extractChecks(n)
	for _, c := range p.Checks {
		switch c.State {
		case CheckStateFailure:
			p.FailingChecks = append(p.FailingChecks, c)
		case CheckStatePending:
			p.PendingChecks++
		}
	}
	p.TotalChecks = len(p.Checks)

	for _, t := range n.ReviewThreads.Nodes {
		if !t.IsResolved {
			p.UnresolvedCount++
		}
	}

	p.RequiredReviews = mergeRequiredReviews(n.ReviewRequests.Nodes, n.LatestReviews.Nodes)
	return p
}

// extractChecks normalizes CheckRun and StatusContext nodes into a single
// []Check slice.
func extractChecks(n rawNode) []Check {
	if len(n.Commits.Nodes) == 0 || n.Commits.Nodes[0].Commit.StatusCheckRollup == nil {
		return nil
	}
	ctxs := n.Commits.Nodes[0].Commit.StatusCheckRollup.Contexts.Nodes
	out := make([]Check, 0, len(ctxs))
	for _, c := range ctxs {
		switch c.Typename {
		case "CheckRun":
			out = append(out, Check{
				Name:       c.Name,
				State:      mapCheckRunState(c.Status, c.Conclusion),
				DetailsURL: c.DetailsURL,
			})
		case "StatusContext":
			out = append(out, Check{
				Name:       c.Context,
				State:      mapStatusContextState(c.State),
				DetailsURL: c.TargetURL,
			})
		}
	}
	return out
}

// mapCheckRunState converts GitHub's CheckRun status+conclusion pair into
// our simplified CheckState. A CheckRun is either in-progress (status !=
// COMPLETED) or completed with a conclusion.
func mapCheckRunState(status, conclusion string) CheckState {
	if status != "COMPLETED" && status != "" {
		return CheckStatePending
	}
	switch conclusion {
	case "SUCCESS":
		return CheckStateSuccess
	case "FAILURE", "TIMED_OUT", "CANCELLED", "ACTION_REQUIRED", "STARTUP_FAILURE":
		return CheckStateFailure
	case "NEUTRAL":
		return CheckStateNeutral
	case "SKIPPED":
		return CheckStateSkipped
	case "":
		return CheckStatePending
	default:
		return CheckStateUnknown
	}
}

// mapStatusContextState converts a legacy commit-status state ("SUCCESS",
// "ERROR", "FAILURE", "PENDING", "EXPECTED") into our CheckState.
func mapStatusContextState(state string) CheckState {
	switch state {
	case "SUCCESS":
		return CheckStateSuccess
	case "ERROR", "FAILURE":
		return CheckStateFailure
	case "PENDING", "EXPECTED":
		return CheckStatePending
	default:
		return CheckStateUnknown
	}
}

// mergeRequiredReviews combines still-pending review requests with the
// latest-per-reviewer review states. See plan "Required reviews derivation"
// for the rules this implements.
func mergeRequiredReviews(requests []rawReviewRequest, reviews []rawLatestReview) []RequiredReview {
	byKey := make(map[string]*RequiredReview)

	addPending := func(key, name string, kind ReviewerKind, codeOwner bool) {
		if key == "" {
			return
		}
		if existing, ok := byKey[key]; ok {
			if codeOwner {
				existing.RequiredByCodeOwner = true
			}
			return
		}
		byKey[key] = &RequiredReview{
			Name:                name,
			Kind:                kind,
			RequiredByCodeOwner: codeOwner,
			State:               ReviewStatePending,
		}
	}

	for _, r := range requests {
		switch r.RequestedReviewer.Typename {
		case "User":
			login := r.RequestedReviewer.Login
			if login == "" {
				continue
			}
			addPending("user:"+login, "@"+login, ReviewerKindUser, r.AsCodeOwner)
		case "Team":
			slug := r.RequestedReviewer.CombinedSlug
			if slug == "" {
				continue
			}
			addPending("team:"+slug, "@"+slug, ReviewerKindTeam, r.AsCodeOwner)
		}
	}

	upsertReview := func(key, name string, kind ReviewerKind, state ReviewState, approvedBy string) {
		if key == "" {
			return
		}
		if existing, ok := byKey[key]; ok {
			if stateRank(state) > stateRank(existing.State) {
				existing.State = state
				if approvedBy != "" && existing.ApprovedByLogin == "" {
					existing.ApprovedByLogin = approvedBy
				}
			}
			return
		}
		byKey[key] = &RequiredReview{
			Name:            name,
			Kind:            kind,
			State:           state,
			ApprovedByLogin: approvedBy,
		}
	}

	for _, rv := range reviews {
		login := rv.Author.Login
		state := reviewStateFromString(rv.State)
		if state == "" || login == "" {
			continue
		}
		upsertReview("user:"+login, "@"+login, ReviewerKindUser, state, "")

		for _, team := range rv.OnBehalfOf.Nodes {
			slug := team.CombinedSlug
			if slug == "" {
				continue
			}
			upsertReview("team:"+slug, "@"+slug, ReviewerKindTeam, state, login)
		}
	}

	out := make([]RequiredReview, 0, len(byKey))
	for _, v := range byKey {
		out = append(out, *v)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].RequiredByCodeOwner != out[j].RequiredByCodeOwner {
			return out[i].RequiredByCodeOwner
		}
		if out[i].Kind != out[j].Kind {
			return out[i].Kind == ReviewerKindTeam
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func reviewStateFromString(s string) ReviewState {
	switch s {
	case "APPROVED":
		return ReviewStateApproved
	case "CHANGES_REQUESTED":
		return ReviewStateChangesRequested
	case "COMMENTED":
		return ReviewStateCommented
	case "DISMISSED":
		return ReviewStateDismissed
	case "PENDING":
		return ReviewStatePending
	}
	return ""
}

// stateRank defines which review state "wins" when a reviewer appears in
// multiple places. Approved / changes-requested are authoritative; a later
// COMMENTED does not overwrite them.
func stateRank(s ReviewState) int {
	switch s {
	case ReviewStateChangesRequested:
		return 4
	case ReviewStateApproved:
		return 3
	case ReviewStateCommented:
		return 2
	case ReviewStateDismissed:
		return 1
	case ReviewStatePending:
		return 0
	}
	return -1
}
