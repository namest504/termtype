package themes

import "termtype/internal/domain"

// Themes is the map storing all themes registered in the program.
var Themes = make(map[string]domain.Theme)
