# TermType

A simple typing practice application for your terminal.

## Build and Run

To build and run the application, you need to have Go installed.

```bash
go run ./cmd/termtype
```

## Usage

### Themes

You can choose a theme using the `-theme` flag. For example, to use the `matrix` theme, run:

```bash
go run ./cmd/termtype -theme=matrix
```

### List Themes

To see the list of available themes, use the `-list-themes` flag:

```bash
go run ./cmd/termtype -list-themes
```

## Available Themes

- `simple`: A simple, clean interface.
- `log`: A theme that simulates a log stream.
- `matrix`: A theme inspired by The Matrix.
- `hex`: A theme that mimics a hex editor.
- `diff`: A theme that looks like a git diff.
- `code`: A theme that resembles a code editor.
