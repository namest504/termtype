package themes

import (
	"strings"
	"testing"

	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/ui"
)

func TestTutorRegistered(t *testing.T) {
	if _, ok := Themes["tutor"]; !ok {
		t.Fatal("tutor theme is not registered")
	}
}

func TestKeyFingerCoversLayout(t *testing.T) {
	for _, row := range kbRows {
		for _, k := range row {
			if _, ok := keyFinger[k]; !ok {
				t.Errorf("key %q has no finger assignment", k)
			}
		}
	}
	if keyFinger[' '] != rThumb {
		t.Error("space should belong to a thumb")
	}
}

func TestNextKeyInfo(t *testing.T) {
	target := []rune("Go, fast!")

	key, shift, ok := nextKeyInfo(target, []rune(""))
	if !ok || key != 'g' || !shift {
		t.Errorf("next of %q = (%q, %v, %v), want shifted g", "G", key, shift, ok)
	}

	key, shift, ok = nextKeyInfo(target, []rune("Go"))
	if !ok || key != ',' || shift {
		t.Errorf("next of %q = (%q, %v, %v), want plain comma", ",", key, shift, ok)
	}

	key, shift, ok = nextKeyInfo(target, []rune("Go, fast"))
	if !ok || key != '1' || !shift {
		t.Errorf("next of %q = (%q, %v, %v), want shifted 1", "!", key, shift, ok)
	}

	if _, _, ok = nextKeyInfo(target, target); ok {
		t.Error("no guidance expected past the end of the target")
	}

	if _, _, ok = nextKeyInfo([]rune("한글"), nil); ok {
		t.Error("no guidance expected for Hangul runes")
	}
}

func TestFingerCellsInsideHandArt(t *testing.T) {
	cases := []struct {
		name  string
		art   []string
		cells [][]handCell
	}{
		{"ascii", handArtASCII, fingerCellsASCII},
		{"unicode", handArtUnicode, fingerCellsUnicode},
	}
	for _, tc := range cases {
		for f, cells := range tc.cells {
			for _, c := range cells {
				if c.dy < 0 || c.dy >= len(tc.art) {
					t.Fatalf("%s: finger %d cell row %d outside the art", tc.name, f, c.dy)
				}
				row := []rune(tc.art[c.dy])
				if c.dx < 0 || c.dx >= len(row) {
					t.Fatalf("%s: finger %d cell col %d outside the art", tc.name, f, c.dx)
				}
				if row[c.dx] != c.r {
					t.Errorf("%s: finger %d cell (%d,%d) glyph %q does not match the art %q",
						tc.name, f, c.dx, c.dy, c.r, row[c.dx])
				}
			}
		}
	}
}

// 결과 화면에도 키보드가 남아 있어야 한다 (충분히 큰 화면에서).
func TestTutorResultKeepsKeyboard(t *testing.T) {
	theme := &TutorTheme{}
	w, h := 100, 40

	gs := &domain.GameState{Sentences: []string{"hello world"}}
	theme.ResetState(gs)
	gs.TargetSentence = "hello world"
	gs.UserInput = "hello world"
	gs.IsFinished = true
	gs.WPM = 61.2
	gs.Accuracy = 97.5
	gs.FinalDurS = 12

	r := newGridRenderer(w, h)
	theme.UpdateScreen(r, gs)

	// The second keyboard row is "qwertyuiop[]\\"; look for that run of
	// keys somewhere on the grid.
	found := false
	for y := 0; y < h; y++ {
		if strings.Contains(string(r.grid[y]), "q") && strings.Contains(string(r.grid[y]), "p") {
			// Confirm it's really the qwerty row, not incidental letters
			// elsewhere: q must appear before p on the same row.
			line := string(r.grid[y])
			if strings.Index(line, "q") < strings.Index(line, "p") {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("finished tutor screen should still show the keyboard's q..p row")
	}

	// The result text should also be present somewhere on screen.
	result := ui.ResultText(gs)
	resultFound := false
	for y := 0; y < h; y++ {
		if strings.Contains(string(r.grid[y]), "wpm") {
			resultFound = true
			break
		}
	}
	if !resultFound {
		t.Errorf("finished tutor screen should show the result text %q", result)
	}
}

// 유니코드 손 아트는 라운드 박스 글리프를 쓴다.
func TestTutorHandsRounded(t *testing.T) {
	theme := &TutorTheme{}
	w, h := 100, 40

	render := func() *gridRenderer {
		gs := &domain.GameState{Sentences: []string{"hello world"}}
		theme.ResetState(gs)
		gs.TargetSentence = "hello world"
		gs.UserInput = ""
		r := newGridRenderer(w, h)
		theme.UpdateScreen(r, gs)
		return r
	}

	orig := ui.IsASCII()
	t.Cleanup(func() { ui.SetASCII(orig) })

	ui.SetASCII(false)
	r := render()
	if !gridContainsRune(r, '╭') {
		t.Error("Unicode mode should draw rounded hand-art glyphs like '╭'")
	}

	ui.SetASCII(true)
	r = render()
	ui.SetASCII(false)
	if gridContainsRune(r, '╭') {
		t.Error("ASCII mode should not draw rounded box glyphs")
	}
}

// 오른손 아트는 왼손을 단순 반전한 것이 아니라, 방향성 있는 박스 글리프도
// 좌우가 뒤바뀐 모양이어야 한다 (예: 아랫변 모서리 ╰...╯, 손가락 윗줄 ╭╮).
func TestTutorRightHandMirrorsGlyphs(t *testing.T) {
	orig := ui.IsASCII()
	t.Cleanup(func() { ui.SetASCII(orig) })
	ui.SetASCII(false)

	theme := &TutorTheme{}
	w, h := 100, 40
	gs := &domain.GameState{Sentences: []string{"hello world"}}
	theme.ResetState(gs)
	gs.TargetSentence = "hello world"
	gs.UserInput = ""
	r := newGridRenderer(w, h)
	theme.UpdateScreen(r, gs)

	gap := 4
	lx := (w - (2*handWidth + gap)) / 2
	rx := lx + handWidth + gap

	// Locate the hand-art bottom row by finding the left hand's (unmirrored)
	// bottom border, which starts at lx.
	wantBottom := "╰─────────╯"
	bottomY := -1
	for y := 0; y < h; y++ {
		if lx+handWidth <= len(r.grid[y]) && string(r.grid[y][lx:lx+handWidth]) == wantBottom {
			bottomY = y
			break
		}
	}
	if bottomY == -1 {
		t.Fatal("could not locate hand-art bottom row")
	}

	if rightBottom := string(r.grid[bottomY][rx : rx+handWidth]); rightBottom != wantBottom {
		t.Errorf("right hand bottom row = %q, want %q (mirrored corners, not %q)",
			rightBottom, wantBottom, "╯─────────╰")
	}

	// The finger-top row is 4 rows above the bottom border in handArtUnicode.
	topY := bottomY - 4
	if topY < 0 {
		t.Fatalf("hand-art top row %d out of range", topY)
	}
	rightTop := string(r.grid[topY][rx : rx+handWidth])
	if !strings.Contains(rightTop, "╭╮╭╮") {
		t.Errorf("right hand top row = %q, want it to contain %q (not the un-mirrored %q)",
			rightTop, "╭╮╭╮", "╮╭╮╭")
	}
}

func gridContainsRune(r *gridRenderer, want rune) bool {
	for y := 0; y < r.h; y++ {
		for x := 0; x < r.w; x++ {
			if r.grid[y][x] == want {
				return true
			}
		}
	}
	return false
}
