# TermType Feature Expansion — Design Spec

- **Date:** 2026-07-16
- **Status:** Approved (pending spec review)
- **Scope:** Four features shipped as four independent PRs, merged to `main` in order.

## Overview

TermType currently plays a round and throws the result away. This design adds a
persistence layer and four features on top of it:

1. **History & personal bests** — save every finished round, show progress.
2. **Quick start** — CLI flags to skip the menu; the menu remembers the last selection.
3. **Custom text practice** — practice any file or piped text, including code.
4. **Pace ghost** — an opt-in ghost cursor racing at your personal-best speed.

Non-goals (explicitly out of scope): keystroke-replay ghosts, networked
multiplayer, daily challenges, per-key weakness analysis. These may build on
this foundation later.

## Common foundation: `internal/store` (new package)

All persistence lives in one package with no new external dependencies.

- **Location:** `os.UserConfigDir()/termtype/`
  (Linux `~/.config/termtype/`, macOS `~/Library/Application Support/termtype/`,
  Windows `%AppData%\termtype\`).
- **`config.json`** — last selections and toggles:

  ```json
  {"theme": "matrix", "mode": "normal", "lang": "en", "ghost": false}
  ```

- **`history.jsonl`** — one line per finished round, append-only:

  ```json
  {"ts":"2026-07-16T12:34:56Z","theme":"matrix","mode":"normal","lang":"en","wpm":72.4,"acc":98.1,"dur_s":14.2,"source":"builtin"}
  ```

  - `mode`: `"normal" | "ta30" | "ta60"`
  - `lang`: `"en" | "ko"`, or `"-"` for custom-text runs
  - `source`: `"builtin" | "custom"`

- **Error policy:** storage failures (unwritable dir, corrupt line, missing
  file) are silently ignored. A round is never interrupted, and unparseable
  history lines are skipped, not fatal.
- History is read once at startup and cached in memory; each finished round
  appends one line.

## Feature 1 — History & personal bests (PR 1)

**Recording unit:** Normal mode records one line per finished sentence.
Time Attack records one line per run (at time-up).

**Personal best (PB):** best WPM per `(mode, lang, source)` bucket, so a 30s
English run never competes with a Normal Korean run. Rounds shorter than 5
seconds are excluded from PB eligibility (they are too noisy to be a fair
pace).

**Result screen:** one extra line rendered by the shared result renderer:

- On a new PB: `NEW BEST! 72 wpm`
- Otherwise: `best 72 · recent ▂▄▆▅▇` — a sparkline of the last 10 rounds in
  the same `(mode, lang, source)` bucket, min–max scaled. The ASCII glyph set
  falls back to `.:-=+*#%`.

**`termtype --stats`:** prints a plain-text summary to stdout without starting
the TUI: per `(mode, lang)` bucket — run count, best WPM, average WPM, average
accuracy. Exits 0. Prints a friendly notice when there is no history yet.

## Feature 2 — Quick start (PR 2)

**Flags:** `-t/--theme <name>`, `-m/--mode <normal|30|60>`, `-l/--lang <en|ko>`.

- If **any** of the three is given, the selection menu is skipped; unspecified
  fields fall back to the last-used values from `config.json` (or current
  defaults when no config exists).
- Invalid values exit with an error listing the valid options.

**Menu memory:** whenever a game starts (menu or flags), the effective
selection is written to `config.json`. The menu opens with the last-used
theme/mode/lang preselected.

## Feature 3 — Custom text practice (PR 3)

**Input:** `termtype --file <path>`, or piped stdin (`cat notes.go | termtype`).
If both are present, `--file` wins.

**Chunking:** the text is split into lines; each line becomes one target
sentence.

- Blank lines are dropped.
- Leading/trailing whitespace is trimmed (typing indentation in a terminal is
  miserable).
- Internal runs of whitespace/tabs collapse to a single space.
- Long lines are fine — the renderer already wraps.
- If nothing remains after filtering, exit with an error.

**Ordering:** custom text plays sequentially in file order, wrapping around at
the end (builtin pools stay random). `GameState` grows an ordered-pool mode;
themes keep calling the same next-sentence API.

**Terminal input with piped stdin:** when stdin is a pipe, keyboard input is
reattached via `/dev/tty` (tcell's tty API). This is Unix-only; on Windows,
piped stdin exits with a message directing to `--file`.

**Menu:** `--file`/stdin does not itself skip the menu (theme/mode still
apply), but the language row is hidden — the file is the pool. Combined with
`-t`/`-m` it skips the menu as in Feature 2; `-l` is accepted but ignored for
custom-text runs.

**History:** recorded with `source: "custom"`, `lang: "-"`, so custom runs
never pollute builtin PBs.

## Feature 4 — Pace ghost (PR 4, opt-in, default off)

**Enable:** `--ghost` flag, or `g` toggle on the selection menu. The choice is
persisted in `config.json` (initial default: off).

**Pace:** ghost position = active typing time on the current sentence ×
PB WPM × 5 runes/minute, clamped to the target length. The PB is taken from
the matching `(mode, lang, builtin)` bucket. Ghosts only run on builtin
sentences — custom text difficulty varies too much to be a fair race.

**Rendering:** the shared typing renderer draws a dim underlined marker at the
ghost's rune position (one change covers every theme). The HUD shows
`ghost 58` next to the live WPM readout. The result screen adds one line:
`vs ghost: +4.2 wpm` (or `-2.1 wpm`).

**Smoothness:** with ghost enabled, the game ticker runs at 200ms instead of
1s so the marker moves smoothly. Pause freezes the ghost (it is driven by
active typing time, which already excludes paused spans).

**No PB yet:** the ghost silently does not appear.

## PR breakdown

| PR | Contents |
|----|----------|
| 1  | `internal/store`, round recording, result-screen PB/sparkline, `--stats` |
| 2  | quick-start flags, config-backed menu memory, preselected menu |
| 3  | `--file`/stdin custom text, sequential pool, `/dev/tty` input |
| 4  | pace ghost: flag + menu toggle, renderer marker, HUD, result line, 200ms ticker |

Each PR is independently releasable, updates the README for its feature, and
lands with unit tests.

## Testing

Following the existing table-driven test conventions:

- `internal/store`: append/read round-trip in a temp dir, corrupt-line
  tolerance, PB bucketing, 5-second PB eligibility rule, config round-trip.
- Sparkline scaling (including the ASCII fallback).
- Custom-text chunking rules (blank/whitespace/tab cases, empty-file error).
- Sequential pool ordering and wrap-around.
- Ghost position math (clamping, pause exclusion).
- Flag parsing/validation for `-t/-m/-l`, `--file`, `--ghost`, `--stats`.
