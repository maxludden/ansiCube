package main

import (
	"image/color"
	"math"
)

// Palette geometry. The 256-color palette is shown in three sections:
// the 16 system colors, the 6×6×6 color cube (216 colors) and the
// 24-step grayscale ramp.
const (
	systemCols = 8
	systemRows = 2

	cubeCols  = 12
	cubeRows  = 6
	cubeTiers = 3

	grayCols = 12
	grayRows = 2

	cubeBase = 16  // first color cube index
	grayBase = 232 // first grayscale index
)

// paletteRow is one visual row of tiles. A nil ids slice denotes an empty
// spacer row between sections.
type paletteRow struct {
	ids []int
}

// isGap reports whether the row is a spacer with no tiles.
func (r paletteRow) isGap() bool { return r.ids == nil }

// paletteRows returns the full, width-independent palette layout from top
// to bottom. It is computed once at startup and shared by rendering and
// mouse hit-testing so both always agree.
func paletteRows() []paletteRow {
	rows := make([]paletteRow, 0, systemRows+1+cubeRows*cubeTiers+1+grayRows)

	// System colors: 0–7 on the first row, 8–15 on the second.
	for r := 0; r < systemRows; r++ {
		ids := make([]int, systemCols)
		for c := range ids {
			ids[c] = r*systemCols + c
		}
		rows = append(rows, paletteRow{ids: ids})
	}
	rows = append(rows, paletteRow{})

	// Color cube: each tier covers 72 consecutive indices (two 36-color
	// blocks). The left half of a row walks the first block forwards and
	// the right half walks the second block backwards, so adjacent tiles
	// across the seam are visually similar.
	for tier := 0; tier < cubeTiers; tier++ {
		start := cubeBase + tier*72
		for tierRow := 0; tierRow < cubeRows; tierRow++ {
			ids := make([]int, cubeCols)
			for c := range ids {
				if c < cubeCols/2 {
					ids[c] = start + c*6 + tierRow
				} else {
					ids[c] = start + 66 - (c-cubeCols/2)*6 + tierRow
				}
			}
			rows = append(rows, paletteRow{ids: ids})
		}
	}
	rows = append(rows, paletteRow{})

	// Grayscale ramp: 232–243 then 244–255.
	for r := 0; r < grayRows; r++ {
		ids := make([]int, grayCols)
		for c := range ids {
			ids[c] = grayBase + r*grayCols + c
		}
		rows = append(rows, paletteRow{ids: ids})
	}
	return rows
}

// systemRGB holds the conventional xterm values for colors 0–15. Real
// terminals often remap these via their theme, so they are a best guess.
var systemRGB = [16]color.RGBA{
	{0x00, 0x00, 0x00, 0xff}, {0xcd, 0x00, 0x00, 0xff}, {0x00, 0xcd, 0x00, 0xff}, {0xcd, 0xcd, 0x00, 0xff},
	{0x00, 0x00, 0xee, 0xff}, {0xcd, 0x00, 0xcd, 0xff}, {0x00, 0xcd, 0xcd, 0xff}, {0xe5, 0xe5, 0xe5, 0xff},
	{0x7f, 0x7f, 0x7f, 0xff}, {0xff, 0x00, 0x00, 0xff}, {0x00, 0xff, 0x00, 0xff}, {0xff, 0xff, 0x00, 0xff},
	{0x5c, 0x5c, 0xff, 0xff}, {0xff, 0x00, 0xff, 0xff}, {0x00, 0xff, 0xff, 0xff}, {0xff, 0xff, 0xff, 0xff},
}

// cubeLevels are the six channel intensities used by the xterm color cube.
var cubeLevels = [6]uint8{0, 95, 135, 175, 215, 255}

// xtermRGB returns the approximate sRGB value of an ANSI 256-color index.
// Indices outside 0–255 are clamped.
func xtermRGB(id int) color.RGBA {
	id = min(max(id, 0), 255)
	switch {
	case id < cubeBase:
		return systemRGB[id]
	case id < grayBase:
		i := id - cubeBase
		return color.RGBA{cubeLevels[i/36], cubeLevels[(i/6)%6], cubeLevels[i%6], 0xff}
	default:
		v := uint8(8 + 10*(id-grayBase))
		return color.RGBA{v, v, v, 0xff}
	}
}

// relativeLuminance computes WCAG 2.x relative luminance (0 = black,
// 1 = white) for an sRGB color.
func relativeLuminance(c color.RGBA) float64 {
	lin := func(v uint8) float64 {
		s := float64(v) / 255
		if s <= 0.04045 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*lin(c.R) + 0.7152*lin(c.G) + 0.0722*lin(c.B)
}

// useLightText reports whether white text has a higher WCAG contrast ratio
// than black text against the given palette color.
func useLightText(id int) bool {
	l := relativeLuminance(xtermRGB(id))
	contrastWhite := 1.05 / (l + 0.05)
	contrastBlack := (l + 0.05) / 0.05
	return contrastWhite > contrastBlack
}
