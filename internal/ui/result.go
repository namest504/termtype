package ui

import (
	"fmt"

	"github.com/namest504/termtype/internal/domain"
)

// ResultText returns the WPM/accuracy/time summary string shown on the
// completion screen. Centralized here so all themes use the same format,
// matching the tone of the graph view's summary line.
func ResultText(gs *domain.GameState) string {
	sep := Glyphs().Sep
	return fmt.Sprintf("%.0f wpm %s %.1f%% acc %s %.0fs", gs.WPM, sep, gs.Accuracy, sep, gs.FinalDurS)
}
