package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	infoStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

const (
	cubeCols         = 12
	cubeRows         = 6
	cubeTiers        = 3
	grayCols         = 12
	grayRows         = 2
	systemCols       = 8
	systemRows       = 2
	sectionGapRows   = 1
	defaultTileWidth = 5
)

type model struct {
	lastCopied string
	// clickMap stores: map[row_index][column_index]ColorID
	clickMap      map[int]map[int]int
	scrollY       int
	winHeight     int
	winWidth      int
	confirmQ      bool
	rowColWidths  map[int][]int
	rowLineStarts []int
	rowHeights    []int
	totalLines    int
}

func initialModel() *model {
	m := &model{
		clickMap:     make(map[int]map[int]int),
		rowColWidths: make(map[int][]int),
	}
	return m
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.confirmQ {
			switch msg.String() {
			case "y", "Y", "enter":
				return m, tea.Quit
			case "n", "N", "esc":
				m.confirmQ = false
				return m, nil
			}
		}
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if msg.String() == "g" {
			m.confirmQ = true
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.winHeight = msg.Height
		m.winWidth = msg.Width
	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			if len(m.rowLineStarts) == 0 || len(m.rowHeights) == 0 {
				return m, nil
			}
			line := msg.Y + m.scrollY
			row := -1
			for i, start := range m.rowLineStarts {
				if line >= start && line < start+m.rowHeights[i] {
					row = i
					break
				}
			}
			if row < 0 {
				return m, nil
			}
			colWidths := m.rowColWidths[row]
			if len(colWidths) == 0 {
				return m, nil
			}
			col := -1
			x := msg.X
			pos := 0
			for i, w := range colWidths {
				if w <= 0 {
					continue
				}
				if x >= pos && x < pos+w {
					col = i
					break
				}
				pos += w
			}
			if col < 0 {
				return m, nil
			}
			if r, ok := m.clickMap[row]; ok {
				if colorID, ok := r[col]; ok {
					colorStr := strconv.Itoa(colorID)
					clipboard.WriteAll(colorStr)
					m.lastCopied = colorStr
				}
			}
		}
		if msg.Type == tea.MouseWheelUp {
			if m.scrollY > 0 {
				m.scrollY--
			}
		}
		if msg.Type == tea.MouseWheelDown {
			if m.scrollY < max(0, m.contentHeight()-m.winHeight) {
				m.scrollY++
			}
		}
	}
	return m, nil
}

// renderTile creates a single color square and registers it in the clickMap
func (m *model) renderTile(id, row, col, width, height int) string {
	if m.clickMap[row] == nil {
		m.clickMap[row] = make(map[int]int)
	}
	m.clickMap[row][col] = id

	bg := lipgloss.Color(strconv.Itoa(id))
	// Determine text contrast
	fg := lipgloss.Color("0") // Black text
	if id < 8 || (id >= 16 && id <= 231 && id%36 < 18) || (id >= 232 && id <= 243) {
		fg = lipgloss.Color("15") // White text
	}

	return lipgloss.NewStyle().
		Background(bg).
		Foreground(fg).
		Width(width).
		Height(height).
		Align(lipgloss.Center).
		AlignVertical(lipgloss.Center).
		Render(fmt.Sprintf("%03d", id))
}

func calcRowWidths(totalWidth, cols int) []int {
	if cols <= 0 {
		return nil
	}
	if totalWidth <= 0 {
		totalWidth = cols * defaultTileWidth
	}
	widths := make([]int, cols)
	base := totalWidth / cols
	extra := totalWidth % cols
	for i := 0; i < cols; i++ {
		widths[i] = base
		if i < extra {
			widths[i]++
		}
	}
	if base == 0 && extra < cols {
		for i := extra; i < cols; i++ {
			widths[i] = 0
		}
	}
	return widths
}

func (m *model) layoutRowHeights(totalRows int) ([]int, int) {
	extraLines := 0
	if m.lastCopied != "" {
		extraLines++
	}
	if m.confirmQ {
		extraLines++
	}
	available := m.winHeight - extraLines
	if available <= 0 {
		available = totalRows
	}
	base := 1
	extra := 0
	if available >= totalRows {
		base = available / totalRows
		extra = available % totalRows
	}
	heights := make([]int, totalRows)
	total := 0
	for i := 0; i < totalRows; i++ {
		h := base
		if i < extra {
			h++
		}
		if h < 1 {
			h = 1
		}
		heights[i] = h
		total += h
	}
	total += extraLines
	return heights, total
}

func (m *model) View() string {
	m.clickMap = make(map[int]map[int]int)
	m.rowWidths = make(map[int]int)
	m.rowOffsets = make(map[int]int)
	if m.blocksPerRow <= 0 {
		m.blocksPerRow = cubeBlocks
	}

	var b strings.Builder
	row := 0

	b.WriteString(titleStyle.Render("XTERM-256 COLOR CHART (Click to Copy)"))
	b.WriteString("\n\n")
	row += 2

	gapTile := strings.Repeat(" ", m.tileWidth*blockGapCols)
	cubeAreaCols := (m.blocksPerRow * cubeCols) + ((m.blocksPerRow - 1) * blockGapCols)
	cubeAreaWidth := cubeAreaCols * m.tileWidth
	padCubeChars := max(0, (m.winWidth-cubeAreaWidth)/2)
	padCube := strings.Repeat(" ", padCubeChars)
	systemTileWidth := calcFullWidthTile(m.winWidth, systemCols)
	grayTileWidth := calcFullWidthTile(m.winWidth, grayCols)

	for r := 0; r < systemRows; r++ {
		var rowStr strings.Builder
		col := 0
		m.rowWidths[row] = systemTileWidth
		m.rowOffsets[row] = 0
		for c := 0; c < systemCols; c++ {
			id := (r * 8) + c
			rowStr.WriteString(m.renderTile(id, row, col, systemTileWidth))
			col++
		}

		b.WriteString(rowStr.String())
		b.WriteString("\n")
		row++
	}

	b.WriteString("\n")
	row++

	blockRows := (cubeBlocks + m.blocksPerRow - 1) / m.blocksPerRow
	for blockRow := 0; blockRow < blockRows; blockRow++ {
		for r := 0; r < cubeRows; r++ {
			var rowStr strings.Builder
			col := 0
			m.rowWidths[row] = m.tileWidth
			m.rowOffsets[row] = padCubeChars
			rowStr.WriteString(padCube)
			for blockCol := 0; blockCol < m.blocksPerRow; blockCol++ {
				blockIndex := (blockRow * m.blocksPerRow) + blockCol
				if blockIndex >= cubeBlocks {
					rowStr.WriteString(strings.Repeat(" ", cubeCols*m.tileWidth))
					col += cubeCols
				} else {
					start := 16 + (blockIndex * 72)
					for c := 0; c < cubeCols; c++ {
						var id int
						if c < 6 {
							id = start + (c * 6) + r
						} else {
							id = (start + 66) - ((c - 6) * 6) + r
						}
						rowStr.WriteString(m.renderTile(id, row, col, m.tileWidth))
						col++
					}
				}
				if blockCol < m.blocksPerRow-1 {
					rowStr.WriteString(gapTile)
					col += blockGapCols
				}
			}
			b.WriteString(rowStr.String())
			b.WriteString("\n")
			row++
		}
	}

	b.WriteString("\n")
	row++

	for r := 0; r < grayRows; r++ {
		var rowStr strings.Builder
		col := 0
		m.rowWidths[row] = grayTileWidth
		m.rowOffsets[row] = 0
		for c := 0; c < grayCols; c++ {
			id := 232 + (r * 12) + c
			rowStr.WriteString(m.renderTile(id, row, col, grayTileWidth))
			col++
		}
		b.WriteString(rowStr.String())
		b.WriteString("\n")
		row++
	}

	if m.lastCopied != "" {
		b.WriteString("\n")
		b.WriteString(infoStyle.Render(fmt.Sprintf("Copied ANSI %s to clipboard!", m.lastCopied)))
	}
	if m.confirmQ {
		b.WriteString("\n")
		b.WriteString(infoStyle.Render("Quit? Press y to confirm, n to cancel."))
	}

	content := strings.TrimRight(b.String(), "\n")
	lines := strings.Split(content, "\n")
	if m.winHeight <= 0 {
		return content
	}
	maxScroll := max(0, len(lines)-m.winHeight)
	if m.scrollY > maxScroll {
		m.scrollY = maxScroll
	}
	start := min(m.scrollY, len(lines))
	end := min(start+m.winHeight, len(lines))
	return strings.Join(lines[start:end], "\n")
}

func (m *model) contentHeight() int {
	// Keep in sync with View() row accounting.
	height := 2 // title + blank line
	height += systemRows
	height += 1
	blocksPerRow := m.blocksPerRow
	if blocksPerRow <= 0 {
		blocksPerRow = cubeBlocks
	}
	blockRows := (cubeBlocks + blocksPerRow - 1) / blocksPerRow
	height += cubeRows * blockRows
	height += 1
	height += grayRows
	if m.lastCopied != "" {
		height += 2
	}
	if m.confirmQ {
		height += 2
	}
	return height
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
