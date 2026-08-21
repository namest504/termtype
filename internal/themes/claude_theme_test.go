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

// A tool's result block (the ⎿ line plus its diff/output lines) must reveal
// as one atomic group, not line by line: in scenario 1's active turn, the
// group after "Update(cache_test.go)" must bundle the "Updated..." result
// line together with all 3 diff lines.
func TestClaudeGroupLines_ResultBlockIsAtomic(t *testing.T) {
	active := claudeScenarios[0].active
	groups := claudeGroupLines(active)

	// active[0] = Update(cache_test.go) tool call -> its own group.
	// active[1..4] = toolRes + 3 diff lines -> one atomic group.
	if len(groups) < 2 {
		t.Fatalf("expected at least 2 groups, got %d", len(groups))
	}
	if len(groups[0]) != 1 || groups[0][0].kind != cToolCall {
		t.Fatalf("group 0: want [cToolCall], got %v", groups[0])
	}
	want := []cKind{cToolRes, cDiffDel, cDiffAdd, cDiffAdd}
	if len(groups[1]) != len(want) {
		t.Fatalf("group 1: want %d lines (result + 3 diffs), got %d: %v", len(want), len(groups[1]), groups[1])
	}
	for i, k := range want {
		if groups[1][i].kind != k {
			t.Errorf("group 1 line %d: want kind %v, got %v", i, k, groups[1][i].kind)
		}
	}

	// Reveal step-by-step: revealing through group 0 shows only the tool
	// call; the NEXT step must reveal the result line and all 3 diff lines
	// together (4 more lines at once), not one at a time.
	afterGroup0 := revealLineCount(active, 1)
	afterGroup1 := revealLineCount(active, 2)
	if afterGroup0 != 1 {
		t.Errorf("after group 0: want 1 line revealed, got %d", afterGroup0)
	}
	if got := afterGroup1 - afterGroup0; got != 4 {
		t.Errorf("group 1 step: want 4 lines revealed at once (result + 3 diffs), got %d", got)
	}
}

// A blank line attaches to the group that follows it, so the blank and its
// following block reveal together in a single step.
func TestClaudeGroupLines_BlankAttachesToFollowingGroup(t *testing.T) {
	active := []cLine{
		{cToolCall, "Bash(go test ./...)"},
		{cOut, "ok"},
		{cBlank, ""},
		{cAsst, "Done."},
	}
	groups := claudeGroupLines(active)
	if len(groups) != 3 {
		t.Fatalf("want 3 groups (tool call / result / blank+asst), got %d: %v", len(groups), groups)
	}
	last := groups[2]
	if len(last) != 2 || last[0].kind != cBlank || last[1].kind != cAsst {
		t.Fatalf("last group: want [cBlank, cAsst], got %v", last)
	}
	// The step that reveals the final group must reveal both the blank and
	// the assistant line together, not the blank alone first.
	beforeLast := revealLineCount(active, 2)
	afterLast := revealLineCount(active, 3)
	if got := afterLast - beforeLast; got != 2 {
		t.Errorf("final step: want 2 lines revealed at once (blank + asst), got %d", got)
	}
}

// A trailing blank with nothing after it attaches to the previous group
// instead of forming its own (empty-looking) group.
func TestClaudeGroupLines_TrailingBlankAttachesToPreviousGroup(t *testing.T) {
	active := []cLine{
		{cToolCall, "Read(fetch.go)"},
		{cToolRes, "Read 88 lines"},
		{cBlank, ""},
	}
	groups := claudeGroupLines(active)
	if len(groups) != 2 {
		t.Fatalf("want 2 groups, got %d: %v", len(groups), groups)
	}
	last := groups[1]
	if len(last) != 2 || last[0].kind != cToolRes || last[1].kind != cBlank {
		t.Fatalf("last group: want [cToolRes, cBlank], got %v", last)
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
