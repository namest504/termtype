package ui

// GlyphSet is the set of decorative symbols the UI uses. Two variants exist: a
// Unicode set (default) and a plain-ASCII set for terminals or fonts that render
// the fancy glyphs as tofu (□) or at the wrong width. The active set is chosen
// once at startup via SetASCII and read everywhere through Glyphs().
type GlyphSet struct {
	Clock      string // time-attack countdown badge
	Pause      string // pause banner marker
	AsstDot    string // claude assistant turn marker
	ToolBranch string // claude tool-result branch (2 cells wide)
	Send       string // "send/enter" hint marker
	Check      string // success marker
	Up         string // upload/token arrow
	Ellipsis   string // trailing "…"
	BoxTL      string
	BoxTR      string
	BoxBL      string
	BoxBR      string
	BoxH       string
	BoxV       string
	Spinner    []rune // animation frames
	ArrowUD    string // menu: cycle up/down
	ArrowLR    string // menu: cycle left/right
	Enter      string // menu: start key
	Sep        string // inline separator between fields
}

var unicodeGlyphs = GlyphSet{
	Clock:      "⏱",
	Pause:      "⏸",
	AsstDot:    "⏺",
	ToolBranch: "⎿ ",
	Send:       "⏎",
	Check:      "✓",
	Up:         "↑",
	Ellipsis:   "…",
	BoxTL:      "╭",
	BoxTR:      "╮",
	BoxBL:      "╰",
	BoxBR:      "╯",
	BoxH:       "─",
	BoxV:       "│",
	Spinner:    []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"),
	ArrowUD:    "↑/↓",
	ArrowLR:    "←/→",
	Enter:      "Enter",
	Sep:        "·",
}

var asciiGlyphs = GlyphSet{
	Clock:      "TIME",
	Pause:      "||",
	AsstDot:    "*",
	ToolBranch: "- ",
	Send:       ">",
	Check:      "OK",
	Up:         "^",
	Ellipsis:   "...",
	BoxTL:      "+",
	BoxTR:      "+",
	BoxBL:      "+",
	BoxBR:      "+",
	BoxH:       "-",
	BoxV:       "|",
	Spinner:    []rune(`|/-\`),
	ArrowUD:    "Up/Dn",
	ArrowLR:    "L/R",
	Enter:      "Enter",
	Sep:        "-",
}

var useASCII bool

// SetASCII selects the plain-ASCII glyph set. Call once at startup.
func SetASCII(b bool) { useASCII = b }

// IsASCII reports whether the ASCII glyph set is active.
func IsASCII() bool { return useASCII }

// Glyphs returns the active glyph set.
func Glyphs() GlyphSet {
	if useASCII {
		return asciiGlyphs
	}
	return unicodeGlyphs
}
