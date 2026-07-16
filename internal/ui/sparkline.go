package ui

// Sparkline level glyphs, lowest to highest. The ASCII set mirrors the
// GlyphSet fallback idea for terminals that render block elements as tofu.
var (
	sparkUnicode = []rune("▁▂▃▄▅▆▇█")
	sparkASCII   = []rune(".:-=+*#%")
)

// Sparkline renders values as one glyph per value, min–max scaled to the
// glyph range. Equal values render at a middle level so a flat series stays
// visible. An empty input yields an empty string.
func Sparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}
	glyphs := sparkUnicode
	if IsASCII() {
		glyphs = sparkASCII
	}
	lo, hi := values[0], values[0]
	for _, v := range values[1:] {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	out := make([]rune, len(values))
	for i, v := range values {
		level := len(glyphs) / 2
		if hi > lo {
			level = int((v - lo) / (hi - lo) * float64(len(glyphs)-1))
		}
		out[i] = glyphs[level]
	}
	return string(out)
}
