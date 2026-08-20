package chart

// Sparkline level glyphs, lowest to highest. The ASCII set is for terminals
// that render block elements as tofu.
var (
	sparkUnicode = []rune("▁▂▃▄▅▆▇█")
	sparkASCII   = []rune(".:-=+*#%")
)

// Sparkline renders values as one glyph per value, min–max scaled to the
// glyph range. Equal values render at a middle level so a flat series stays
// visible. An empty input yields an empty string.
func Sparkline(values []float64, ascii bool) string {
	if len(values) == 0 {
		return ""
	}
	glyphs := sparkUnicode
	if ascii {
		glyphs = sparkASCII
	}
	lo, hi := bounds(values)
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
