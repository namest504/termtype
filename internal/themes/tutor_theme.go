package themes

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/ui"
)

func init() {
	Themes["tutor"] = &TutorTheme{}
}

// TutorTheme is a touch-typing tutor screen: the target text above a QWERTY
// keyboard colored by finger zone, with the next key highlighted, and a pair
// of ASCII hands whose finger for that key lights up. When the terminal is
// short the hands fold away first, then the keyboard.
type TutorTheme struct{}

// Fingers are numbered 0-9: left pinky..left thumb, right thumb..right pinky.
const (
	lPinky = iota
	lRing
	lMiddle
	lIndex
	lThumb
	rThumb
	rIndex
	rMiddle
	rRing
	rPinky
)

// keyFinger maps each unshifted key to the finger that presses it.
var keyFinger = map[rune]int{
	'`': lPinky, '1': lPinky, 'q': lPinky, 'a': lPinky, 'z': lPinky,
	'2': lRing, 'w': lRing, 's': lRing, 'x': lRing,
	'3': lMiddle, 'e': lMiddle, 'd': lMiddle, 'c': lMiddle,
	'4': lIndex, '5': lIndex, 'r': lIndex, 't': lIndex,
	'f': lIndex, 'g': lIndex, 'v': lIndex, 'b': lIndex,
	'6': rIndex, '7': rIndex, 'y': rIndex, 'u': rIndex,
	'h': rIndex, 'j': rIndex, 'n': rIndex, 'm': rIndex,
	'8': rMiddle, 'i': rMiddle, 'k': rMiddle, ',': rMiddle,
	'9': rRing, 'o': rRing, 'l': rRing, '.': rRing,
	'0': rPinky, '-': rPinky, '=': rPinky, 'p': rPinky,
	'[': rPinky, ']': rPinky, '\\': rPinky, ';': rPinky,
	'\'': rPinky, '/': rPinky,
	' ': rThumb,
}

// shiftedKeys maps a shifted symbol to the key that produces it.
var shiftedKeys = map[rune]rune{
	'!': '1', '@': '2', '#': '3', '$': '4', '%': '5', '^': '6',
	'&': '7', '*': '8', '(': '9', ')': '0', '_': '-', '+': '=',
	'{': '[', '}': ']', '|': '\\', ':': ';', '"': '\'',
	'<': ',', '>': '.', '?': '/', '~': '`',
}

// nextKeyInfo resolves the next rune to type into the physical key and
// whether shift is needed. ok is false when there is no guidance for it
// (end of the target, or a rune outside the US layout — e.g. Hangul).
func nextKeyInfo(target, input []rune) (key rune, shift, ok bool) {
	if len(input) >= len(target) {
		return 0, false, false
	}
	r := target[len(input)]
	if r >= 'A' && r <= 'Z' {
		return unicode.ToLower(r), true, true
	}
	if base, found := shiftedKeys[r]; found {
		return base, true, true
	}
	if _, found := keyFinger[r]; found {
		return r, false, true
	}
	return 0, false, false
}

// kbRows is the physical layout; row i is indented i keycap-halves.
var kbRows = []string{
	"`1234567890-=",
	"qwertyuiop[]\\",
	"asdfghjkl;'",
	"zxcvbnm,./",
}

const (
	kbKeycap  = 2  // cells per keycap
	kbWidth   = 30 // widest row (13 caps + indent + shift caps)
	kbRowsH   = 5  // 4 rows + space bar
	handsH    = 5
	handWidth = 11
)

// fingerColor is the zone color per finger; mirrored on both hands.
func fingerColor(f int) tcell.Color {
	switch f {
	case lPinky, rPinky:
		return tcell.ColorPurple
	case lRing, rRing:
		return tcell.ColorBlue
	case lMiddle, rMiddle:
		return tcell.ColorGreen
	case lIndex, rIndex:
		return tcell.ColorOrange
	default: // thumbs
		return tcell.ColorGray
	}
}

func (t *TutorTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()
	gs.TargetSentence = gs.RandomSentence()
}

func (t *TutorTheme) OnTick(gs *domain.GameState) {}

func (t *TutorTheme) UpdateScreen(renderer domain.Renderer, gs *domain.GameState) {
	renderer.Clear()
	w, h := renderer.Size()

	const textLines = 2
	showKB := h >= textLines+kbRowsH+4 && w >= kbWidth+2
	showHands := showKB && h >= textLines+kbRowsH+handsH+6 && w >= 2*handWidth+6

	// Center the whole ensemble vertically; the text column shares the
	// keyboard's center and sits one row above it, so the eye never has to
	// travel between them.
	blockH := textLines
	if showKB {
		blockH += 1 + kbRowsH
	}
	if showHands {
		blockH += 1 + handsH
	}
	top := (h - blockH) / 2
	if top < 1 {
		top = 1
	}

	colW := kbWidth + 12
	if colW > w-2 {
		colW = w - 2
	}
	colX := (w - colW) / 2
	if colX < 1 {
		colX = 1
	}

	tr := &ui.TypingRenderer{}
	rows := tr.Draw(renderer, gs, ui.TypingRendererOptions{
		StartY:      top,
		Width:       colX + colW + 2, // renderer draws from colX via PrefixWidth
		PrefixWidth: colX - 1,
		MaxLines:    textLines,
	})

	if gs.IsFinished {
		renderer.HideCursor()
		renderer.DrawText(colX, top+rows+1, tcell.StyleDefault.Foreground(tcell.ColorWhite), ui.Truncate(ui.ResultText(gs), colW))
		renderer.Show()
		return
	}

	key, shift, ok := nextKeyInfo([]rune(gs.TargetSentence), []rune(gs.UserInput))

	kbTop := top + rows + 1
	if showKB {
		t.drawKeyboard(renderer, (w-kbWidth)/2, kbTop, key, shift, ok)
	}
	if showHands {
		finger := -1
		if ok {
			finger = keyFinger[key]
		}
		t.drawHands(renderer, w, kbTop+kbRowsH+1, finger)
	}
	renderer.Show()
}

// drawKeyboard renders the zone-colored layout. The next key is reversed;
// when shift is needed, the shift cap opposite the key's hand is reversed
// too, touch-typing style.
func (t *TutorTheme) drawKeyboard(renderer domain.Renderer, x, y int, key rune, shift, ok bool) {
	for row, keys := range kbRows {
		cx := x + row // stagger
		if row == 3 {
			// Left shift cap sits before the bottom row.
			t.drawCap(renderer, cx-1, y+row, '^', tcell.ColorGray, ok && shift && keyFinger[key] >= rThumb)
			cx += kbKeycap
		}
		for _, k := range keys {
			f := keyFinger[k]
			t.drawCap(renderer, cx, y+row, k, fingerColor(f), ok && k == key)
			cx += kbKeycap
		}
		if row == 3 {
			// Right shift cap after the bottom row.
			t.drawCap(renderer, cx+1, y+row, '^', tcell.ColorGray, ok && shift && keyFinger[key] < lThumb)
		}
	}
	// Space bar under the home row, thumb-colored.
	spaceStyle := tcell.StyleDefault.Foreground(fingerColor(lThumb))
	if ok && key == ' ' {
		spaceStyle = spaceStyle.Reverse(true).Bold(true)
	}
	for i := 0; i < 8*kbKeycap; i++ {
		renderer.SetContent(x+7+i, y+4, '_', spaceStyle)
	}
}

func (t *TutorTheme) drawCap(renderer domain.Renderer, x, y int, k rune, color tcell.Color, active bool) {
	style := tcell.StyleDefault.Foreground(color)
	if active {
		style = style.Reverse(true).Bold(true)
	}
	renderer.SetContent(x, y, k, style)
}

// handArt is one hand, thumb-side blank; the right hand mirrors it. Fingers
// occupy fixed columns so the active one can be recolored.
var handArt = []string{
	"  _ _ _ _  ",
	" | | | | | ",
	" | | | | |_",
	" |        |",
	" |________|",
}

// handCell is one cell of a finger's highlight: its position in the (left)
// hand art and the glyph drawn there.
type handCell struct {
	dx, dy int
	r      rune
}

// fingerCells maps art fingers 0..3 (pinky..index on the left hand, index..
// pinky mirrored on the right) plus 4 (thumb) to their highlight cells.
var fingerCells = [][]handCell{
	{{2, 0, '_'}, {1, 1, '|'}, {3, 1, '|'}, {1, 2, '|'}, {3, 2, '|'}}, // pinky
	{{4, 0, '_'}, {3, 1, '|'}, {5, 1, '|'}, {3, 2, '|'}, {5, 2, '|'}}, // ring
	{{6, 0, '_'}, {5, 1, '|'}, {7, 1, '|'}, {5, 2, '|'}, {7, 2, '|'}}, // middle
	{{8, 0, '_'}, {7, 1, '|'}, {9, 1, '|'}, {7, 2, '|'}, {9, 2, '|'}}, // index
	{{10, 2, '_'}, {10, 3, '|'}},                                      // thumb hook
}

func (t *TutorTheme) drawHands(renderer domain.Renderer, w, y, finger int) {
	dim := tcell.StyleDefault.Foreground(tcell.ColorDimGray)
	accent := tcell.StyleDefault.Foreground(tcell.ColorOrange).Bold(true)

	gap := 4
	lx := (w - (2*handWidth + gap)) / 2
	rx := lx + handWidth + gap

	// Base art: the right hand is the left mirrored.
	for row, line := range handArt {
		runes := []rune(line)
		for col, r := range runes {
			if r != ' ' {
				renderer.SetContent(lx+col, y+row, r, dim)
			}
		}
		for col, r := range mirrorRunes(runes) {
			if r != ' ' {
				renderer.SetContent(rx+col, y+row, r, dim)
			}
		}
	}

	if finger < 0 {
		return
	}

	// Recolor the active finger's cells ('_' and '|' mirror onto themselves).
	if finger <= lThumb { // left hand: finger index matches art order
		for _, c := range fingerCells[finger] {
			renderer.SetContent(lx+c.dx, y+c.dy, c.r, accent)
		}
	} else { // right hand: mirror the column
		art := rPinky - finger // 9→0 pinky .. 6→3 index
		if finger == rThumb {
			art = 4
		}
		for _, c := range fingerCells[art] {
			renderer.SetContent(rx+(handWidth-1-c.dx), y+c.dy, c.r, accent)
		}
	}
}

// mirrorRunes flips a hand row horizontally, swapping the glyphs that have
// a direction.
func mirrorRunes(runes []rune) []rune {
	out := make([]rune, len(runes))
	for i, r := range runes {
		out[len(runes)-1-i] = r
	}
	return out
}
