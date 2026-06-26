package themes

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/gdamore/tcell/v2"
	"termtype/internal/domain"
	"termtype/internal/ui"
)

func init() {
	Themes["claude"] = &ClaudeTheme{}
}

// ClaudeTheme is a theme that looks like composing a message in a claude-code CLI session.
// There's an input box at the bottom with a faint conversation transcript above it. It lays
// out from the bottom up to fit any terminal size.
type ClaudeTheme struct{}

type ClaudeThemeState struct {
	tick int
}

// Offset where text starts inside the input box (border + "> ").
const claudePromptWidth = 3

var claudeSpinner = []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")

// Fake conversation transcript for ambiance (shown faint at the top).
var claudeTranscript = []string{
	"> refactor the parser and add tests",
	"⏺ Done. Updated parser.go and added 12 tests — all passing.",
	"  ⎿ parser.go (+48 -10)   parser_test.go (+96)",
}

func (t *ClaudeTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()
	gs.TargetSentence = domain.Sentences[rand.Intn(len(domain.Sentences))]
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
	orange := tcell.StyleDefault.Foreground(tcell.ColorOrange)

	// Compute input box size. wrapWidth matches the TypingRenderer's internal usable width.
	wrapWidth := (w - 1) - claudePromptWidth - 3
	if wrapWidth < 1 {
		wrapWidth = 1
	}
	wrapped := ui.WrapText(gs.TargetSentence, wrapWidth)
	boxRows := len(wrapped)
	if boxRows < 1 {
		boxRows = 1
	}

	// Layout from the bottom up: hint(h-1) → box bottom(h-2) → content → box top → status row → transcript.
	hintRow := h - 1
	boxBottom := h - 2
	boxTop := boxBottom - (boxRows + 1)
	statusRow := boxTop - 1

	t.drawTranscript(renderer, statusRow, w, faint)
	t.drawStatus(renderer, gs, st, statusRow, orange)
	t.drawBox(renderer, boxTop, boxBottom, w, dim)
	t.drawInput(renderer, gs, wrapped, boxTop, w, orange, dim)
	t.drawHint(renderer, gs, hintRow, w, dim, faint)

	renderer.Show()
}

func (t *ClaudeTheme) drawTranscript(renderer domain.Renderer, statusRow, w int, style tcell.Style) {
	n := len(claudeTranscript)
	for i, line := range claudeTranscript {
		y := statusRow - n + i
		if y >= 0 && y < statusRow {
			renderer.DrawText(0, y, style, ui.Truncate(line, w))
		}
	}
}

func (t *ClaudeTheme) drawStatus(renderer domain.Renderer, gs *domain.GameState, st *ClaudeThemeState, statusRow int, style tcell.Style) {
	if statusRow < 0 {
		return
	}
	if gs.IsFinished {
		renderer.DrawText(0, statusRow, style, "⏺ Message sent")
		return
	}
	frame := claudeSpinner[st.tick%len(claudeSpinner)]
	renderer.DrawText(0, statusRow, style, fmt.Sprintf("%c Composing message…  (esc to interrupt)", frame))
}

func (t *ClaudeTheme) drawBox(renderer domain.Renderer, top, bottom, w int, style tcell.Style) {
	if w < 2 {
		return
	}
	mid := strings.Repeat("─", w-2)
	if top >= 0 {
		renderer.DrawText(0, top, style, "╭"+mid+"╮")
	}
	for y := top + 1; y < bottom; y++ {
		if y >= 0 {
			renderer.SetContent(0, y, '│', style)
			renderer.SetContent(w-1, y, '│', style)
		}
	}
	if bottom >= 0 {
		renderer.DrawText(0, bottom, style, "╰"+mid+"╯")
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
	// Finished: show the sent message in green
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
	if gs.IsFinished {
		result := ui.ResultText(gs) + "   ⏎ Enter for a new message · esc quit"
		renderer.DrawText(0, hintRow, dim, ui.Truncate(result, w))
		return
	}
	renderer.DrawText(0, hintRow, faint, ui.Truncate("⏎ send when complete    esc quit", w))
}
