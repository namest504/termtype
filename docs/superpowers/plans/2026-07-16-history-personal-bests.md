# History & Personal Bests Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Persist every finished round, show personal-best / recent-progress feedback on the result screen, and add a `--stats` summary flag.

**Architecture:** A new `internal/store` package owns all persistence (config + JSONL history under `os.UserConfigDir()/termtype/`), with best-effort semantics — a nil or dir-less `Store` silently no-ops. The game records one round per `Finalize()` via a single `finalizeRound()` hook and renders the PB line theme-independently in the existing overlay (`drawOverlay`), so no theme files change. A `ui.Sparkline` helper renders the recent-WPM trend.

**Tech Stack:** Go 1.25, stdlib only (`encoding/json`, `bufio`, `os`). TUI via existing `tcell` + `runewidth`. Tests with `tcell.SimulationScreen` (existing pattern in `internal/app/overlay_test.go`).

**Spec:** `docs/superpowers/specs/2026-07-16-stats-quickstart-custom-ghost-design.md` (Feature 1).

**Spec deviation (intentional):** The spec says Time Attack records "one line per run (at time-up)". In the current implementation a Time Attack round IS one sentence raced against the countdown (finishing the sentence calls `Finalize()`, Enter starts a fresh countdown). So we record one line per `Finalize()` in every mode, which covers both sentence completion and time-up. Task 6 amends the spec wording.

## Global Constraints

- No new external dependencies — stdlib only on top of existing `tcell`/`runewidth`.
- **Never add Claude trailers or footers**: no `Co-Authored-By: Claude ...` in commits, no `Generated with Claude Code` in PR bodies. (Explicit user rule.)
- Workflow: branch `feat/history-stats` off up-to-date `main`, commit per task, one PR at the end, squash-merge after CI is green.
- Storage failures must never interrupt gameplay — every store operation is best-effort and nil-safe.
- Commit message style matches repo history: `feat: Add ...` / `docs: ...` (capitalized after colon).
- Code comments in English, matching existing density and tone.
- `gofmt -l .` must print nothing; `go test ./...` must pass before every commit.
- History JSONL schema (exact field names, from the spec): `ts`, `theme`, `mode` (`normal|ta30|ta60`), `lang` (`en|ko|-`), `wpm`, `acc`, `dur_s`, `source` (`builtin|custom`).
- Config JSON schema: `{"theme","mode","lang","ghost"}` — `ghost` is written now for forward-compatibility with Feature 4 but unused in this PR.
- PB eligibility: rounds shorter than 5 seconds never set a personal best.

---

### Task 1: `internal/store` — Store, Config, Round, history I/O

**Files:**
- Create: `internal/store/store.go`
- Test: `internal/store/store_test.go`

**Interfaces:**
- Consumes: nothing (stdlib only).
- Produces (used by Tasks 2, 4, 5):
  - `type Config struct { Theme, Mode, Lang string; Ghost bool }`
  - `type Round struct { TS time.Time; Theme, Mode, Lang string; WPM, Acc, DurS float64; Source string }`
  - `func New(dir string) *Store`
  - `func Default() *Store`
  - `func (s *Store) LoadConfig() Config` / `func (s *Store) SaveConfig(c Config)`
  - `func (s *Store) LoadHistory() []Round` / `func (s *Store) AppendRound(r Round)`
  - All methods are safe on a nil `*Store` and on a `Store` with an empty dir (no-ops).

- [ ] **Step 1: Create the branch**

```bash
cd /home/gnt/projects/03_personal/termtype
git checkout main && git pull --ff-only
git checkout -b feat/history-stats
```

- [ ] **Step 2: Write the failing tests**

Create `internal/store/store_test.go`:

```go
package store

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigRoundTrip(t *testing.T) {
	s := New(t.TempDir())
	want := Config{Theme: "matrix", Mode: "ta30", Lang: "ko", Ghost: true}
	s.SaveConfig(want)
	if got := s.LoadConfig(); got != want {
		t.Errorf("LoadConfig() = %+v, want %+v", got, want)
	}
}

func TestLoadConfigMissing(t *testing.T) {
	s := New(t.TempDir())
	if got := s.LoadConfig(); got != (Config{}) {
		t.Errorf("LoadConfig() with no file = %+v, want zero value", got)
	}
}

func TestHistoryAppendLoad(t *testing.T) {
	s := New(t.TempDir())
	r1 := Round{TS: time.Date(2026, 7, 16, 9, 0, 0, 0, time.UTC), Theme: "simple",
		Mode: "normal", Lang: "en", WPM: 60.5, Acc: 97.2, DurS: 12.3, Source: "builtin"}
	r2 := r1
	r2.WPM = 65.0
	s.AppendRound(r1)
	s.AppendRound(r2)
	got := s.LoadHistory()
	if len(got) != 2 || got[0].WPM != 60.5 || got[1].WPM != 65.0 {
		t.Errorf("LoadHistory() = %+v, want the two appended rounds in order", got)
	}
	if !got[0].TS.Equal(r1.TS) || got[0].Theme != "simple" || got[0].Source != "builtin" {
		t.Errorf("LoadHistory()[0] = %+v, want %+v", got[0], r1)
	}
}

func TestLoadHistorySkipsCorruptLines(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	s.AppendRound(Round{TS: time.Now(), WPM: 50, DurS: 10})
	f, err := os.OpenFile(filepath.Join(dir, "history.jsonl"), os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("not json\n{}\n")
	f.Close()
	s.AppendRound(Round{TS: time.Now(), WPM: 55, DurS: 10})
	if got := s.LoadHistory(); len(got) != 2 {
		t.Errorf("LoadHistory() returned %d rounds, want 2 (corrupt lines skipped)", len(got))
	}
}

func TestNoopStore(t *testing.T) {
	var nilStore *Store
	for name, s := range map[string]*Store{"nil": nilStore, "empty": {}} {
		s.SaveConfig(Config{Theme: "x"})
		s.AppendRound(Round{TS: time.Now()})
		if got := s.LoadHistory(); got != nil {
			t.Errorf("%s store LoadHistory() = %v, want nil", name, got)
		}
		if got := s.LoadConfig(); got != (Config{}) {
			t.Errorf("%s store LoadConfig() = %+v, want zero value", name, got)
		}
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `go test ./internal/store/`
Expected: FAIL — `undefined: New`, `undefined: Config`, etc. (package doesn't compile yet).

- [ ] **Step 4: Write the implementation**

Create `internal/store/store.go`:

```go
// Package store persists TermType's config and round history under the
// user's config directory. Every operation is best-effort: storage failures
// never interrupt a game, they just disable persistence.
package store

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Config holds the last-used selections and toggles. Ghost is written for
// forward-compatibility; the pace-ghost feature reads it later.
type Config struct {
	Theme string `json:"theme"`
	Mode  string `json:"mode"`
	Lang  string `json:"lang"`
	Ghost bool   `json:"ghost"`
}

// Round is one finished typing round — one line in history.jsonl.
type Round struct {
	TS     time.Time `json:"ts"`
	Theme  string    `json:"theme"`
	Mode   string    `json:"mode"`   // "normal" | "ta30" | "ta60"
	Lang   string    `json:"lang"`   // "en" | "ko" | "-" (custom text)
	WPM    float64   `json:"wpm"`
	Acc    float64   `json:"acc"`
	DurS   float64   `json:"dur_s"`
	Source string    `json:"source"` // "builtin" | "custom"
}

// Store reads and writes files under one directory. A nil Store or one with
// an empty dir silently no-ops, so callers never branch on availability.
type Store struct{ dir string }

// New returns a Store rooted at dir.
func New(dir string) *Store { return &Store{dir: dir} }

// Default returns a Store under os.UserConfigDir()/termtype, or a no-op
// Store when the config dir cannot be resolved.
func Default() *Store {
	base, err := os.UserConfigDir()
	if err != nil {
		return &Store{}
	}
	return &Store{dir: filepath.Join(base, "termtype")}
}

func (s *Store) usable() bool { return s != nil && s.dir != "" }

func (s *Store) configPath() string  { return filepath.Join(s.dir, "config.json") }
func (s *Store) historyPath() string { return filepath.Join(s.dir, "history.jsonl") }

// LoadConfig returns the saved config, or a zero Config on any error.
func (s *Store) LoadConfig() Config {
	var c Config
	if !s.usable() {
		return c
	}
	b, err := os.ReadFile(s.configPath())
	if err != nil {
		return c
	}
	if json.Unmarshal(b, &c) != nil {
		return Config{}
	}
	return c
}

// SaveConfig writes the config, best-effort.
func (s *Store) SaveConfig(c Config) {
	if !s.usable() || os.MkdirAll(s.dir, 0o755) != nil {
		return
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(s.configPath(), b, 0o644)
}

// LoadHistory returns all parseable rounds in file order. Corrupt lines are
// skipped; a missing or unreadable file yields nil.
func (s *Store) LoadHistory() []Round {
	if !s.usable() {
		return nil
	}
	f, err := os.Open(s.historyPath())
	if err != nil {
		return nil
	}
	defer f.Close()
	var rounds []Round
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var r Round
		// A zero timestamp means the line was empty or junk (e.g. "{}").
		if json.Unmarshal(sc.Bytes(), &r) == nil && !r.TS.IsZero() {
			rounds = append(rounds, r)
		}
	}
	return rounds
}

// AppendRound appends one round to the history file, best-effort.
func (s *Store) AppendRound(r Round) {
	if !s.usable() || os.MkdirAll(s.dir, 0o755) != nil {
		return
	}
	b, err := json.Marshal(r)
	if err != nil {
		return
	}
	f, err := os.OpenFile(s.historyPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(b, '\n'))
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/store/`
Expected: PASS (5 tests).

- [ ] **Step 6: Format and commit**

```bash
gofmt -l . && go test ./... && git add internal/store/ && \
git commit -m "feat: Add persistent store for config and round history"
```

Expected: `gofmt -l .` prints nothing, all tests pass, commit succeeds.

---

### Task 2: `internal/store` — PB/recent helpers and `--stats` formatting

**Files:**
- Create: `internal/store/stats.go`
- Test: `internal/store/stats_test.go`

**Interfaces:**
- Consumes: `Round` from Task 1.
- Produces (used by Tasks 4, 5):
  - `type Key struct { Mode, Lang, Source string }`
  - `func KeyOf(r Round) Key`
  - `func PBEligible(r Round) bool` — false for rounds shorter than 5s
  - `func Best(rounds []Round, k Key) (float64, bool)`
  - `func RecentWPMs(rounds []Round, k Key, n int) []float64` — oldest first
  - `func ModeString(limit time.Duration) string` — `"normal"`, `"ta30"`, `"ta60"`
  - `func FormatStats(rounds []Round) string`

- [ ] **Step 1: Write the failing tests**

Create `internal/store/stats_test.go`:

```go
package store

import (
	"strings"
	"testing"
	"time"
)

func TestBestRespectsBucketAndEligibility(t *testing.T) {
	rounds := []Round{
		{Mode: "normal", Lang: "en", Source: "builtin", WPM: 90, DurS: 3},  // too short for PB
		{Mode: "normal", Lang: "en", Source: "builtin", WPM: 70, DurS: 10},
		{Mode: "normal", Lang: "ko", Source: "builtin", WPM: 80, DurS: 10}, // other bucket
		{Mode: "ta30", Lang: "en", Source: "builtin", WPM: 85, DurS: 30},   // other bucket
		{Mode: "normal", Lang: "en", Source: "custom", WPM: 95, DurS: 10},  // other bucket
	}
	best, ok := Best(rounds, Key{Mode: "normal", Lang: "en", Source: "builtin"})
	if !ok || best != 70 {
		t.Errorf("Best() = %v, %v; want 70, true", best, ok)
	}
}

func TestBestEmpty(t *testing.T) {
	if _, ok := Best(nil, Key{Mode: "normal", Lang: "en", Source: "builtin"}); ok {
		t.Error("Best() on empty history reported ok")
	}
}

func TestRecentWPMsLastN(t *testing.T) {
	k := Key{Mode: "normal", Lang: "en", Source: "builtin"}
	var rounds []Round
	for i := 1; i <= 12; i++ {
		rounds = append(rounds, Round{Mode: "normal", Lang: "en", Source: "builtin",
			WPM: float64(i), DurS: 10})
	}
	rounds = append(rounds, Round{Mode: "ta30", Lang: "en", Source: "builtin", WPM: 99, DurS: 30})
	got := RecentWPMs(rounds, k, 10)
	if len(got) != 10 || got[0] != 3 || got[9] != 12 {
		t.Errorf("RecentWPMs() = %v, want WPMs 3 through 12", got)
	}
}

func TestModeString(t *testing.T) {
	cases := []struct {
		limit time.Duration
		want  string
	}{{0, "normal"}, {30 * time.Second, "ta30"}, {60 * time.Second, "ta60"}}
	for _, c := range cases {
		if got := ModeString(c.limit); got != c.want {
			t.Errorf("ModeString(%v) = %q, want %q", c.limit, got, c.want)
		}
	}
}

func TestFormatStats(t *testing.T) {
	rounds := []Round{
		{Mode: "normal", Lang: "en", Source: "builtin", WPM: 60, Acc: 95, DurS: 10},
		{Mode: "normal", Lang: "en", Source: "builtin", WPM: 70, Acc: 97, DurS: 10},
	}
	out := FormatStats(rounds)
	for _, want := range []string{"normal", "en", "runs 2", "best  70.0", "avg  65.0", "acc 96.0%"} {
		if !strings.Contains(out, want) {
			t.Errorf("FormatStats() output %q missing %q", out, want)
		}
	}
}

func TestFormatStatsEmpty(t *testing.T) {
	if out := FormatStats(nil); !strings.Contains(out, "No history") {
		t.Errorf("FormatStats(nil) = %q, want a no-history notice", out)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/store/`
Expected: FAIL — `undefined: Best`, `undefined: Key`, etc.

- [ ] **Step 3: Write the implementation**

Create `internal/store/stats.go`:

```go
package store

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// minPBDurS is the shortest round (seconds) eligible for a personal best;
// shorter rounds are too noisy to be a fair pace.
const minPBDurS = 5.0

// Key identifies a comparable bucket of rounds. Bests and recents are always
// computed within one bucket so different modes/languages never compete.
type Key struct {
	Mode   string
	Lang   string
	Source string
}

// KeyOf returns the bucket a round belongs to.
func KeyOf(r Round) Key { return Key{Mode: r.Mode, Lang: r.Lang, Source: r.Source} }

// PBEligible reports whether a round may set a personal best.
func PBEligible(r Round) bool { return r.DurS >= minPBDurS }

// ModeString converts a time-attack limit to its history/config value.
func ModeString(limit time.Duration) string {
	if limit <= 0 {
		return "normal"
	}
	return fmt.Sprintf("ta%d", int(limit.Seconds()))
}

// Best returns the highest PB-eligible WPM in the bucket, and whether one exists.
func Best(rounds []Round, k Key) (float64, bool) {
	best, ok := 0.0, false
	for _, r := range rounds {
		if KeyOf(r) != k || !PBEligible(r) {
			continue
		}
		if !ok || r.WPM > best {
			best, ok = r.WPM, true
		}
	}
	return best, ok
}

// RecentWPMs returns the WPMs of the last n rounds in the bucket, oldest first.
func RecentWPMs(rounds []Round, k Key, n int) []float64 {
	var wpms []float64
	for _, r := range rounds {
		if KeyOf(r) == k {
			wpms = append(wpms, r.WPM)
		}
	}
	if len(wpms) > n {
		wpms = wpms[len(wpms)-n:]
	}
	return wpms
}

// FormatStats renders the --stats summary: one line per bucket with run
// count, best and average WPM, and average accuracy.
func FormatStats(rounds []Round) string {
	if len(rounds) == 0 {
		return "No history yet — play a round first!\n"
	}
	type agg struct {
		n      int
		wpmSum float64
		accSum float64
	}
	aggs := map[Key]*agg{}
	for _, r := range rounds {
		k := KeyOf(r)
		if aggs[k] == nil {
			aggs[k] = &agg{}
		}
		aggs[k].n++
		aggs[k].wpmSum += r.WPM
		aggs[k].accSum += r.Acc
	}
	keys := make([]Key, 0, len(aggs))
	for k := range aggs {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Mode != keys[j].Mode {
			return keys[i].Mode < keys[j].Mode
		}
		if keys[i].Lang != keys[j].Lang {
			return keys[i].Lang < keys[j].Lang
		}
		return keys[i].Source < keys[j].Source
	})
	var b strings.Builder
	for _, k := range keys {
		a := aggs[k]
		best, _ := Best(rounds, k)
		src := ""
		if k.Source != "builtin" {
			src = " (" + k.Source + ")"
		}
		fmt.Fprintf(&b, "%-7s %-3s%s runs %-4d best %5.1f  avg %5.1f wpm  acc %.1f%%\n",
			k.Mode, k.Lang, src, a.n, best, a.wpmSum/float64(a.n), a.accSum/float64(a.n))
	}
	return b.String()
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/store/`
Expected: PASS (all store tests).

- [ ] **Step 5: Format and commit**

```bash
gofmt -l . && go test ./... && git add internal/store/ && \
git commit -m "feat: Add personal-best and stats helpers"
```

---

### Task 3: `ui.Sparkline`

**Files:**
- Create: `internal/ui/sparkline.go`
- Test: `internal/ui/sparkline_test.go`

**Interfaces:**
- Consumes: `ui.IsASCII()` / `ui.SetASCII()` from `internal/ui/glyphs.go` (existing).
- Produces (used by Task 4): `func Sparkline(values []float64) string` — one glyph per value, min–max scaled, `""` for empty input.

- [ ] **Step 1: Write the failing tests**

Create `internal/ui/sparkline_test.go`:

```go
package ui

import "testing"

func TestSparklineScaling(t *testing.T) {
	SetASCII(false)
	// min=1 -> lowest glyph, max=8 -> highest, 4.5 -> level 3 of 0..7.
	if got := Sparkline([]float64{1, 8, 4.5}); got != "▁█▄" {
		t.Errorf("Sparkline() = %q, want ▁█▄", got)
	}
}

func TestSparklineFlat(t *testing.T) {
	SetASCII(false)
	// Equal values sit at the middle level so a flat series stays visible.
	if got := Sparkline([]float64{5, 5}); got != "▅▅" {
		t.Errorf("Sparkline() = %q, want ▅▅", got)
	}
}

func TestSparklineASCII(t *testing.T) {
	SetASCII(true)
	defer SetASCII(false)
	if got := Sparkline([]float64{1, 8}); got != ".%" {
		t.Errorf("Sparkline() = %q, want .%%", got)
	}
}

func TestSparklineEmpty(t *testing.T) {
	if got := Sparkline(nil); got != "" {
		t.Errorf("Sparkline(nil) = %q, want empty string", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/ui/`
Expected: FAIL — `undefined: Sparkline`.

- [ ] **Step 3: Write the implementation**

Create `internal/ui/sparkline.go`:

```go
package ui

// Sparkline level glyphs, lowest to highest. The ASCII set mirrors the
// GlyphSet fallback idea for terminals that render block elements as tofu.
var (
	sparkUnicode = []rune("▁▂▃▄▅▆▇█")
	sparkASCII   = []rune(".:-=+*#%")
)

// Sparkline renders values as one glyph per value, min–max scaled to the
// glyph range. Equal values render at a middle level so a flat series stays
// visible. An empty input yields an empty string.
func Sparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}
	glyphs := sparkUnicode
	if IsASCII() {
		glyphs = sparkASCII
	}
	lo, hi := values[0], values[0]
	for _, v := range values[1:] {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	out := make([]rune, len(values))
	for i, v := range values {
		level := len(glyphs) / 2
		if hi > lo {
			level = int((v - lo) / (hi - lo) * float64(len(glyphs)-1))
		}
		out[i] = glyphs[level]
	}
	return string(out)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/ui/`
Expected: PASS.

- [ ] **Step 5: Format and commit**

```bash
gofmt -l . && go test ./... && git add internal/ui/sparkline.go internal/ui/sparkline_test.go && \
git commit -m "feat: Add sparkline renderer for WPM trends"
```

---

### Task 4: Record rounds in the game and show the PB line in the overlay

**Files:**
- Modify: `internal/app/game.go`
- Test: Create `internal/app/record_test.go`; extend `internal/app/overlay_test.go`

**Interfaces:**
- Consumes: `store.Store`, `store.Round`, `store.KeyOf`, `store.Best`, `store.PBEligible`, `store.RecentWPMs`, `store.ModeString` (Tasks 1–2); `ui.Sparkline` (Task 3); `ui.Glyphs().Sep` (existing).
- Produces (used by Task 5):
  - `type RoundMeta struct { Theme, Lang, Source string }`
  - `func NewGame(s tcell.Screen, theme domain.Theme, timeLimit time.Duration, sentences []string, meta RoundMeta, st *store.Store) (*Game, error)` — note the two new trailing params.

- [ ] **Step 1: Write the failing tests**

Create `internal/app/record_test.go` (helpers `newSimScreen`, `newGame` already exist in `overlay_test.go`, same package):

```go
package app

import (
	"strings"
	"testing"
	"time"

	"termtype/internal/domain"
	"termtype/internal/store"
)

// newRecordingGame returns a Game whose round is fully typed and ~10s old,
// ready for finalizeRound, persisting to a store rooted at dir.
func newRecordingGame(t *testing.T, dir string) *Game {
	t.Helper()
	ss := newSimScreen(t, 60, 12)
	st := &domain.GameState{TargetSentence: "hello world", UserInput: "hello world",
		TimerStarted: true, StartTime: time.Now().Add(-10 * time.Second)}
	g := newGame(ss, st)
	g.meta = RoundMeta{Theme: "simple", Lang: "en", Source: "builtin"}
	g.store = store.New(dir)
	g.history = g.store.LoadHistory()
	return g
}

func TestFinalizeRoundRecordsHistory(t *testing.T) {
	dir := t.TempDir()
	g := newRecordingGame(t, dir)
	g.finalizeRound()

	rounds := store.New(dir).LoadHistory()
	if len(rounds) != 1 {
		t.Fatalf("history has %d rounds, want 1", len(rounds))
	}
	r := rounds[0]
	if r.Theme != "simple" || r.Lang != "en" || r.Mode != "normal" || r.Source != "builtin" {
		t.Errorf("recorded round meta = %+v", r)
	}
	if r.WPM <= 0 || r.DurS < 9 {
		t.Errorf("recorded round stats = %+v, want positive wpm and ~10s duration", r)
	}
	if !g.state.IsFinished {
		t.Error("finalizeRound did not finalize the game state")
	}
}

func TestFinalizeRoundFirstRoundIsNewBest(t *testing.T) {
	g := newRecordingGame(t, t.TempDir())
	g.finalizeRound()
	if !strings.Contains(g.resultLine, "NEW BEST") {
		t.Errorf("resultLine = %q, want NEW BEST on first eligible round", g.resultLine)
	}
}

func TestFinalizeRoundShowsBestAndSparkline(t *testing.T) {
	dir := t.TempDir()
	seed := store.New(dir)
	for i := 0; i < 3; i++ {
		seed.AppendRound(store.Round{TS: time.Now(), Theme: "simple", Mode: "normal",
			Lang: "en", WPM: 200, Acc: 99, DurS: 10, Source: "builtin"})
	}
	g := newRecordingGame(t, dir)
	g.finalizeRound()
	if !strings.Contains(g.resultLine, "best 200") {
		t.Errorf("resultLine = %q, want existing best 200 shown", g.resultLine)
	}
	if strings.Contains(g.resultLine, "NEW BEST") {
		t.Errorf("resultLine = %q must not claim a new best below 200 wpm", g.resultLine)
	}
}

func TestFinalizeRoundNilStore(t *testing.T) {
	ss := newSimScreen(t, 60, 12)
	st := &domain.GameState{TargetSentence: "hi", UserInput: "hi",
		TimerStarted: true, StartTime: time.Now().Add(-6 * time.Second)}
	g := newGame(ss, st)
	g.finalizeRound() // must not panic with no store/meta configured
	if !g.state.IsFinished {
		t.Error("finalizeRound did not finalize with a nil store")
	}
}
```

Append to `internal/app/overlay_test.go`:

```go
// The result line (PB / recent sparkline) is drawn on the top row once the
// round is finished.
func TestOverlay_ResultLine(t *testing.T) {
	ss := newSimScreen(t, 60, 12)
	st := &domain.GameState{IsFinished: true}
	g := newGame(ss, st)
	g.resultLine = " best 72 · recent ▄▆█ "
	g.drawOverlay()
	ss.Show()
	if got := rowString(ss, 0); !strings.Contains(got, "best 72") {
		t.Errorf("top row = %q, want the result line", got)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/app/`
Expected: FAIL — `g.meta undefined`, `g.finalizeRound undefined`, `g.resultLine undefined`.

- [ ] **Step 3: Implement in `internal/app/game.go`**

3a. Add `"termtype/internal/store"` to the imports.

3b. Add `RoundMeta` and extend the `Game` struct (replace the existing struct definition):

```go
// RoundMeta labels the rounds a game produces for the history file.
type RoundMeta struct {
	Theme  string
	Lang   string // "en" | "ko" | "-" for custom text
	Source string // "builtin" | "custom"
}

// Struct that manages the entire game
type Game struct {
	screen   tcell.Screen
	renderer *ui.Renderer
	state    *domain.GameState
	theme    domain.Theme

	meta       RoundMeta
	store      *store.Store
	history    []store.Round // loaded once at startup, appended in memory
	resultLine string        // PB/recent line for the finished round ("" = hide)
	resultBest bool          // true when resultLine announces a new PB
}
```

3c. Replace `NewGame` (only `cmd/termtype/main.go` calls it; it is updated in Task 5):

```go
// Create a new game. timeLimit > 0 enables time-attack mode. sentences is the
// pool the chosen theme draws targets from; an empty pool falls back to the
// default English set. meta labels recorded rounds; st may be nil to disable
// persistence.
func NewGame(s tcell.Screen, theme domain.Theme, timeLimit time.Duration, sentences []string, meta RoundMeta, st *store.Store) (*Game, error) {
	if len(sentences) == 0 {
		sentences = domain.Sentences
	}
	state := &domain.GameState{Sentences: sentences, TimeLimit: timeLimit}
	theme.ResetState(state)

	return &Game{screen: s, renderer: ui.NewRenderer(s), state: state, theme: theme,
		meta: meta, store: st, history: st.LoadHistory()}, nil
}
```

3d. Add `finalizeRound` below `Run`:

```go
// finalizeRound ends the round, appends it to the history, and prepares the
// result line the overlay shows: a NEW BEST banner, or the current best with
// a sparkline of the last 10 rounds in the same (mode, lang, source) bucket.
func (g *Game) finalizeRound() {
	g.state.Finalize()

	r := store.Round{
		TS:     time.Now(),
		Theme:  g.meta.Theme,
		Mode:   store.ModeString(g.state.TimeLimit),
		Lang:   g.meta.Lang,
		WPM:    g.state.Wpm,
		Acc:    g.state.Accuracy,
		DurS:   g.state.Elapsed().Seconds(),
		Source: g.meta.Source,
	}
	k := store.KeyOf(r)
	prevBest, hadBest := store.Best(g.history, k)
	g.history = append(g.history, r)
	g.store.AppendRound(r)

	switch {
	case store.PBEligible(r) && (!hadBest || r.WPM > prevBest):
		g.resultLine = fmt.Sprintf(" NEW BEST! %.0f wpm ", r.WPM)
		g.resultBest = true
	case hadBest:
		spark := ui.Sparkline(store.RecentWPMs(g.history, k, 10))
		g.resultLine = fmt.Sprintf(" best %.0f %s recent %s ", prevBest, ui.Glyphs().Sep, spark)
		g.resultBest = false
	default:
		g.resultLine = ""
	}
}
```

3e. Route both `Finalize` call sites through it. In `Run`'s ticker branch replace:

```go
g.state.TimedOut = true
g.state.Finalize()
```

with:

```go
g.state.TimedOut = true
g.finalizeRound()
```

In `handleKeyEvent` replace:

```go
	// Finalize once the whole sentence has been typed (rune count, not bytes).
	if len([]rune(g.state.UserInput)) >= len([]rune(g.state.TargetSentence)) {
		g.state.Finalize()
	}
```

with:

```go
	// Finalize once the whole sentence has been typed (rune count, not bytes).
	if len([]rune(g.state.UserInput)) >= len([]rune(g.state.TargetSentence)) {
		g.finalizeRound()
	}
```

3f. Draw the line in `drawOverlay`. Insert after the live-stats block (the `if g.state.TimerStarted && !g.state.IsFinished ...` block) and before the pause block:

```go
	// Once the round is finished the live stats disappear, freeing the top-right
	// corner for the personal-best line.
	if g.state.IsFinished && g.resultLine != "" {
		style := tcell.StyleDefault.Foreground(tcell.ColorGray)
		if g.resultBest {
			style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorGreen)
		}
		if runewidth.StringWidth(g.resultLine) <= right {
			g.drawRightAligned(g.resultLine, 0, right, style)
		}
	}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/app/`
Expected: PASS. (`./...` will still fail to build: `cmd/termtype` calls the old `NewGame` — expected until Task 5. Verify with `go build ./...`: the only error must be in `cmd/termtype/main.go`.)

- [ ] **Step 5: Commit**

```bash
gofmt -l . && go test ./internal/... && git add internal/app/ && \
git commit -m "feat: Record finished rounds and show the personal best in the overlay"
```

---

### Task 5: Wire `main.go` — round metadata and `--stats`

**Files:**
- Modify: `cmd/termtype/main.go`

**Interfaces:**
- Consumes: `app.NewGame` new signature + `app.RoundMeta` (Task 4); `store.Default`, `store.FormatStats` (Tasks 1–2).
- Produces: `termtype --stats` CLI behavior; language codes `"en"`/`"ko"` recorded per run.

- [ ] **Step 1: Update `cmd/termtype/main.go`**

1a. Add `"termtype/internal/store"` to the imports.

1b. Give languages a code (replace the `language` type and `languages` var):

```go
type language struct {
	name      string
	code      string // history/config value: "en" | "ko"
	sentences []string
}

var languages = []language{
	{"English", "en", domain.Sentences},
	{"한국어 (Korean)", "ko", domain.KoreanSentences},
}
```

1c. Return the selection wholesale. Replace `selectTheme`'s signature and its `KeyEnter` case:

```go
// selection is everything the menu picks: the theme (and its registry name,
// recorded in history), the mode limit, and the language.
type selection struct {
	theme     domain.Theme
	themeName string
	limit     time.Duration
	lang      language
}

func selectTheme(s tcell.Screen) (selection, error) {
```

Every `return nil, 0, nil, fmt.Errorf(...)` in the function becomes:

```go
				return selection{}, fmt.Errorf("theme selection cancelled")
```

and the `KeyEnter` case becomes:

```go
			case tcell.KeyEnter:
				name := themeNames[selectedIndex]
				return selection{theme: themes.Themes[name], themeName: name,
					limit: gameModes[modeIndex].limit, lang: languages[langIndex]}, nil
```

1d. Add the `--stats` flag next to the existing flags in `main`:

```go
	statsFlag := flag.Bool("stats", false, "Print a typing history summary and exit")
```

and immediately after the version check:

```go
	if *statsFlag {
		fmt.Print(store.FormatStats(store.Default().LoadHistory()))
		os.Exit(0)
	}
```

1e. Update the game construction at the bottom of `main`:

```go
	// Select theme, mode, and language
	sel, err := selectTheme(s)
	if err != nil {
		return // user cancelled; the deferred Fini restores the terminal
	}

	// Create and run the game
	meta := app.RoundMeta{Theme: sel.themeName, Lang: sel.lang.code, Source: "builtin"}
	game, err := app.NewGame(s, sel.theme, sel.limit, sel.lang.sentences, meta, store.Default())
	if err != nil {
		return
	}

	game.Run()
```

- [ ] **Step 2: Build and run the full test suite**

Run: `go build ./... && go test ./...`
Expected: build succeeds, all tests PASS.

- [ ] **Step 3: Smoke-test `--stats`**

Run: `XDG_CONFIG_HOME=$(mktemp -d) go run ./cmd/termtype --stats`
Expected output: `No history yet — play a round first!`

- [ ] **Step 4: Format and commit**

```bash
gofmt -l . && git add cmd/termtype/main.go && \
git commit -m "feat: Add --stats flag and record language and theme per round"
```

---

### Task 6: Docs, spec amendment, and the PR

**Files:**
- Modify: `README.md`
- Modify: `docs/superpowers/specs/2026-07-16-stats-quickstart-custom-ghost-design.md`

- [ ] **Step 1: Update the README**

In `README.md`, insert a new section between the `## Usage` block (after the "A live WPM/accuracy readout..." paragraph) and `### Any terminal, any width`:

````markdown
### History & personal bests

Every finished round is saved to `~/.config/termtype/history.jsonl` (or your
platform's config directory). The result screen shows `NEW BEST!` when you set
a personal best, or your current best plus a sparkline of the last 10 rounds.
Bests are tracked separately per mode and language.

See a summary of your history without starting the game:

```bash
termtype --stats
```
````

- [ ] **Step 2: Amend the spec's recording-unit wording**

In `docs/superpowers/specs/2026-07-16-stats-quickstart-custom-ghost-design.md`, replace:

```markdown
**Recording unit:** Normal mode records one line per finished sentence.
Time Attack records one line per run (at time-up).
```

with:

```markdown
**Recording unit:** one line per finished round. In Normal mode a round is
one sentence; in Time Attack a round ends at sentence completion or time-up
(the countdown races one sentence at a time in the current implementation).
```

- [ ] **Step 3: Final verification**

```bash
gofmt -l . && go vet ./... && go test ./...
```

Expected: no gofmt output, vet clean, all tests PASS.

- [ ] **Step 4: Commit and push**

```bash
git add README.md docs/superpowers/specs/2026-07-16-stats-quickstart-custom-ghost-design.md
git commit -m "docs: Document history, personal bests, and --stats"
git push -u origin feat/history-stats
```

- [ ] **Step 5: Open the PR (NO Claude footer)**

```bash
gh pr create --title "feat: History and personal bests" --body "## Summary

Implements Feature 1 of the design spec (docs/superpowers/specs/2026-07-16-stats-quickstart-custom-ghost-design.md):

- New \`internal/store\` package: best-effort persistence under \`os.UserConfigDir()/termtype/\` — \`config.json\` + append-only \`history.jsonl\`. Storage failures never interrupt a game.
- Every finished round is recorded (theme, mode, lang, wpm, acc, duration, source).
- Result screen shows \`NEW BEST!\` on a personal best, otherwise the current best plus a sparkline of the last 10 rounds — rendered in the shared overlay, so all themes get it without changes.
- Personal bests are bucketed per (mode, lang, source); rounds under 5s are not PB-eligible.
- New \`termtype --stats\` prints a per-bucket summary without starting the TUI.

## Testing

- Unit tests: store round-trip, corrupt-line tolerance, no-op store safety, PB bucketing/eligibility, sparkline scaling + ASCII fallback, round recording, overlay result line.
- Manual: \`termtype --stats\` smoke test with a fresh config dir."
```

- [ ] **Step 6: Merge after CI is green**

```bash
gh pr checks --watch
gh pr merge --squash --delete-branch --subject "feat: Add round history, personal bests, and --stats" --body ""
git checkout main && git pull --ff-only
```

Expected: CI green, PR merged, local main synced.
