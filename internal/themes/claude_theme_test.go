package themes

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/namest504/termtype/internal/domain"
)

// 시나리오 데이터가 최신 CLI 형식을 따르는지 — 렌더 없이 데이터로 검증
func TestClaudeScenarios_ModernFormat(t *testing.T) {
	for i, sc := range claudeScenarios {
		lines := append(append([]cLine{}, sc.history...), sc.active...)
		var calls, results int
		for _, ln := range lines {
			switch ln.kind {
			case cToolCall:
				calls++
				// "Update Todos" is a real no-argument tool call in the CLI and is
				// exempt from the Name(args) shape.
				if ln.text != "Update Todos" && (!strings.Contains(ln.text, "(") || !strings.HasSuffix(ln.text, ")")) {
					t.Errorf("scenario %d: tool call %q must be Name(args)", i, ln.text)
				}
			case cToolRes:
				results++
			}
		}
		if calls == 0 || results == 0 {
			t.Errorf("scenario %d: needs tool calls and results", i)
		}
	}
}

// 한 시나리오에는 Todo 장면이 있어야 한다
func TestClaudeScenarios_HasTodoScene(t *testing.T) {
	for _, sc := range claudeScenarios {
		for _, ln := range append(append([]cLine{}, sc.history...), sc.active...) {
			if ln.kind == cToolCall && strings.Contains(ln.text, "Update Todos") {
				return
			}
		}
	}
	t.Fatal("no scenario contains an Update Todos scene")
}

// The scenario set as a whole should be textured with breathing room, diff
// hunks, and output dump lines — not just an unbroken stream of tool calls.
func TestClaudeScenarios_HasTexture(t *testing.T) {
	var blanks, dels, adds, outs int
	for _, sc := range claudeScenarios {
		lines := append(append([]cLine{}, sc.history...), sc.active...)
		for _, ln := range lines {
			switch ln.kind {
			case cBlank:
				blanks++
			case cDiffDel:
				dels++
			case cDiffAdd:
				adds++
			case cOut:
				outs++
			}
		}
	}
	if blanks == 0 {
		t.Error("scenarios: need at least one cBlank")
	}
	if dels == 0 || adds == 0 {
		t.Error("scenarios: need at least one cDiffDel+cDiffAdd pair")
	}
	if outs == 0 {
		t.Error("scenarios: need at least one cOut line")
	}
}

// recordingRenderer captures drawn text so render output can be inspected.
type recordingRenderer struct {
	w, h int
	rows []string
}

func newRecordingRenderer(w, h int) *recordingRenderer {
	rows := make([]string, h)
	for i := range rows {
		rows[i] = strings.Repeat(" ", w)
	}
	return &recordingRenderer{w: w, h: h, rows: rows}
}

func (r *recordingRenderer) DrawText(x, y int, style tcell.Style, text string) {
	if y < 0 || y >= r.h {
		return
	}
	rw := []rune(r.rows[y])
	for _, ch := range text {
		if x < 0 || x >= len(rw) {
			x += runewidth.RuneWidth(ch)
			continue
		}
		rw[x] = ch
		x += runewidth.RuneWidth(ch)
	}
	r.rows[y] = string(rw)
}
func (r *recordingRenderer) DrawRune(x, y int, ru rune, style tcell.Style) int {
	r.DrawText(x, y, style, string(ru))
	return runewidth.RuneWidth(ru)
}
func (r *recordingRenderer) Clear()           {}
func (r *recordingRenderer) Show()            {}
func (r *recordingRenderer) Size() (int, int) { return r.w, r.h }
func (r *recordingRenderer) SetContent(x, y int, ru rune, style tcell.Style) {
	r.DrawRune(x, y, ru, style)
}
func (r *recordingRenderer) HideCursor()         {}
func (r *recordingRenderer) ShowCursor(x, y int) {}

func (r *recordingRenderer) dump() string {
	return strings.Join(r.rows, "\n")
}

// Render smoke test: forcing the scenario to fully reveal should surface the
// diff "- " marker text and a tail-style cOut line in the drawn output.
func TestClaudeTheme_RenderShowsDiffAndOutput(t *testing.T) {
	theme := &ClaudeTheme{}
	for i := range claudeScenarios {
		gs := &domain.GameState{Sentences: domain.Sentences}
		theme.ResetState(gs)
		st := gs.CustomState.(*ClaudeThemeState)
		st.scen = i
		// Tick enough to reveal every line of the active turn.
		sc := claudeScenarios[i]
		for st.tick < (len(sc.active)+1)*claudeRevealEvery {
			theme.OnTick(gs)
		}
		r := newRecordingRenderer(120, 40)
		theme.UpdateScreen(r, gs)
		out := r.dump()

		wantDiff, wantOut := false, false
		for _, ln := range append(append([]cLine{}, sc.history...), sc.active...) {
			if ln.kind == cDiffDel || ln.kind == cDiffAdd {
				wantDiff = true
			}
			if ln.kind == cOut {
				wantOut = true
			}
		}
		if wantDiff && !strings.Contains(out, "- ") && !strings.Contains(out, "+ ") {
			t.Errorf("scenario %d: expected diff marker in render output", i)
		}
		if wantOut {
			// At least one cOut line's text should appear verbatim somewhere.
			found := false
			for _, ln := range append(append([]cLine{}, sc.history...), sc.active...) {
				if ln.kind == cOut && strings.Contains(out, ln.text) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("scenario %d: expected an output-dump line in render output", i)
			}
		}
	}
}
