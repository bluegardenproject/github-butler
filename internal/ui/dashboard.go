package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bluegardenproject/github-butler/internal/config"
	"github.com/bluegardenproject/github-butler/internal/github"
	"github.com/bluegardenproject/github-butler/internal/ui/components"
	"github.com/bluegardenproject/github-butler/internal/ui/theme"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Column widths for the dashboard table. Kept here so header + rows use the
// same values.
const (
	colRepoW   = 22
	colNumW    = 6
	colTitleW  = 52
	colTagsW   = 15
	colBranchW = 22
	colCIW     = 12
	colRevW    = 14
	colReqW    = 7
	colUnresW  = 6
	colAgeW    = 8
	colActW    = 8

	compactAutoWidthThreshold = 150
	compactMinContentWidth    = 60
)

func (m Model) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(km, keys.Quit):
		return m, tea.Quit
	case key.Matches(km, keys.Menu):
		m.screen = screenMenu
		m.menuCursor = 0
		return m, nil
	case key.Matches(km, keys.Refresh):
		if !m.loading {
			m.loading = true
			return m, fetchPRsCmd(m.client, m.cfg.Repos)
		}
	case key.Matches(km, keys.Up):
		if m.selected > 0 {
			m.selected--
		}
	case key.Matches(km, keys.Down):
		if m.selected < len(m.prs)-1 {
			m.selected++
		}
	case key.Matches(km, keys.OpenPR), key.Matches(km, keys.Select):
		if len(m.prs) > 0 && m.selected >= 0 && m.selected < len(m.prs) {
			return m, openURLCmd(m.prs[m.selected].URL)
		}
	}
	return m, nil
}

func (m Model) viewDashboard() string {
	bigBanner := components.BigBanner("GITHUB-BUTLER")
	smallBanner := components.Banner(" PR DASHBOARD ")
	subtitle := theme.Dimmed.Render(fmt.Sprintf("%d PR(s) across %d repo(s)", len(m.prs), len(m.cfg.Repos)))

	var body string
	switch {
	case len(m.cfg.Repos) == 0:
		body = theme.Panel.Render(theme.Pending.Render("No repositories configured.") +
			"\n\n" +
			theme.KeyLabel.Render("Press ") +
			theme.KeyHint.Render("m") +
			theme.KeyLabel.Render(" to open the menu and add one."))
	case len(m.prs) == 0 && m.loading:
		body = theme.Panel.Render(components.Loading("Fetching PRs..."))
	case len(m.prs) == 0:
		body = theme.Panel.Render(theme.Dimmed.Render("No open PRs authored by you in the configured repos."))
	default:
		body = m.renderTable()
	}

	detail := m.renderDetail()
	footer := m.renderFooter()

	// Once we have data, subsequent fetches don't replace the table —
	// they happen in the background. Surface them with a small inline
	// spinner next to the subtitle so the user sees something is
	// happening even though the existing rows stay visible.
	subHeaderParts := []string{smallBanner, "  ", subtitle}
	if m.loading && len(m.prs) > 0 {
		subHeaderParts = append(subHeaderParts, "  ", components.Loading("refreshing"))
	}
	subHeader := lipgloss.JoinHorizontal(lipgloss.Bottom, subHeaderParts...)
	return lipgloss.JoinVertical(lipgloss.Left, bigBanner, "", subHeader, body, detail, footer)
}

func (m Model) renderTable() string {
	if m.useCompactTable() {
		return m.renderCompactTable()
	}

	header := renderHeaderRow()
	rows := make([]string, 0, len(m.prs)+1)
	rows = append(rows, header)

	var prevRepo string
	for i, pr := range m.prs {
		if m.cfg.GroupByRepo && pr.Repo != prevRepo {
			rows = append(rows, renderGroupHeader(pr.Repo))
			prevRepo = pr.Repo
		}
		rows = append(rows, m.renderRow(pr, i == m.selected))
	}
	return theme.Panel.Render(strings.Join(rows, "\n"))
}

func (m Model) useCompactTable() bool {
	switch m.cfg.DashboardView {
	case config.DashboardViewCompact:
		return true
	case config.DashboardViewFull:
		return false
	default:
		return m.width > 0 && m.width < compactAutoWidthThreshold
	}
}

func (m Model) renderCompactTable() string {
	contentWidth := m.tableContentWidth()
	rows := make([]string, 0, len(m.prs)+1)
	rows = append(rows, renderCompactHeaderRow(contentWidth))

	var prevRepo string
	for i, pr := range m.prs {
		if m.cfg.GroupByRepo && pr.Repo != prevRepo {
			rows = append(rows, renderGroupHeader(pr.Repo))
			prevRepo = pr.Repo
		}
		rows = append(rows, m.renderCompactRow(pr, i == m.selected, contentWidth))
	}
	return theme.Panel.Render(strings.Join(rows, "\n"))
}

func (m Model) tableContentWidth() int {
	// Panel adds a rounded border and horizontal padding, so keep the inner
	// layout a few columns narrower than the terminal width.
	if m.width <= 0 {
		return 100
	}
	w := m.width - 4
	if w < compactMinContentWidth {
		return compactMinContentWidth
	}
	return w
}

// renderGroupHeader renders a single separator row above each repo group
// when group-by-repo is enabled. Kept visually light so it doesn't
// compete with the gradient table header.
func renderGroupHeader(repo string) string {
	label := theme.Gradient("▸ "+repo, theme.NeonPink, theme.NeonCyan)
	return label
}

// OrderPRs returns a copy of prs ordered for display on the dashboard.
//
// When groupByRepo is false the slice is sorted "most recently updated
// first" — the canonical list ordering used everywhere else.
//
// When groupByRepo is true PRs belonging to the same repo sit next to
// each other; the relative order within a group remains "most recently
// updated first". Repo groups themselves are ordered to match the
// user's configured repo list, with any repos not in that list
// appended alphabetically at the end.
//
// Toggling the flag re-applies the correct order to the existing slice
// so the UI updates immediately, without waiting for the next fetch.
func OrderPRs(prs []github.PR, groupByRepo bool, repoOrder []string) []github.PR {
	out := make([]github.PR, len(prs))
	copy(out, prs)
	if !groupByRepo {
		sort.SliceStable(out, func(i, j int) bool {
			return out[i].UpdatedAt.After(out[j].UpdatedAt)
		})
		return out
	}
	rank := make(map[string]int, len(repoOrder))
	for i, r := range repoOrder {
		rank[r] = i
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Repo != out[j].Repo {
			ri, oki := rank[out[i].Repo]
			rj, okj := rank[out[j].Repo]
			switch {
			case oki && okj:
				return ri < rj
			case oki:
				return true
			case okj:
				return false
			default:
				return out[i].Repo < out[j].Repo
			}
		}
		return out[i].UpdatedAt.After(out[j].UpdatedAt)
	})
	return out
}

func renderHeaderRow() string {
	cells := []string{
		pad("REPO", colRepoW),
		pad("#", colNumW),
		pad("TITLE", colTitleW),
		pad("TAGS", colTagsW),
		pad("BRANCH", colBranchW),
		pad("CI", colCIW),
		pad("REVIEW", colRevW),
		pad("REQ", colReqW),
		pad("UNRES", colUnresW),
		pad("AGE", colAgeW),
		pad("ACT", colActW),
	}
	return theme.Gradient(strings.Join(cells, " "), theme.HeaderStops...)
}

func renderCompactHeaderRow(width int) string {
	repoW, numW, titleW, tagsW, statusW := compactColumnWidths(width)
	cells := []string{
		pad("REPO", repoW),
		pad("#", numW),
		pad("TITLE", titleW),
		pad("TAGS", tagsW),
		pad("STATUS", statusW),
	}
	return theme.Gradient(strings.Join(cells, " "), theme.HeaderStops...)
}

func (m Model) renderRow(pr github.PR, selected bool) string {
	repo := pad(shortRepo(pr.Repo), colRepoW)
	num := pad(fmt.Sprintf("#%d", pr.Number), colNumW)
	title := pad(truncate(pr.Title, colTitleW), colTitleW)

	tags := padVisible(renderTags(pr), colTagsW)
	branch := pad(truncate(pr.HeadRef, colBranchW), colBranchW)
	ci := padVisible(renderCIStatus(pr), colCIW)
	rev := padVisible(renderReviewStatus(pr), colRevW)
	req := padVisible(renderRequiredStatus(pr), colReqW)
	unres := padVisible(renderUnresolved(pr.UnresolvedCount), colUnresW)
	age := pad(relativeTime(pr.CreatedAt), colAgeW)
	act := pad(relativeTime(pr.UpdatedAt), colActW)

	// Only the first three columns participate in the selection highlight.
	// The columns to the right contain styled spans (chips with their own
	// background, colored CI/review states, etc.) whose ANSI resets would
	// otherwise truncate the selected-row background inconsistently
	// depending on which chips happen to be present.
	leading := strings.Join([]string{repo, num, title}, " ")
	if selected {
		leading = theme.SelectedRow.Render(leading)
	}
	trailing := strings.Join([]string{tags, branch, ci, rev, req, unres, age, act}, " ")
	return leading + " " + trailing
}

func (m Model) renderCompactRow(pr github.PR, selected bool, width int) string {
	repoW, numW, titleW, tagsW, statusW := compactColumnWidths(width)
	repo := pad(shortRepo(pr.Repo), repoW)
	num := pad(fmt.Sprintf("#%d", pr.Number), numW)
	title := pad(truncate(pr.Title, titleW), titleW)
	tags := padVisible(renderTags(pr), tagsW)
	status := padVisible(compactStatus(pr), statusW)

	leading := strings.Join([]string{repo, num, title}, " ")
	if selected {
		leading = theme.SelectedRow.Render(leading)
	}
	return strings.Join([]string{leading, tags, status}, " ")
}

func compactColumnWidths(width int) (repoW, numW, titleW, tagsW, statusW int) {
	repoW, numW, tagsW, statusW = 18, 7, 12, 24
	titleW = width - repoW - numW - tagsW - statusW - 4
	if titleW >= 18 {
		return repoW, numW, titleW, tagsW, statusW
	}

	for titleW < 18 && statusW > 18 {
		statusW--
		titleW++
	}
	for titleW < 18 && tagsW > 8 {
		tagsW--
		titleW++
	}
	for titleW < 18 && repoW > 12 {
		repoW--
		titleW++
	}
	if titleW < 12 {
		titleW = 12
	}
	return repoW, numW, titleW, tagsW, statusW
}

func compactStatus(pr github.PR) string {
	parts := []string{compactCIStatus(pr), compactReviewStatus(pr)}
	if pr.UnresolvedCount > 0 {
		parts = append(parts, theme.Pending.Render(fmt.Sprintf("UNRES %d", pr.UnresolvedCount)))
	}
	return strings.Join(parts, " ")
}

func compactCIStatus(pr github.PR) string {
	switch {
	case pr.TotalChecks == 0:
		return theme.Dimmed.Render("—")
	case len(pr.FailingChecks) > 0:
		return theme.Fail.Render(fmt.Sprintf("FAIL×%d", len(pr.FailingChecks)))
	case pr.PendingChecks > 0:
		return theme.Pending.Render(fmt.Sprintf("RUN×%d", pr.PendingChecks))
	default:
		return theme.OK.Render("PASS")
	}
}

func compactReviewStatus(pr github.PR) string {
	total := len(pr.CodeOwnerReviews())
	if total > 0 {
		approved := pr.CodeOwnerApprovalCount()
		switch {
		case pr.CodeOwnerHasChangesRequested():
			return theme.Fail.Render("CHG")
		case approved == total:
			return theme.OK.Render("OK")
		default:
			return theme.Info.Render(fmt.Sprintf("REQ %d/%d", approved, total))
		}
	}

	switch pr.ReviewDecision {
	case "APPROVED":
		return theme.OK.Render("OK")
	case "CHANGES_REQUESTED":
		return theme.Fail.Render("CHG")
	case "REVIEW_REQUIRED":
		return theme.Info.Render("PEND")
	default:
		return theme.Dimmed.Render("—")
	}
}

// renderTags joins the DRAFT/STALE chips for a PR into a single styled
// string. Empty when the PR has neither flag; the row renderer pads the
// result out to colTagsW so every column to the right stays aligned.
func renderTags(pr github.PR) string {
	var chips []string
	if pr.IsDraft {
		chips = append(chips, theme.DraftChip.Render("DRAFT"))
	}
	if pr.IsStale() {
		chips = append(chips, theme.StaleChip.Render("STALE"))
	}
	return strings.Join(chips, " ")
}

func shortRepo(full string) string {
	if i := strings.Index(full, "/"); i >= 0 {
		return full[i+1:]
	}
	return full
}

func renderCIStatus(pr github.PR) string {
	switch {
	case pr.TotalChecks == 0:
		return theme.Dimmed.Render("—")
	case len(pr.FailingChecks) > 0:
		return theme.Fail.Render(fmt.Sprintf("FAIL ×%d", len(pr.FailingChecks)))
	case pr.PendingChecks > 0:
		return theme.Pending.Render(fmt.Sprintf("RUNNING ×%d", pr.PendingChecks))
	default:
		return theme.OK.Render("PASS")
	}
}

func renderReviewStatus(pr github.PR) string {
	required := len(pr.RequiredReviews)
	approved := pr.ApprovalCount()

	var base string
	switch pr.ReviewDecision {
	case "APPROVED":
		base = theme.OK.Render("APPROVED")
	case "CHANGES_REQUESTED":
		base = theme.Fail.Render("CHANGES")
	case "REVIEW_REQUIRED":
		base = theme.Info.Render("REQUIRED")
	default:
		if required > 0 {
			base = theme.Info.Render("PENDING")
		} else {
			base = theme.Dimmed.Render("—")
		}
	}
	if required > 0 {
		ratio := lipgloss.NewStyle().Foreground(theme.NeonYellow).Render(fmt.Sprintf(" %d/%d", approved, required))
		base += ratio
	}
	if pr.HasChangesRequested() && pr.ReviewDecision != "CHANGES_REQUESTED" {
		base += theme.Fail.Render("!")
	}
	return base
}

// renderRequiredStatus renders the REQ column: approved / total among
// CODEOWNERS-required reviewers. GitHub doesn't expose branch-protection
// rules without admin scope, so CODEOWNERS is the best permission-free
// stand-in for "required to merge".
//
// Color rules mirror the existing CI/Review conventions:
//   - no required reviewers   → dim "—"
//   - any changes requested   → red "✗ a/t"
//   - all approved (a == t)   → green "✓ a/t"
//   - otherwise (pending)     → cyan "a/t"
func renderRequiredStatus(pr github.PR) string {
	total := len(pr.CodeOwnerReviews())
	if total == 0 {
		return theme.Dimmed.Render("—")
	}
	approved := pr.CodeOwnerApprovalCount()
	ratio := fmt.Sprintf("%d/%d", approved, total)
	switch {
	case pr.CodeOwnerHasChangesRequested():
		return theme.Fail.Render("✗ " + ratio)
	case approved == total:
		return theme.OK.Render("✓ " + ratio)
	default:
		return theme.Info.Render(ratio)
	}
}

func renderUnresolved(n int) string {
	if n == 0 {
		return theme.Dimmed.Render("0")
	}
	return lipgloss.NewStyle().Foreground(theme.NeonOrange).Bold(true).Render(fmt.Sprintf("%d", n))
}

func relativeTime(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	default:
		return fmt.Sprintf("%dw", int(d.Hours()/(24*7)))
	}
}

func (m Model) renderDetail() string {
	if len(m.prs) == 0 || m.selected < 0 || m.selected >= len(m.prs) {
		return ""
	}
	pr := m.prs[m.selected]

	title := theme.PanelTitle.Render(
		theme.Gradient("■ Details ", theme.NeonPink, theme.NeonCyan) +
			theme.Dimmed.Render(pr.URL),
	)

	checks := renderChecksSection(pr)
	reviews := renderReviewsSection(pr)
	merge := renderMergeSection(pr)

	gap := lipgloss.NewStyle().Width(2).Render(" ")
	bottomRow := lipgloss.JoinHorizontal(lipgloss.Top,
		reviews,
		gap,
		merge,
	)
	body := lipgloss.JoinVertical(lipgloss.Left, checks, "", bottomRow)
	return theme.Panel.Render(lipgloss.JoinVertical(lipgloss.Left, title, body))
}

// renderMergeSection renders the rightmost "MERGE" column in the detail
// pane. When the PR is mergeable it shows a green "ready to merge";
// otherwise it lists every blocker returned by PR.MergeBlockReasons()
// one per line, with a severity-appropriate color.
func renderMergeSection(pr github.PR) string {
	header := theme.Gradient("MERGE", theme.NeonOrange, theme.NeonPink)
	reasons := pr.MergeBlockReasons()
	if len(reasons) == 0 {
		state := theme.OK.Render("  ✓ ready to merge")
		return lipgloss.JoinVertical(lipgloss.Left, header, state)
	}
	style := mergeReasonStyle(pr.MergeStateStatus)
	lines := []string{header}
	for _, r := range reasons {
		lines = append(lines, style.Render("  ✗ "+r))
	}
	if pr.MergeStateStatus != "" {
		lines = append(lines, theme.Dimmed.Render("  ("+strings.ToLower(pr.MergeStateStatus)+")"))
	}
	return strings.Join(lines, "\n")
}

// mergeReasonStyle picks the color for blocker lines based on the
// severity of the underlying mergeStateStatus. Hard blockers
// (conflicts / changes requested / behind) render red; soft blockers
// (pending checks, unknown) render cyan so they don't scream.
func mergeReasonStyle(status string) lipgloss.Style {
	switch status {
	case "DIRTY", "BEHIND", "BLOCKED":
		return theme.Fail
	case "DRAFT", "UNSTABLE":
		return theme.Pending
	case "UNKNOWN", "":
		return theme.Info
	default:
		return theme.Fail
	}
}

func renderChecksSection(pr github.PR) string {
	header := theme.Gradient("CHECKS", theme.NeonPink, theme.NeonMagenta)
	if pr.TotalChecks == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, theme.Dimmed.Render("  none"))
	}
	lines := []string{header}
	if len(pr.FailingChecks) == 0 {
		lines = append(lines, theme.OK.Render("  ✓ all passing"))
	} else {
		for _, c := range pr.FailingChecks {
			line := theme.Fail.Render("  ✗ " + c.Name)
			if c.DetailsURL != "" {
				line += theme.Dimmed.Render("  " + c.DetailsURL)
			}
			lines = append(lines, line)
		}
	}
	if pr.PendingChecks > 0 {
		lines = append(lines, theme.Pending.Render(fmt.Sprintf("  … %d running", pr.PendingChecks)))
	}
	return strings.Join(lines, "\n")
}

func renderReviewsSection(pr github.PR) string {
	header := theme.Gradient("REVIEWS", theme.NeonCyan, theme.NeonPurple)
	if len(pr.RequiredReviews) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, header, theme.Dimmed.Render("  none"))
	}

	required := pr.CodeOwnerReviews()
	optional := pr.OptionalReviews()

	lines := []string{header}
	lines = append(lines, renderReviewGroup("REQUIRED (CODEOWNERS)", required)...)
	if len(optional) > 0 {
		lines = append(lines, "")
		lines = append(lines, renderReviewGroup("OPTIONAL", optional)...)
	}
	return strings.Join(lines, "\n")
}

// renderReviewGroup renders a subsection of the reviews pane with its
// own subtitle and one line per reviewer. Returns "(none)" when the
// group is empty so the subtitle stays meaningful.
func renderReviewGroup(title string, reviews []github.RequiredReview) []string {
	sub := theme.Dimmed.Render("  " + title)
	if len(reviews) == 0 {
		return []string{sub, theme.Dimmed.Render("    (none)")}
	}
	lines := []string{sub}
	for _, r := range reviews {
		lines = append(lines, renderReviewLine(r))
	}
	return lines
}

// renderReviewLine formats a single reviewer row: status marker,
// handle, and optional "via @login" attribution for approvals made on
// behalf of a team.
func renderReviewLine(r github.RequiredReview) string {
	var marker string
	switch r.State {
	case github.ReviewStateApproved:
		marker = theme.OK.Render("    ✓")
	case github.ReviewStateChangesRequested:
		marker = theme.Fail.Render("    ✗")
	case github.ReviewStateCommented:
		marker = theme.Info.Render("    ●")
	default:
		marker = theme.Pending.Render("    …")
	}
	name := lipgloss.NewStyle().Bold(true).Render(r.Name)
	tail := ""
	if r.ApprovedByLogin != "" && r.State == github.ReviewStateApproved {
		tail = theme.Dimmed.Render("  via @" + r.ApprovedByLogin)
	}
	return fmt.Sprintf("%s %s%s", marker, name, tail)
}

func (m Model) renderFooter() string {
	var left string
	if m.lastFetched.IsZero() {
		left = theme.Dimmed.Render("never refreshed")
	} else {
		left = theme.Dimmed.Render("last: " + relativeTime(m.lastFetched))
	}

	remaining := time.Until(m.nextTick).Seconds()
	total := m.cfg.PollInterval().Seconds()
	bar := components.Countdown(remaining, total, 20)
	next := theme.Dimmed.Render(fmt.Sprintf(" next: %ds", int(remaining)))

	hints := footerHints(
		keys.Refresh, keys.Menu, keys.OpenPR, keys.Up, keys.Down, keys.Quit,
	)

	top := lipgloss.JoinHorizontal(lipgloss.Left, left, "   ", bar, next)
	return lipgloss.JoinVertical(lipgloss.Left, top, hints)
}
