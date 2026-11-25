package main

import "testing"

// TestCalcRowWidths checks widths sum to the total and differ by at most 1.
func TestCalcRowWidths(t *testing.T) {
	for _, total := range []int{0, 1, 7, 8, 12, 80, 81, 157} {
		for _, cols := range []int{8, 12} {
			w := calcRowWidths(total, cols)
			if len(w) != cols {
				t.Fatalf("calcRowWidths(%d,%d): len %d", total, cols, len(w))
			}
			want := total
			if total <= 0 {
				want = cols * defaultTileWidth
			}
			sum, lo, hi := 0, w[0], w[0]
			for _, v := range w {
				sum += v
				lo, hi = min(lo, v), max(hi, v)
			}
			if sum != want {
				t.Errorf("calcRowWidths(%d,%d) sums to %d, want %d", total, cols, sum, want)
			}
			if hi-lo > 1 {
				t.Errorf("calcRowWidths(%d,%d) uneven: %v", total, cols, w)
			}
		}
	}
	if calcRowWidths(10, 0) != nil {
		t.Error("cols=0 should return nil")
	}
}

// TestTileAt checks hit-testing against known positions.
func TestTileAt(t *testing.T) {
	rows := paletteRows()
	const w, h = 96, 40 // 12 cols × 8, 8 cols × 12

	tests := []struct {
		name       string
		x, y, scrl int
		want       int
		ok         bool
	}{
		{"title", 5, 0, 0, 0, false},
		{"first system tile", 0, headerLines, 0, 0, true},
		{"last system tile", w - 1, headerLines, 0, 7, true},
		{"second system row", 12, headerLines + tileHeight, 0, 9, true},
		{"spacer", 0, headerLines + 2*tileHeight, 0, 0, false},
		{"first cube tile", 0, headerLines + 3*tileHeight, 0, 16, true},
		{"cube right half", 6 * 8, headerLines + 3*tileHeight, 0, 82, true},
		{"scrolled", 0, headerLines, tileHeight, 8, true},
		{"status line", 0, h - 2, 0, 0, false},
		{"negative x", -1, headerLines, 0, 0, false},
		{"past right edge", w, headerLines, 0, 0, false},
	}
	for _, tt := range tests {
		got, ok := tileAt(rows, tt.x, tt.y, tt.scrl, w, h)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Errorf("%s: tileAt = (%d,%v), want (%d,%v)", tt.name, got, ok, tt.want, tt.ok)
		}
	}
}

// TestTileAtMatchesLayout round-trips every tile's on-screen position.
func TestTileAtMatchesLayout(t *testing.T) {
	rows := paletteRows()
	const w = 101
	h := contentLines(rows) + chromeLines // everything visible
	for r, row := range rows {
		if row.isGap() {
			continue
		}
		x := 0
		for c, width := range calcRowWidths(w, len(row.ids)) {
			for dy := range tileHeight {
				y := headerLines + r*tileHeight + dy
				for _, dx := range []int{0, width - 1} {
					got, ok := tileAt(rows, x+dx, y, 0, w, h)
					if !ok || got != row.ids[c] {
						t.Fatalf("row %d col %d: tileAt(%d,%d) = (%d,%v), want %d", r, c, x+dx, y, got, ok, row.ids[c])
					}
				}
			}
			x += width
		}
	}
}

// TestMaxScroll checks the scroll range for small and large windows.
func TestMaxScroll(t *testing.T) {
	rows := paletteRows()
	if got := maxScroll(rows, 1000); got != 0 {
		t.Errorf("tall window: maxScroll = %d, want 0", got)
	}
	if got, want := maxScroll(rows, 20), contentLines(rows)-18; got != want {
		t.Errorf("maxScroll(20) = %d, want %d", got, want)
	}
}
