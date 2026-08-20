// Package themes implements the typing screens: each theme renders the
// domain.GameState in its own visual style and registers itself in Themes.
package themes

import "github.com/namest504/termtype/internal/domain"

// Themes is the map storing all themes registered in the program.
var Themes = make(map[string]domain.Theme)
