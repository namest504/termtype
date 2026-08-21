package themes

import (
	"strings"
	"testing"
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
