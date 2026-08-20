package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/namest504/termtype/internal/app"
	"github.com/namest504/termtype/internal/chart"
	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/store"
	"github.com/namest504/termtype/internal/themes"
	"github.com/namest504/termtype/internal/ui"
)

var version = "dev"

type gameMode struct {
	name  string
	limit time.Duration
}

var gameModes = []gameMode{
	{"Normal", 0},
	{"Time Attack (15s)", 15 * time.Second},
	{"Time Attack (30s)", 30 * time.Second},
	{"Time Attack (60s)", 60 * time.Second},
}

// textSource is what the round asks you to type: the built-in sentence pool,
// or a random stream of common English words.
type textSource struct {
	name string
	code string // history/config value: "builtin" | "words"
}

var textSources = []textSource{
	{"Sentences", "builtin"},
	{"Words", "words"},
}

// normalWordCount is the length of a words round without a time limit.
// streamWordCount is the words-mode buffer for a timed round: long enough
// that the clock always runs out first.
const (
	normalWordCount = 25
	streamWordCount = 250
)

// resolveASCII decides whether to use the plain-ASCII glyph set. An explicit
// --ascii flag or TERMTYPE_ASCII env var wins; otherwise we keep Unicode unless
// the locale is clearly not UTF-8 (e.g. "C"/"POSIX"), since most modern
// terminals are UTF-8 even when the locale is unset.
func resolveASCII(flagSet bool) bool {
	if flagSet {
		return true
	}
	switch strings.ToLower(os.Getenv("TERMTYPE_ASCII")) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	loc := os.Getenv("LC_ALL")
	if loc == "" {
		loc = os.Getenv("LC_CTYPE")
	}
	if loc == "" {
		loc = os.Getenv("LANG")
	}
	if loc == "" {
		return false // unknown locale → assume a modern UTF-8 terminal
	}
	loc = strings.ToLower(loc)
	return !strings.Contains(loc, "utf-8") && !strings.Contains(loc, "utf8")
}

type language struct {
	name      string
	code      string // history/config value: "en" | "ko"
	sentences []string
}

var languages = []language{
	{"English", "en", domain.Sentences},
	{"한국어 (Korean)", "ko", domain.KoreanSentences},
}

func drawText(s tcell.Screen, x, y int, style tcell.Style, text string) {
	col := x
	for _, r := range []rune(text) {
		s.SetContent(col, y, r, nil, style)
		col += runewidth.RuneWidth(r)
	}
}

// chartOptionsFor maps a config style code onto chart options. Unknown
// codes fall back to the default so an edited config never breaks startup.
func chartOptionsFor(code string) chart.Options {
	o := chart.Options{Style: chart.StyleBraille, Interp: chart.InterpSmooth, Thickness: 2}
	switch code {
	case "braille1":
		o.Thickness = 1
	case "braille3":
		o.Thickness = 3
	case "box":
		o.Style, o.Thickness = chart.StyleBox, 1
	}
	return o
}

// selection is everything the menu picks: the theme (and its registry name,
// recorded in history), the mode limit, the text source, and the language.
type selection struct {
	theme     domain.Theme
	themeName string
	limit     time.Duration
	src       textSource
	lang      language
	graphOn   bool
}

// indexOf returns the index where match is true, or 0.
func indexOf(n int, match func(int) bool) int {
	for i := 0; i < n; i++ {
		if match(i) {
			return i
		}
	}
	return 0
}

// runMenu is the menu ↔ settings/history hub. It returns the round
// selection on Enter, or an error when the player quits.
func runMenu(s tcell.Screen, events <-chan tcell.Event, cfg *store.Config, st *store.Store) (selection, error) {
	m := newMenuModel(cfg.Theme)
	for {
		drawMenu(s, m, summaryLine(*cfg))
		switch ev := (<-events).(type) {
		case nil:
			return selection{}, fmt.Errorf("screen closed")
		case *tcell.EventResize:
			s.Sync()
		case *tcell.EventKey:
			switch m.handleKey(ev) {
			case actQuit:
				return selection{}, fmt.Errorf("menu cancelled")
			case actSettings:
				runSettings(s, events, cfg, st)
			case actHistory:
				showHistory(s, events, st.LoadHistory())
			case actStart:
				name := m.themes[m.idx]
				sm := newSettingsModel(*cfg)
				return selection{
					theme:     themes.Themes[name],
					themeName: name,
					limit:     gameModes[sm.modeIdx].limit,
					src:       textSources[sm.srcIdx],
					lang:      languages[sm.langIdx],
					graphOn:   cfg.GraphAuto(),
				}, nil
			}
		}
	}
}

// summaryLine is the read-only settings recap under the carousel.
func summaryLine(cfg store.Config) string {
	sm := newSettingsModel(cfg)
	lang := languages[sm.langIdx].name
	if textSources[sm.srcIdx].code == "words" {
		lang = "English"
	}
	sep := " " + ui.Glyphs().Sep + " "
	return gameModes[sm.modeIdx].name + sep + textSources[sm.srcIdx].name + sep + lang
}

func main() {
	versionFlag := flag.Bool("version", false, "Print version information")
	vFlag := flag.Bool("v", false, "Print version information (shorthand)")
	asciiFlag := flag.Bool("ascii", false, "Use ASCII-only symbols (for terminals/fonts that can't render the Unicode glyphs)")
	statsFlag := flag.Bool("stats", false, "Print a typing history summary and exit")
	flag.Parse()

	if *versionFlag || *vFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if *statsFlag {
		fmt.Print(store.FormatStats(store.Default().LoadHistory()))
		os.Exit(0)
	}

	ui.SetASCII(resolveASCII(*asciiFlag))

	// Initialize screen
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	// Restore the terminal on any return path or panic in the game loop.
	defer s.Fini()

	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	s.SetStyle(defStyle)
	s.EnablePaste()
	s.Clear()

	// One goroutine owns event polling for the whole program; the menu, the
	// history browser, and the game all read from this channel, so leaving
	// one screen for another never loses or steals events.
	events := make(chan tcell.Event, 8)
	quit := make(chan struct{})
	defer close(quit)
	go s.ChannelEvents(events, quit)

	// Menu ↔ game loop: Esc in a game returns here; Esc on the menu (or
	// Ctrl-C anywhere) leaves the program.
	st := store.Default()
	cfg := st.LoadConfig()
	ui.SetChartOptions(chartOptionsFor(cfg.ChartStyle()))
	for {
		sel, err := runMenu(s, events, &cfg, st)
		if err != nil {
			return // menu cancelled; the deferred Fini restores the terminal
		}

		// Settings save on change; only the theme needs saving here.
		cfg.Theme = sel.themeName
		st.SaveConfig(cfg)

		// The words source replaces the sentence pool with a generated stream:
		// a fixed 25-word target in Normal mode, and a buffer the countdown can
		// never outrun in Time Attack.
		sentences := sel.lang.sentences
		langCode := sel.lang.code
		var targetGen func() string
		if sel.src.code == "words" {
			langCode = "en"
			sentences = nil
			count := normalWordCount
			if sel.limit > 0 {
				count = streamWordCount
			}
			targetGen = func() string { return domain.RandomWords(count) }
		}

		meta := app.RoundMeta{Theme: sel.themeName, Lang: langCode, Source: sel.src.code}
		game, err := app.NewGame(s, sel.theme, sel.limit, sentences, targetGen, meta, st, sel.graphOn)
		if err != nil {
			return
		}
		if game.Run(events) {
			return
		}
	}
}
