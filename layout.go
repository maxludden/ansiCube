package main

// Screen geometry, in terminal cells/lines.
const (
	headerLines      = 2 // title line + blank line
	chromeLines      = 2 // status line (toast / quit prompt) + footer
	tileHeight       = 3
	defaultTileWidth = 5
)

// calcRowWidths splits totalWidth cells across cols tiles as evenly as
// possible, giving the leftmost tiles one extra cell when it does not divide
// evenly. A non-positive totalWidth (window size not yet known) falls back to
// defaultTileWidth per tile. The returned widths always sum to the effective
// total width.
func calcRowWidths(totalWidth, cols int) []int {
	if cols <= 0 {
		return nil
	}
	if totalWidth <= 0 {
		totalWidth = cols * defaultTileWidth
	}
	widths := make([]int, cols)
	base, extra := totalWidth/cols, totalWidth%cols
	for i := range widths {
		widths[i] = base
		if i < extra {
			widths[i]++
		}
	}
	return widths
}

// contentLines is the total height of the scrollable content (header plus
// every palette row).
func contentLines(rows []paletteRow) int {
	return headerLines + len(rows)*tileHeight
}

// viewportLines is how many content lines fit on screen above the status
// line and footer. It is always at least 1.
func viewportLines(winHeight int) int {
	return max(1, winHeight-chromeLines)
}

// maxScroll is the largest valid scroll offset for the given window height.
func maxScroll(rows []paletteRow, winHeight int) int {
	return max(0, contentLines(rows)-viewportLines(winHeight))
}

// tileAt maps a screen coordinate to the palette color under it, using the
// same geometry as rendering. It returns false for the header, spacer rows,
// the status/footer area, or anything outside the tile grid.
func tileAt(rows []paletteRow, x, y, scrollY, winWidth, winHeight int) (int, bool) {
	if x < 0 || y < 0 || y >= viewportLines(winHeight) {
		return 0, false
	}
	line := y + scrollY - headerLines
	if line < 0 {
		return 0, false
	}
	r := line / tileHeight
	if r >= len(rows) || rows[r].isGap() {
		return 0, false
	}
	ids := rows[r].ids
	pos := 0
	for c, w := range calcRowWidths(winWidth, len(ids)) {
		if x >= pos && x < pos+w {
			return ids[c], true
		}
		pos += w
	}
	return 0, false
}
