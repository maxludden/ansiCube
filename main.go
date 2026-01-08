package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

var (
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	infoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("0"))
	toastStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
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
	tileHeight       = 3
	scrollStep       = 3
	footerLines      = 1
)

type toastClearMsg struct {
	id int
}

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
	toastText     string
	toastID       int
	toastColor    lipgloss.Color
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
					m.toastText = fmt.Sprintf("Copied ANSI %s to clipboard", colorStr)
					m.toastColor = lipgloss.Color(colorStr)
					m.toastID++
					id := m.toastID
					return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
						return toastClearMsg{id: id}
					})
				}
			}
		}
		if msg.Type == tea.MouseWheelUp {
			if m.scrollY > 0 {
				m.scrollY -= scrollStep
				if m.scrollY < 0 {
					m.scrollY = 0
				}
			}
		}
		if msg.Type == tea.MouseWheelDown {
			if m.scrollY < max(0, m.contentHeight()-m.viewportHeight()) {
				m.scrollY += scrollStep
				maxScroll := max(0, m.contentHeight()-m.viewportHeight())
				if m.scrollY > maxScroll {
					m.scrollY = maxScroll
				}
			}
		}
	case toastClearMsg:
		if msg.id == m.toastID {
			m.toastText = ""
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

	style := lipgloss.NewStyle().Background(bg).Foreground(fg)
	content := fmt.Sprintf("%03d", id)
	placed := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
	return style.Render(placed)
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
	heights := make([]int, totalRows)
	for i := 0; i < totalRows; i++ {
		heights[i] = tileHeight
	}
	total := 0
	for i := 0; i < totalRows; i++ {
		if heights[i] < 1 {
			heights[i] = 1
		}
		total += heights[i]
	}
	return heights, total
}

func (m *model) viewportHeight() int {
	toastLines := 0
	if m.toastText != "" {
		toastLines = 1
	}
	confirmLines := 0
	if m.confirmQ {
		confirmLines = 1
	}
	visible := m.winHeight - footerLines - toastLines - confirmLines
	if visible < 1 {
		visible = 1
	}
	return visible
}

func (m *model) View() string {
	m.clickMap = make(map[int]map[int]int)
	m.rowColWidths = make(map[int][]int)

	var b strings.Builder

	titleLine := titleStyle.Render("ANSI Colors")
	b.WriteString(lipgloss.Place(m.winWidth, 1, lipgloss.Center, lipgloss.Center, titleLine))
	b.WriteString("\n\n")
	titleHeight := 2

	totalRows := systemRows + sectionGapRows + (cubeRows * cubeTiers) + sectionGapRows + grayRows
	cubeStart := systemRows + sectionGapRows
	cubeEnd := cubeStart + (cubeRows * cubeTiers)
	rowHeights, totalLines := m.layoutRowHeights(totalRows)
	m.rowHeights = rowHeights
	m.rowLineStarts = make([]int, totalRows)
	m.totalLines = totalLines

	systemWidths := calcRowWidths(m.winWidth, systemCols)
	cubeWidths := calcRowWidths(m.winWidth, cubeCols)
	grayWidths := calcRowWidths(m.winWidth, grayCols)

	lineCursor := titleHeight
	for row := 0; row < totalRows; row++ {
		height := rowHeights[row]
		m.rowLineStarts[row] = lineCursor

		switch {
		case row < systemRows:
			m.rowColWidths[row] = systemWidths
			tiles := make([]string, 0, systemCols)
			for c := 0; c < systemCols; c++ {
				id := (row * 8) + c
				tiles = append(tiles, m.renderTile(id, row, c, systemWidths[c], height))
			}
			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tiles...))
		case row == systemRows:
			b.WriteString(lipgloss.NewStyle().Width(m.winWidth).Height(height).Render(""))
		case row < cubeEnd:
			cubeIndex := row - (systemRows + sectionGapRows)
			tier := cubeIndex / cubeRows
			tierRow := cubeIndex % cubeRows
			m.rowColWidths[row] = cubeWidths
			tiles := make([]string, 0, cubeCols)
			start := 16 + (tier * 72)
			for c := 0; c < cubeCols; c++ {
				var id int
				if c < 6 {
					id = start + (c * 6) + tierRow
				} else {
					id = (start + 66) - ((c - 6) * 6) + tierRow
				}
				tiles = append(tiles, m.renderTile(id, row, c, cubeWidths[c], height))
			}
			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tiles...))
		case row == cubeEnd:
			b.WriteString(lipgloss.NewStyle().Width(m.winWidth).Height(height).Render(""))
		default:
			grayIndex := row - (systemRows + sectionGapRows + (cubeRows * cubeTiers) + sectionGapRows)
			m.rowColWidths[row] = grayWidths
			tiles := make([]string, 0, grayCols)
			for c := 0; c < grayCols; c++ {
				id := 232 + (grayIndex * 12) + c
				tiles = append(tiles, m.renderTile(id, row, c, grayWidths[c], height))
			}
			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tiles...))
		}

		b.WriteString("\n")
		lineCursor += height
	}

	content := strings.TrimRight(b.String(), "\n")
	lines := strings.Split(content, "\n")
	if m.winHeight <= 0 {
		return content
	}
	visibleHeight := m.viewportHeight()
	maxScroll := max(0, len(lines)-visibleHeight)
	if m.scrollY > maxScroll {
		m.scrollY = maxScroll
	}
	start := min(m.scrollY, len(lines))
	end := min(start+visibleHeight, len(lines))
	view := strings.Join(lines[start:end], "\n")
	if m.toastText != "" {
		toast := toastStyle.Copy().Foreground(m.toastColor).Render(m.toastText)
		view += "\n" + toast
	}
	if m.confirmQ {
		view += "\n" + infoStyle.Render("Quit? Press y to confirm, n to cancel.")
	}
	left := footerStyle.Render("Press q or Ctrl+C to quit.")
	rightText := "Designed by Max Ludden"
	right := osc8Link(gradientText(rightText), "https://github.com/maxludden")
	view += "\n" + footerLine(m.winWidth, left, right, "Press q or Ctrl+C to quit.", rightText)
	return view
}

func (m *model) contentHeight() int {
	totalRows := systemRows + sectionGapRows + (cubeRows * cubeTiers) + sectionGapRows + grayRows
	_, total := m.layoutRowHeights(totalRows)
	return total + 2
}

func gradientText(text string) string {
	start := [3]int{255, 122, 0}
	end := [3]int{0, 210, 255}
	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}
	var b strings.Builder
	for i, r := range runes {
		t := float64(i) / float64(len(runes)-1)
		if len(runes) == 1 {
			t = 0
		}
		red := int(float64(start[0]) + (float64(end[0])-float64(start[0]))*t)
		green := int(float64(start[1]) + (float64(end[1])-float64(start[1]))*t)
		blue := int(float64(start[2]) + (float64(end[2])-float64(start[2]))*t)
		color := lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", red, green, blue))
		b.WriteString(lipgloss.NewStyle().Foreground(color).Render(string(r)))
	}
	return b.String()
}

func osc8Link(text, url string) string {
	return "\x1b]8;;" + url + "\x07" + text + "\x1b]8;;\x07"
}

func footerLine(width int, left, right, leftPlain, rightPlain string) string {
	if width <= 0 {
		return left + " " + right
	}
	leftWidth := runewidth.StringWidth(leftPlain)
	rightWidth := runewidth.StringWidth(rightPlain)
	gap := width - leftWidth - rightWidth
	if gap < 1 {
		return left + " " + right
	}
	return left + strings.Repeat(" ", gap) + right
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
