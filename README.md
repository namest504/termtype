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

Pick a theme with `↑`/`↓` and a mode with `Tab`, then press `Enter`:

- **Normal** — type the sentence; WPM and accuracy are shown when you finish.
- **Time Attack** — race a 30s or 60s countdown and type as much as you can
  before time runs out.

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