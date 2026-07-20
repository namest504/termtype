package ui

import (
	"fmt"

	"github.com/namest504/termtype/internal/domain"
)

// ResultText returns the WPM/accuracy summary string shown on the completion screen.
// Centralized here so all themes use the same format.
func ResultText(gs *domain.GameState) string {
	return fmt.Sprintf("WPM: %.2f | Accuracy: %.2f%%", gs.Wpm, gs.Accuracy)
}
