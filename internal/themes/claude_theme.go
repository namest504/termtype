package themes

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/ui"
)

func init() {
	Themes["claude"] = &ClaudeTheme{}
}

// ClaudeTheme renders the session like a live claude-code conversation: an input
// box at the bottom where you compose your next message, a scrollback of earlier
// turns above it, and a most-recent turn that streams in (tool calls + reply)
// while a spinner/token counter ticks. It lays out from the bottom up to fit any
// terminal size.
type ClaudeTheme struct{}

// ClaudeThemeState is ClaudeTheme's per-round state: which scenario is
// playing and how many of its lines have been revealed so far.
type ClaudeThemeState struct {
	tick int
	scen int // index into claudeScenarios for this round
}

const claudePromptWidth = 3

// How many ticks before each new line of the active turn is revealed.
const claudeRevealEvery = 2

type cKind int

const (
	cHuman cKind = iota
	cAsst
	cTool
)

type cLine struct {
	kind cKind
	text string
}

// claudeScenario is one session's conversation: completed earlier turns in
// the scrollback, and a most-recent turn that streams in line by line.
type claudeScenario struct {
	history []cLine
	active  []cLine
}

var claudeScenarios = []claudeScenario{
	{
		history: []cLine{
			{cHuman, "split the monolith handler into smaller functions"},
			{cAsst, "I'll extract request parsing and validation first."},
			{cTool, "Read handler.go (210 lines)"},
			{cTool, "Updated handler.go (+96 -120)"},
			{cAsst, "Done - split into parseRequest, validate, and dispatch."},
		},
		active: []cLine{
			{cAsst, "I'll add graceful shutdown so in-flight work can finish."},
			{cTool, "Read server.go (142 lines)"},
			{cTool, "Updated server.go (+38 -6)"},
			{cTool, "Updated main.go (+9 -1)"},
			{cAsst, "Done - the server now drains connections on SIGTERM."},
		},
	},
	{
		history: []cLine{
			{cHuman, "add retry with backoff to the fetcher"},
			{cAsst, "I'll wrap the client call in a retry loop."},
			{cTool, "Read fetch.go (88 lines)"},
			{cTool, "Updated fetch.go (+41 -7)"},
			{cAsst, "Done - three attempts with jittered backoff."},
		},
		active: []cLine{
			{cAsst, "I'll make the timeout configurable next."},
			{cTool, "Read config.go (54 lines)"},
			{cTool, "Updated config.go (+12 -2)"},
			{cTool, "Updated fetch.go (+6 -3)"},
			{cAsst, "Done - FETCH_TIMEOUT now overrides the default."},
		},
	},
	{
		history: []cLine{
			{cHuman, "why is the cache test flaky"},
			{cAsst, "The TTL assertion races the clock; I'll inject a fake timer."},
			{cTool, "Read cache_test.go (131 lines)"},
			{cTool, "Updated cache_test.go (+18 -9)"},
			{cAsst, "Done - the test drives a fake clock now."},
		},
		active: []cLine{
			{cAsst, "I'll run the suite to confirm the flake is gone."},
			{cTool, "go test ./... (42 tests)"},
			{cTool, "PASS in 3.1s"},
			{cAsst, "All green - want me to check the other suites?"},
		},
	},
}

func claudeToolCount(active []cLine) int {
	n := 0
	for _, l := range active {
		if l.kind == cTool {
			n++
		}
	}
	return n
}

func (t *ClaudeTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()
	gs.TargetSentence = gs.RandomSentence()
	st, ok := gs.CustomState.(*ClaudeThemeState)
	if !ok {
		st = &ClaudeThemeState{}
		gs.CustomState = st
	}
	// A fresh round picks a scenario and replays its stream from the top.
	st.tick = 0
	st.scen = rand.Intn(len(claudeScenarios))
}

func (t *ClaudeTheme) OnTick(gs *domain.GameState) {
	if st, ok := gs.CustomState.(*ClaudeThemeState); ok {
		st.tick++
	}
}

func (t *ClaudeTheme) UpdateScreen(renderer domain.Renderer, gs *domain.GameState) {
	st, ok := gs.CustomState.(*ClaudeThemeState)
	if !ok {
		return
	}
	renderer.Clear()
	w, h := renderer.Size()
	if w < 2 || h < 1 {
		renderer.Show()
		return
	}

	dim := tcell.StyleDefault.Foreground(tcell.ColorGray)
	faint := tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
	white := tcell.StyleDefault.Foreground(tcell.ColorWhite)
	orange := tcell.StyleDefault.Foreground(tcell.ColorOrange)

	wrapWidth := (w - 1) - claudePromptWidth - 3
	if wrapWidth < 1 {
		wrapWidth = 1
	}
	wrapped := ui.WrapText(gs.TargetSentence, wrapWidth)
	boxRows := len(wrapped)
	// Cap the input box height so long targets (the words stream) scroll
	// inside it instead of swallowing the whole screen.
	maxBox := h - 8
	if maxBox > 6 {
		maxBox = 6
	}
	if maxBox < 1 {
		maxBox = 1
	}
	if boxRows > maxBox {
		boxRows = maxBox
	}
	if boxRows < 1 {
		boxRows = 1
	}

	hintRow := h - 1
	boxBottom := h - 2
	boxTop := boxBottom - (boxRows + 1)
	statusRow := boxTop - 1

	sc := claudeScenarios[st.scen%len(claudeScenarios)]

	// How much of the active (in-progress) turn has streamed in so far.
	reveal := len(sc.active)
	if !gs.IsFinished {
		reveal = st.tick / claudeRevealEvery
		if reveal > len(sc.active) {
			reveal = len(sc.active)
		}
	}

	convo := make([]cLine, 0, len(sc.history)+reveal)
	convo = append(convo, sc.history...)
	convo = append(convo, sc.active[:reveal]...)

	// Render the conversation bottom-anchored, ending just above the status row.
	firstRow := statusRow - len(convo)
	for i, ln := range convo {
		y := firstRow + i
		if y >= 0 && y < statusRow {
			t.drawConvoLine(renderer, y, w, ln, white, orange, faint)
		}
	}

	t.drawStatus(renderer, gs, st, sc.active, reveal, statusRow, w, orange, faint)
	t.drawBox(renderer, boxTop, boxBottom, w, dim)
	t.drawInput(renderer, gs, wrapped, boxTop, boxRows, w, orange, dim)
	t.drawHint(renderer, gs, hintRow, w, dim, faint)

	renderer.Show()
}

func (t *ClaudeTheme) drawConvoLine(renderer domain.Renderer, y, w int, ln cLine, white, orange, faint tcell.Style) {
	gl := ui.Glyphs()
	switch ln.kind {
	case cHuman:
		renderer.DrawText(0, y, faint, "> ")
		renderer.DrawText(2, y, white, ui.Truncate(ln.text, w-2))
	case cAsst:
		renderer.DrawText(0, y, orange, gl.AsstDot)
		renderer.DrawText(2, y, white, ui.Truncate(ln.text, w-2))
	case cTool:
		renderer.DrawText(2, y, faint, gl.ToolBranch)
		renderer.DrawText(4, y, faint, ui.Truncate(ln.text, w-4))
	}
}

func (t *ClaudeTheme) drawStatus(renderer domain.Renderer, gs *domain.GameState, st *ClaudeThemeState, active []cLine, reveal, statusRow, w int, orange, faint tcell.Style) {
	if statusRow < 0 {
		return
	}
	gl := ui.Glyphs()
	if gs.IsFinished {
		renderer.DrawText(0, statusRow, orange, ui.Truncate(gl.AsstDot+" Message sent", w))
		return
	}
	if reveal >= len(active) {
		doneAt := len(active) * claudeRevealEvery
		green := tcell.StyleDefault.Foreground(tcell.ColorGreen)
		msg := fmt.Sprintf("%s responded in %ds %s %d edits %s esc to interrupt", gl.Check, doneAt, gl.Sep, claudeToolCount(active), gl.Sep)
		renderer.DrawText(0, statusRow, green, ui.Truncate(msg, w))
		return
	}
	frame := gl.Spinner[st.tick%len(gl.Spinner)]
	tokens := float64(300+st.tick*137) / 1000.0
	msg := fmt.Sprintf("%c Working%s %ds %s %s %.1fk tokens %s esc to interrupt", frame, gl.Ellipsis, st.tick, gl.Sep, gl.Up, tokens, gl.Sep)
	renderer.DrawText(0, statusRow, orange, ui.Truncate(msg, w))
}

func (t *ClaudeTheme) drawBox(renderer domain.Renderer, top, bottom, w int, style tcell.Style) {
	if w < 2 {
		return
	}
	gl := ui.Glyphs()
	mid := strings.Repeat(gl.BoxH, w-2)
	vbar := []rune(gl.BoxV)[0]
	if top >= 0 {
		renderer.DrawText(0, top, style, gl.BoxTL+mid+gl.BoxTR)
	}
	for y := top + 1; y < bottom; y++ {
		if y >= 0 {
			renderer.SetContent(0, y, vbar, style)
			renderer.SetContent(w-1, y, vbar, style)
		}
	}
	if bottom >= 0 {
		renderer.DrawText(0, bottom, style, gl.BoxBL+mid+gl.BoxBR)
	}
}

func (t *ClaudeTheme) drawInput(renderer domain.Renderer, gs *domain.GameState, wrapped []string, boxTop, boxRows, w int, orange, dim tcell.Style) {
	firstLine := boxTop + 1
	if !gs.IsFinished {
		renderer.DrawText(2, firstLine, orange, ">")
		tr := &ui.TypingRenderer{}
		tr.Draw(renderer, gs, ui.TypingRendererOptions{
			StartY:      firstLine,
			Width:       w - 1,
			PrefixWidth: claudePromptWidth,
			CenterText:  false,
			MaxLines:    boxRows,
		})
		return
	}
	// Finished: show the sent message in green — the tail of it, when the
	// message is taller than the box.
	renderer.HideCursor()
	renderer.DrawText(2, firstLine, dim, ">")
	sent := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	shown := wrapped
	if len(shown) > boxRows {
		shown = shown[len(shown)-boxRows:]
	}
	for i, line := range shown {
		if firstLine+i >= 0 {
			renderer.DrawText(4, firstLine+i, sent, strings.TrimRight(line, " "))
		}
	}
}

func (t *ClaudeTheme) drawHint(renderer domain.Renderer, gs *domain.GameState, hintRow, w int, dim, faint tcell.Style) {
	if hintRow < 0 {
		return
	}
	gl := ui.Glyphs()
	if gs.IsFinished {
		result := ui.ResultText(gs) + fmt.Sprintf("   %s Enter for a new message %s esc menu", gl.Send, gl.Sep)
		renderer.DrawText(0, hintRow, dim, ui.Truncate(result, w))
		return
	}
	renderer.DrawText(0, hintRow, faint, ui.Truncate(gl.Send+" send when complete    esc menu", w))
}
