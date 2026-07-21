package themes

import (
	"math/rand"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/namest504/termtype/internal/domain"
	"github.com/namest504/termtype/internal/ui"
)

func init() {
	Themes["matrix"] = &MatrixTheme{}
}

// MatrixTheme implements the falling-character effect.
type MatrixTheme struct{}

// Raindrop is the state of each "raindrop" in the Matrix effect.
type Raindrop struct {
	X, Y   int
	Speed  int
	Chars  []rune
	Length int
}

// MatrixThemeState is the overall state of the Matrix theme.
type MatrixThemeState struct {
	drops         []*Raindrop
	width, height int
}

var matrixChars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}|;:'\",./<>?")

func (t *MatrixTheme) ResetState(gs *domain.GameState) {
	gs.ResetCommon()
	gs.TargetSentence = gs.RandomSentence()

	// Create a new MatrixThemeState if one hasn't been initialized.
	if _, ok := gs.CustomState.(*MatrixThemeState); !ok {
		gs.CustomState = &MatrixThemeState{}
	}
}

func (t *MatrixTheme) UpdateScreen(renderer domain.Renderer, gs *domain.GameState) {
	matrixState, ok := gs.CustomState.(*MatrixThemeState)
	if !ok {
		return // state is not ready yet
	}

	renderer.Clear()
	w, h := renderer.Size()

	t.initializeRaindrops(matrixState, w, h)
	t.drawRaindrops(renderer, matrixState, w, h)

	// Starting Y coordinate for drawing the text.
	startY := h/2 - 2

	if !gs.IsFinished {
		t.drawTypingArea(renderer, gs, w, h, startY)
	} else {
		t.drawResultArea(renderer, gs, w, h, startY)
	}

	renderer.Show()
}

func (t *MatrixTheme) initializeRaindrops(matrixState *MatrixThemeState, w, h int) {
	if matrixState.width != w || matrixState.height != h {
		matrixState.width = w
		matrixState.height = h
		matrixState.drops = make([]*Raindrop, w)
		for i := 0; i < w; i++ {
			matrixState.drops[i] = &Raindrop{X: i, Y: rand.Intn(h), Speed: rand.Intn(4) + 1, Length: rand.Intn(10) + 5}
		}
	}
}

func (t *MatrixTheme) drawRaindrops(renderer domain.Renderer, matrixState *MatrixThemeState, w, h int) {
	// Fill the background with black.
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			renderer.SetContent(x, y, ' ', tcell.StyleDefault.Background(tcell.ColorBlack))
		}
	}

	// Draw the raindrops.
	for _, drop := range matrixState.drops {
		for i := 0; i < drop.Length; i++ {
			y := drop.Y - i
			if y >= 0 && y < h {
				style := tcell.StyleDefault.Foreground(tcell.ColorGreen)
				if i == 0 {
					style = tcell.StyleDefault.Foreground(tcell.ColorWhite)
				} else if i > drop.Length-3 {
					style = tcell.StyleDefault.Foreground(tcell.ColorDarkGreen)
				}
				if len(drop.Chars) > 0 {
					renderer.SetContent(drop.X, y, drop.Chars[i%len(drop.Chars)], style)
				}
			}
		}
	}
}

// matrixMaxLines bounds the typing window so long targets (the words
// stream) scroll instead of running off the bottom of the screen.
func matrixMaxLines(h, startY int) int {
	maxLines := h - startY - 2
	if maxLines > 7 {
		maxLines = 7
	}
	if maxLines < 1 {
		maxLines = 1
	}
	return maxLines
}

func (t *MatrixTheme) drawTypingArea(renderer domain.Renderer, gs *domain.GameState, w, h, startY int) {
	tr := &ui.TypingRenderer{}
	tr.Draw(renderer, gs, ui.TypingRendererOptions{
		StartY:      startY,
		Width:       w - 4, // 2 cells of padding on each side
		PrefixWidth: 0,
		CenterText:  true,
		MaxLines:    matrixMaxLines(h, startY),
	})
}

func (t *MatrixTheme) drawResultArea(renderer domain.Renderer, gs *domain.GameState, w, h, startY int) {
	// Recompute how many rows the typing window occupied.
	rows := len(ui.WrapText(gs.TargetSentence, w-4))
	if cap := matrixMaxLines(h, startY); rows > cap {
		rows = cap
	}

	renderer.HideCursor()
	resultText := ui.ResultText(gs)
	x := (w - runewidth.StringWidth(resultText)) / 2
	if x < 0 {
		x = 0
	}
	renderer.DrawText(x, startY+rows+1, tcell.StyleDefault.Background(tcell.ColorBlack), resultText)
}

func (t *MatrixTheme) OnTick(gs *domain.GameState) {
	matrixState, ok := gs.CustomState.(*MatrixThemeState)
	if !ok || matrixState.drops == nil {
		return
	}

	for _, drop := range matrixState.drops {
		drop.Y += drop.Speed
		if drop.Y-drop.Length > matrixState.height {
			drop.Y = 0
		}
		// Keep changing the raindrop's character content.
		drop.Chars = append(drop.Chars, matrixChars[rand.Intn(len(matrixChars))])
		if len(drop.Chars) > 50 {
			drop.Chars = drop.Chars[1:]
		}
	}
}
