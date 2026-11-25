package main

import (
	"image/color"
	"testing"
)

// TestPaletteRowsCoverAllColorsOnce checks the strongest layout invariant:
// every ANSI 256 index appears exactly once.
func TestPaletteRowsCoverAllColorsOnce(t *testing.T) {
	seen := make(map[int]int)
	for _, row := range paletteRows() {
		for _, id := range row.ids {
			seen[id]++
		}
	}
	for id := 0; id < 256; id++ {
		if seen[id] != 1 {
			t.Errorf("color %d appears %d times, want 1", id, seen[id])
		}
	}
	if len(seen) != 256 {
		t.Errorf("got %d distinct ids, want 256", len(seen))
	}
}

// TestPaletteRowsShape checks section sizes and spacer placement.
func TestPaletteRowsShape(t *testing.T) {
	rows := paletteRows()
	want := systemRows + 1 + cubeRows*cubeTiers + 1 + grayRows
	if len(rows) != want {
		t.Fatalf("len(rows) = %d, want %d", len(rows), want)
	}
	if !rows[systemRows].isGap() || !rows[systemRows+1+cubeRows*cubeTiers].isGap() {
		t.Error("expected spacer rows between sections")
	}
	if got := rows[0].ids; got[0] != 0 || got[7] != 7 {
		t.Errorf("first system row = %v", got)
	}
	if got := rows[len(rows)-1].ids; got[len(got)-1] != 255 {
		t.Errorf("last gray row = %v", got)
	}
}

// TestXtermRGB spot-checks the index → RGB conversion.
func TestXtermRGB(t *testing.T) {
	tests := []struct {
		id   int
		want color.RGBA
	}{
		{0, color.RGBA{0, 0, 0, 255}},
		{15, color.RGBA{255, 255, 255, 255}},
		{16, color.RGBA{0, 0, 0, 255}},
		{21, color.RGBA{0, 0, 255, 255}},
		{196, color.RGBA{255, 0, 0, 255}},
		{231, color.RGBA{255, 255, 255, 255}},
		{232, color.RGBA{8, 8, 8, 255}},
		{255, color.RGBA{238, 238, 238, 255}},
		{-5, color.RGBA{0, 0, 0, 255}},        // clamped to 0
		{999, color.RGBA{238, 238, 238, 255}}, // clamped to 255
	}
	for _, tt := range tests {
		if got := xtermRGB(tt.id); got != tt.want {
			t.Errorf("xtermRGB(%d) = %v, want %v", tt.id, got, tt.want)
		}
	}
}

// TestUseLightText checks that text contrast flips at sensible colors.
func TestUseLightText(t *testing.T) {
	light := []int{0, 4, 16, 17, 52, 232, 240}
	dark := []int{7, 11, 15, 46, 226, 231, 250, 255}
	for _, id := range light {
		if !useLightText(id) {
			t.Errorf("color %d: want light text", id)
		}
	}
	for _, id := range dark {
		if useLightText(id) {
			t.Errorf("color %d: want dark text", id)
		}
	}
}
