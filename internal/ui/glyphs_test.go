package ui

import "testing"

// SetASCII toggles between the two glyph sets, and the default is Unicode.
func TestGlyphs_Toggle(t *testing.T) {
	if IsASCII() {
		t.Fatal("default glyph mode should be Unicode")
	}
	if got := Glyphs().Clock; got != "⏱" {
		t.Errorf("unicode Clock = %q, want ⏱", got)
	}

	SetASCII(true)
	defer SetASCII(false)

	if !IsASCII() {
		t.Error("IsASCII should be true after SetASCII(true)")
	}
	if got := Glyphs().Clock; got != "TIME" {
		t.Errorf("ascii Clock = %q, want TIME", got)
	}
	// The box vertical must be a single rune so it can be placed with SetContent.
	if r := []rune(Glyphs().BoxV); len(r) != 1 {
		t.Errorf("ascii BoxV must be one rune, got %q", Glyphs().BoxV)
	}
}
