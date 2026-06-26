# CLAUDE.md

Guidance for working in this repository.

## Overview

`termtype` is a terminal typing-practice game written in Go using
[`tcell`](https://github.com/gdamore/tcell). The user picks a visual theme from
an interactive menu, then types a randomly chosen sentence while WPM and
accuracy are measured. Every theme renders the same typing session with a
different look (plain, log stream, Matrix rain, hex editor, git diff, Claude
Code chat).

## Common commands

```bash
go build ./...      # build everything
go vet ./...        # static checks
gofmt -l .          # list unformatted files (must be empty; CI enforces it)
go test ./...       # run all tests
go run ./cmd/termtype   # run the app in your terminal
go run ./cmd/termtype --version
```

CI (`.github/workflows/ci.yml`) runs gofmt, build, vet, and test on every push
and PR. Keep all four green.

## Architecture

The code follows a small domain-centered layering. Dependencies point inward
toward `internal/domain`.

- `cmd/termtype/main.go` — entry point: screen init, theme-select menu,
  `--version`, then hands off to the game loop.
- `internal/domain/` — core abstractions with no UI dependencies:
  - `theme.go` — the `Theme` and `Renderer` interfaces.
  - `game_state.go` — `GameState` (target, input, timer, WPM, accuracy, and a
    per-theme `CustomState any`).
  - `sentences.go` / `sentences_ko.go` — the English and Korean sentence
    pools. The active pool lives on `GameState.Sentences`; themes pick from it
    via `gs.RandomSentence()` (never reference the package-level `Sentences`
    directly, or the language toggle is ignored).
- `internal/app/game.go` — the game loop: key handling, a 1s ticker for
  animations, completion detection, and WPM/accuracy computation. **All
  length math here is rune-based, not byte-based.**
- `internal/ui/` — the concrete `Renderer` plus shared drawing helpers:
  - `typing_renderer.go` — `TypingRenderer` draws the wrapped target, colors
    each character by correctness, and positions the cursor.
  - `util.go` — `WrapText` (greedy word wrap) and `Truncate`.
  - `result.go` — `ResultText` (the shared "WPM: .. | Accuracy: ..%" string).
- `internal/themes/` — one file per theme; each registers itself in
  `registry.go`'s `Themes` map via `init()`.

## The Theme interface

```go
type Theme interface {
    ResetState(*GameState)               // start a new round
    UpdateScreen(Renderer, *GameState)   // draw the current state
    OnTick(*GameState)                   // advance animation (called ~1/s)
}
```

### Adding a theme

1. Create `internal/themes/<name>_theme.go`.
2. Register it: `func init() { Themes["<name>"] = &MyTheme{} }`.
3. In `ResetState`, call `gs.ResetCommon()`, pick a sentence with
   `gs.RandomSentence()` (honors the selected language), and initialize any
   `gs.CustomState`.
4. In `UpdateScreen`, `renderer.Clear()`, draw, then `renderer.Show()`. Reuse
   `ui.TypingRenderer` for the typing area and `ui.ResultText` for the result.
5. **Be responsive:** read `renderer.Size()` and guard against small terminals.
   `responsive_test.go` renders every theme from 1x1 to 200x60 and must not
   panic — keep it passing.

## Release & distribution

Tagging `vX.Y.Z` triggers `.github/workflows/release.yml` → GoReleaser:

- Builds linux/darwin × amd64/arm64 binaries and publishes a GitHub Release.
- Generates the Homebrew formula and pushes it to the `namest504/homebrew-termtype`
  tap (binary install, no Go required on the user's machine).

Requirements:
- The `HOMEBREW_TAP_GITHUB_TOKEN` secret (a PAT with Contents:write on the tap
  repo) must exist on this repo for the cross-repo push.
- The release workflow pins `goreleaser-action@v6` with `version: '~> v2'`
  because the config is GoReleaser v2 (`version: 2` in `.goreleaser.yml`).

Install for end users:

```bash
brew install namest504/termtype/termtype
```

Note: GoReleaser marks `brews` (formula) as deprecated in favor of
`homebrew_casks`, but casks are macOS-only; `brews` is kept to preserve Linux
Homebrew support.

## Conventions & gotchas

- **Comments and identifiers are English.** Run `gofmt -w` before committing.
- **Commits** follow Conventional Commits (`fix:`, `feat:`, `ci:`, `docs:`,
  `chore:`) with a capitalized summary.
- **Runes, not bytes:** never use `len(string)` for character counts in game or
  rendering logic — multibyte input breaks it. Use `[]rune`.
- **Display width, not rune count, for positioning:** Hangul (and other CJK)
  syllables are two cells wide. When placing glyphs or the cursor, advance by
  `runewidth.RuneWidth`/`StringWidth`, never by rune index. `TypingRenderer`
  already does this; bespoke drawing (e.g. the `diff` theme) must too. The
  Korean pool exercises these paths in `responsive_test.go`.
- `WrapText` keeps the inter-word space as a trailing space on wrapped lines so
  that the sum of wrapped runes equals the original; `TypingRenderer` relies on
  this for cursor/coloring alignment. There is a regression test for it.
- The game refuses to play below 40 columns (see `game.go`); themes should
  still avoid panicking at any size.
