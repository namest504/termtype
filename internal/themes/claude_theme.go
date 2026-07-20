package themes

import (
	"fmt"
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

type ClaudeThemeState struct {
	tick int
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

// Completed earlier turns shown faint in the scrollback.
var claudeHistory = []cLine{
	{cHuman, "split the monolith handler into smaller functions"},
	{cAsst, "I'll extract request parsing and validation first."},
	{cTool, "Read handler.go (210 lines)"},
	{cTool, "Updated handler.go (+96 -120)"},
	{cAsst, "Done - split into parseRequest, validate, and dispatch."},
}

// The most recent turn, revealed line by line over time (streaming).
var claudeActive = []cLine{
	{cAsst, "I'll add graceful shutdown so in-flight work can finish."},
	{cTool, "Read server.go (142 lines)"},
	{cTool, "Updated server.go (+38 -6)"},
	{cTool, "Updated main.go (+9 -1)"},
	{cAsst, "Done - the server now drains connections on SIGTERM."},
}

func claudeToolCount() int {
	n := 0
	for _, l := range claudeActive {
		if l.kind == cTool {
			n++
		}
	}
	return n
}

func (t *ClaudeTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()
	gs.TargetSentence = gs.RandomSentence()
	if _, ok := gs.CustomState.(*ClaudeThemeState); !ok {
		gs.CustomState = &ClaudeThemeState{}
	}
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
	if boxRows < 1 {
		boxRows = 1
	}

	hintRow := h - 1
	boxBottom := h - 2
	boxTop := boxBottom - (boxRows + 1)
	statusRow := boxTop - 1

	// How much of the active (in-progress) turn has streamed in so far.
	reveal := len(claudeActive)
	if !gs.IsFinished {
		reveal = st.tick / claudeRevealEvery
		if reveal > len(claudeActive) {
			reveal = len(claudeActive)
		}
	}

	convo := make([]cLine, 0, len(claudeHistory)+reveal)
	convo = append(convo, claudeHistory...)
	convo = append(convo, claudeActive[:reveal]...)

	// Render the conversation bottom-anchored, ending just above the status row.
	firstRow := statusRow - len(convo)
	for i, ln := range convo {
		y := firstRow + i
		if y >= 0 && y < statusRow {
			t.drawConvoLine(renderer, y, w, ln, white, orange, faint)
		}
	}

	t.drawStatus(renderer, gs, st, reveal, statusRow, w, orange, faint)
	t.drawBox(renderer, boxTop, boxBottom, w, dim)
	t.drawInput(renderer, gs, wrapped, boxTop, w, orange, dim)
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

func (t *ClaudeTheme) drawStatus(renderer domain.Renderer, gs *domain.GameState, st *ClaudeThemeState, reveal, statusRow, w int, orange, faint tcell.Style) {
	if statusRow < 0 {
		return
	}
	gl := ui.Glyphs()
	if gs.IsFinished {
		renderer.DrawText(0, statusRow, orange, ui.Truncate(gl.AsstDot+" Message sent", w))
		return
	}
	if reveal >= len(claudeActive) {
		doneAt := len(claudeActive) * claudeRevealEvery
		green := tcell.StyleDefault.Foreground(tcell.ColorGreen)
		msg := fmt.Sprintf("%s responded in %ds %s %d edits %s esc to interrupt", gl.Check, doneAt, gl.Sep, claudeToolCount(), gl.Sep)
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

func (t *ClaudeTheme) drawInput(renderer domain.Renderer, gs *domain.GameState, wrapped []string, boxTop, w int, orange, dim tcell.Style) {
	firstLine := boxTop + 1
	if !gs.IsFinished {
		renderer.DrawText(2, firstLine, orange, ">")
		tr := &ui.TypingRenderer{}
		tr.Draw(renderer, gs, ui.TypingRendererOptions{
			StartY:      firstLine,
			Width:       w - 1,
			PrefixWidth: claudePromptWidth,
			CenterText:  false,
		})
		return
	}
	// Finished: show the sent message in green.
	renderer.HideCursor()
	renderer.DrawText(2, firstLine, dim, ">")
	sent := tcell.StyleDefault.Foreground(tcell.ColorGreen)
	for i, line := range wrapped {
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
		result := ui.ResultText(gs) + fmt.Sprintf("   %s Enter for a new message %s esc quit", gl.Send, gl.Sep)
		renderer.DrawText(0, hintRow, dim, ui.Truncate(result, w))
		return
	}
	renderer.DrawText(0, hintRow, faint, ui.Truncate(gl.Send+" send when complete    esc quit", w))
}
