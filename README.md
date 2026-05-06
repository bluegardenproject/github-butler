# github-butler

A neon-themed terminal dashboard that polls GitHub via the local `gh` CLI and shows your open PRs across multiple repositories at a glance — including CI status, required reviewers, unresolved review threads, draft/stale flags, and more.

## Features

- Live-refreshing TUI (Bubble Tea) with a neon truecolor palette and gradients
- Polls every N seconds (default 20s, configurable in-app)
- Tracks **only PRs you authored** across a configurable list of repositories
- Shows per-PR:
  - CI status with failing check names in the detail pane
  - Overall review decision (`REVIEW` column) plus a dedicated **`REQ`** column for CODEOWNERS-required reviewers only, so you can tell at a glance what's actually gating the merge
  - Details pane splits reviewers into **REQUIRED (CODEOWNERS)** and **OPTIONAL** sub-sections
  - **`MERGE`** column in the details pane explaining why a PR can't be merged (conflicts, behind base, draft, changes requested, missing required approvals, failing checks, pending checks, branch protection, …) with severity-appropriate colors
  - Unresolved review-thread count
  - `[DRAFT]` chip for draft PRs (still listed, not filtered out)
  - `[STALE]` chip for PRs created more than 4 weeks ago
- In-app menu for **managing repositories** (add/remove) and **settings** (poll interval)
- Config persisted as YAML; atomic saves so a crash can't corrupt the file
- Open a highlighted PR in your browser with one keystroke

## Requirements

- Go 1.25+ to build
- [`gh`](https://cli.github.com/) CLI installed and authenticated (`gh auth login`)

## Install

### From a release

Pre-built binaries for Linux, macOS, and Windows are published on the [Releases](https://github.com/bluegardenproject/github-butler/releases) page. Grab the asset for your platform, mark it executable, and drop it on `$PATH`:

```bash
curl -L -o github-butler https://github.com/bluegardenproject/github-butler/releases/latest/download/github-butler-darwin-arm64
chmod +x github-butler
mv github-butler /usr/local/bin/
```

(Replace `darwin-arm64` with `linux-amd64`, `linux-arm64`, `darwin-amd64`, or `windows-amd64.exe` as appropriate.)

### From source

Clone the repo and build a local binary:

```bash
git clone https://github.com/bluegardenproject/github-butler.git
cd github-butler
make build
```

Or install straight into `$GOBIN` (usually `~/go/bin`):

```bash
go install github.com/bluegardenproject/github-butler@latest
```

Verify with `github-butler --version`.

## Usage

Make sure `gh` is authenticated first (the app shells out to it for every GitHub call):

```bash
gh auth login
gh auth status
```

Then run the binary:

```bash
github-butler
```

### First launch

On the very first run there's no config file yet, so the dashboard opens empty. Add repos from inside the app:

1. Press `m` to open the main menu.
2. Select **Repositories** and press `Enter`.
3. Press `a`, paste a repo reference (`owner/repo`, an HTTPS URL, or an SSH URL), and press `Enter`.
4. Repeat for as many repos as you want to track, then press `Esc` to return to the dashboard.

The app validates each repo via `gh api repos/:owner/:repo` before saving, and writes the updated config to `~/.config/github-butler/config.yaml` atomically.

### Examples

Run with a custom config path:

```bash
github-butler --config ./my-config.yaml
```

Run without any color styling (useful for screenshots, CI logs, or terminals without truecolor support):

```bash
github-butler --no-color
```

`NO_COLOR=1` in the environment has the same effect.

### Quitting

Press `q` or `Ctrl+C` from any screen to exit cleanly.

## Config

Stored at `~/.config/github-butler/config.yaml` (override with `--config`). The app will create it on first save.

```yaml
repos:
  - owner/repo-a
  - owner/repo-b
poll_interval_seconds: 20
group_by_repo: false
```

### Available settings

| Key                     | Type | Default | Description                                                                                  |
| ----------------------- | ---- | ------- | -------------------------------------------------------------------------------------------- |
| `repos`                 | list | `[]`    | List of `owner/repo` slugs to track                                                          |
| `poll_interval_seconds` | int  | `20`    | How often the app polls GitHub (min `2`, max `3600`)                                         |
| `group_by_repo`         | bool | `false` | When `true`, PRs are grouped under per-repo section headers instead of one flat updated list |

All three are editable from inside the app (`m` → **Settings** / **Repositories**), so you rarely need to hand-edit the YAML — but doing so works too.

On first launch with no config, the app opens to an empty dashboard; press `m` → Repositories → `a` to add one.

## Key bindings

### Dashboard

| Key         | Action                      |
| ----------- | --------------------------- |
| `↑` / `k`   | Move selection up           |
| `↓` / `j`   | Move selection down         |
| `o`, Enter  | Open selected PR in browser |
| `r`         | Refresh now                 |
| `m`         | Open main menu              |
| `q`, Ctrl+C | Quit                        |

### Menu / Repositories / Settings

| Key       | Action                                                                    |
| --------- | ------------------------------------------------------------------------- |
| `↑` / `k` | Up                                                                        |
| `↓` / `j` | Down                                                                      |
| Enter     | Select                                                                    |
| `a`       | Add repository (Repositories view)                                        |
| `d`       | Delete highlighted repository                                             |
| `y` / `n` | Confirm / cancel delete                                                   |
| Esc       | Back one level (Repos/Settings → Menu, Menu → Dashboard)                  |
| `m`       | Jump back to the dashboard from anywhere (except while typing in a field) |

When adding a repo you can paste any of:

- `owner/repo`
- `https://github.com/owner/repo` (with or without `.git`)
- `git@github.com:owner/repo.git`

The app validates the repo exists (and you can access it) via `gh api repos/:owner/:repo` before saving.

## Flags

```
--config PATH     path to config file (default: ~/.config/github-butler/config.yaml)
--no-color        disable all color output (also respects the NO_COLOR env var)
--version         print version, build time, Go runtime, platform, and exit
```

## Project layout

```
main.go                    # entrypoint; declares Version / BuildTime ldflag targets
cmd/root.go                # flag parsing + wiring, including --version
internal/
  config/                  # YAML load/save/validate, defaults
  github/                  # gh CLI wrapper, GraphQL query, pure derivation logic
  ui/
    app.go                 # root Bubble Tea model + screen routing
    messages.go            # shared tea.Msg types
    commands.go            # tea.Cmd factories (fetch, save, open URL, validate)
    keys.go                # central key bindings
    helpers.go             # text padding / truncation
    dashboard.go           # dashboard screen + detail pane
    menu.go                # main menu
    repos.go               # repo manager + add/remove flows
    settings.go            # settings list + interval editor
    theme/                 # neon palette, styles, gradient helper
    components/            # small reusable widgets (banner, countdown, toast, confirm)
```

## Development

```bash
make setup            # wire up in-repo git hooks (run once after clone)
make build            # → ./github-butler
make test             # go test ./...
```

Commits must follow [Conventional Commits 1.0.0](https://www.conventionalcommits.org/en/v1.0.0/) and Go files must pass `gofmt`. Both rules are enforced locally by the hooks in [`.githooks/`](.githooks/) (wired up by `make setup`) and re-checked in CI.

### Conventional commit types

Allowed types: `build`, `chore`, `ci`, `docs`, `feat`, `fix`, `perf`, `refactor`, `revert`, `style`, `test`. Scope is optional and lower-case (e.g. `feat(ui): …`). Append `!` after the type/scope to mark a breaking change.

```
feat(ui): add per-PR merge-blocker column

fix(config): tolerate missing poll_interval_seconds

refactor(github)!: rename Client.List → Client.OpenForUser
```

## Releases

Releases are driven by [Release Please](https://github.com/googleapis/release-please) on `main`:

1. Every push to `main` runs the [release-please workflow](.github/workflows/release-please.yml). It opens (or updates) a release PR that aggregates all unreleased Conventional Commits into a `CHANGELOG.md` entry and bumps the version in `main.go` (via the `extra-files` entry in [`release-please-config.json`](release-please-config.json)) and `.release-please-manifest.json`.
2. Merging the release PR creates a Git tag (e.g. `v0.2.0`) and a GitHub Release. The same workflow then cross-compiles the five release binaries (`linux-amd64`, `linux-arm64`, `darwin-amd64`, `darwin-arm64`, `windows-amd64.exe`) with `-ldflags "-X main.Version=… -X main.BuildTime=…"` and uploads them as release assets.

Build artifacts can be reproduced locally with:

```bash
make build-all
```

The workflow uses a `PAT_TOKEN` repository secret with `repo` and `workflow` scope so release-please can push tags and create releases.

## License

MIT
