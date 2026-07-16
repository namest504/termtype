# TermType

A simple typing practice application for your terminal.

![TermType — the claude theme](docs/termtype-claude.gif)

## Installation

### Homebrew

```bash
brew install namest504/termtype/termtype
```

## Usage

```bash
termtype
```

Pick a theme with `↑`/`↓`, a mode with `Tab`, and a language with `←`/`→`, then
press `Enter`:

- **Normal** — type the sentence; WPM and accuracy are shown when you finish.
- **Time Attack** — race a 30s or 60s countdown and type as much as you can
  before time runs out.
- **Language** — choose English or Korean (한국어). Korean uses an IME, and the
  themes lay out the wide Hangul glyphs correctly.

A live WPM/accuracy readout (and the countdown, in Time Attack) is shown in the
top-right corner while you type, on every theme.

### History & personal bests

Every finished round is saved to `~/.config/termtype/history.jsonl` (or your
platform's config directory). The result screen shows `NEW BEST!` when you set
a personal best, or your current best plus a sparkline of the last 10 rounds.
Bests are tracked separately per mode and language.

See a summary of your history without starting the game:

```bash
termtype --stats
```

### Any terminal, any width

TermType adapts to the terminal: the core themes reflow down to 20 columns, the
menu and overlays shrink to fit, and the layout reflows as you resize.

For terminals or fonts that can't render the Unicode symbols (boxes, the
spinner, ⏱/⏸), run with `--ascii` for a plain-ASCII rendering:

```bash
termtype --ascii
```

It is auto-enabled for non-UTF-8 locales, or you can force it with
`TERMTYPE_ASCII=1`.

### Controls

- `Ctrl-P` — pause / resume
- `Enter` — next sentence (after finishing)
- `Esc` / `Ctrl-C` — quit

## Available Themes

- `simple`: A simple, clean interface.
- `log`: A theme that simulates a log stream.
- `matrix`: A theme inspired by The Matrix.
- `hex`: A theme that mimics a hex editor.
- `diff`: A theme that looks like a git diff.
- `claude`: A theme that looks like composing a message in a Claude Code session.

All themes adapt to the terminal size and reflow as you resize the window.

### In action

The `matrix` and `log` themes:

![matrix theme](docs/termtype-matrix.gif)

![log theme](docs/termtype-log.gif)