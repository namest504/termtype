package themes

import "testing"

func TestTutorRegistered(t *testing.T) {
	if _, ok := Themes["tutor"]; !ok {
		t.Fatal("tutor theme is not registered")
	}
}

func TestKeyFingerCoversLayout(t *testing.T) {
	for _, row := range kbRows {
		for _, k := range row {
			if _, ok := keyFinger[k]; !ok {
				t.Errorf("key %q has no finger assignment", k)
			}
		}
	}
	if keyFinger[' '] != rThumb {
		t.Error("space should belong to a thumb")
	}
}

func TestNextKeyInfo(t *testing.T) {
	target := []rune("Go, fast!")

	key, shift, ok := nextKeyInfo(target, []rune(""))
	if !ok || key != 'g' || !shift {
		t.Errorf("next of %q = (%q, %v, %v), want shifted g", "G", key, shift, ok)
	}

	key, shift, ok = nextKeyInfo(target, []rune("Go"))
	if !ok || key != ',' || shift {
		t.Errorf("next of %q = (%q, %v, %v), want plain comma", ",", key, shift, ok)
	}

	key, shift, ok = nextKeyInfo(target, []rune("Go, fast"))
	if !ok || key != '1' || !shift {
		t.Errorf("next of %q = (%q, %v, %v), want shifted 1", "!", key, shift, ok)
	}

	if _, _, ok = nextKeyInfo(target, target); ok {
		t.Error("no guidance expected past the end of the target")
	}

	if _, _, ok = nextKeyInfo([]rune("한글"), nil); ok {
		t.Error("no guidance expected for Hangul runes")
	}
}

func TestFingerCellsInsideHandArt(t *testing.T) {
	for f, cells := range fingerCells {
		for _, c := range cells {
			if c.dy < 0 || c.dy >= len(handArt) {
				t.Fatalf("finger %d cell row %d outside the art", f, c.dy)
			}
			row := []rune(handArt[c.dy])
			if c.dx < 0 || c.dx >= len(row) {
				t.Fatalf("finger %d cell col %d outside the art", f, c.dx)
			}
			if row[c.dx] != c.r {
				t.Errorf("finger %d cell (%d,%d) glyph %q does not match the art %q",
					f, c.dx, c.dy, c.r, row[c.dx])
			}
		}
	}
}
