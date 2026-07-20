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
	Theme  string `json:"theme"`
	Mode   string `json:"mode"`
	Lang   string `json:"lang"`
	Source string `json:"source"` // "builtin" | "words"
	Ghost  bool   `json:"ghost"`
}

// Round is one finished typing round — one line in history.jsonl. The
// RawWPM/CPM/WPMSeries fields were added later and are absent from older
// lines, which decode with zero values.
type Round struct {
	TS        time.Time `json:"ts"`
	Theme     string    `json:"theme"`
	Mode      string    `json:"mode"` // "normal" | "ta15" | "ta30" | "ta60"
	Lang      string    `json:"lang"` // "en" | "ko" | "-" (custom text)
	WPM       float64   `json:"wpm"`
	Acc       float64   `json:"acc"`
	DurS      float64   `json:"dur_s"`
	Source    string    `json:"source"` // "builtin" | "words" | "custom"
	RawWPM    float64   `json:"raw_wpm,omitempty"`
	CPM       float64   `json:"cpm,omitempty"`
	WPMSeries []float64 `json:"wpm_series,omitempty"` // one sample per second
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
