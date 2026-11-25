package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	scrollStep    = 3
	toastDuration = 2 * time.Second
	authorName    = "Designed by Max Ludden"
	authorURL     = "https://github.com/maxludden"
	quitHint      = "Press q to quit."
)

var (
	titleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Bold(true)
	infoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	footerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	toastStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)

	gradientStart = lipgloss.Color("#ff7a00")
	gradientEnd   = lipgloss.Color("#00d2ff")
)

// toastClearMsg asks the model to hide the toast with the given id. Older
// ids are ignored so a fast second click keeps its full display time.
type toastClearMsg struct{ id int }

// clipboardResultMsg reports the outcome of an asynchronous clipboard write
// started by the click with the given id.
type clipboardResultMsg struct {
	id      int
	text    string
	err     error
	skipped bool
}

// model is the Bubble Tea state for the palette explorer. Rendering is a pure
// function of this state; all mutation happens in Update.
type model struct {
	rows      []paletteRow
	clip      *clipboardWriter // serialized native clipboard writes
	winWidth  int
	winHeight int
	scrollY   int
	confirmQ  bool

	toastText  string
	toastColor int // palette id shown as a swatch; -1 for none
	toastErr   bool
	toastID    int
}

// newModel builds the initial model using copyFn for clipboard writes.
func newModel(copyFn func(string) error) *model {
	return &model{rows: paletteRows(), clip: newClipboardWriter(copyFn), toastColor: -1}
}

// Init implements tea.Model.
func (m *model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.winWidth, m.winHeight = msg.Width, msg.Height
		m.scrollTo(m.scrollY)
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			if id, ok := tileAt(m.rows, msg.X, msg.Y, m.scrollY, m.winWidth, m.winHeight); ok {
				return m, m.copyColor(id)
			}
		}
	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelUp:
			m.scrollTo(m.scrollY - scrollStep)
		case tea.MouseWheelDown:
			m.scrollTo(m.scrollY + scrollStep)
		}
	case clipboardResultMsg:
		return m, m.handleClipboardResult(msg)
	case toastClearMsg:
		if msg.id == m.toastID {
			m.toastText = ""
		}
	}
	return m, nil
}

// handleKey processes key presses, including the quit confirmation prompt.
func (m *model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if m.confirmQ {
		switch key {
		case "y", "Y", "enter":
			return m, tea.Quit
		case "n", "N", "esc":
			m.confirmQ = false
			return m, nil
		}
	}
	switch key {
	case "ctrl+c":
		return m, tea.Quit // always an immediate escape hatch
	case "q":
		m.confirmQ = true
	case "up", "k":
		m.scrollTo(m.scrollY - 1)
	case "down", "j":
		m.scrollTo(m.scrollY + 1)
	case "pgup":
		m.scrollTo(m.scrollY - viewportLines(m.winHeight))
	case "pgdown", "space":
		m.scrollTo(m.scrollY + viewportLines(m.winHeight))
	case "home":
		m.scrollTo(0)
	case "end":
		m.scrollTo(maxScroll(m.rows, m.winHeight))
	}
	return m, nil
}

// scrollTo sets the scroll offset, clamped to the valid range.
func (m *model) scrollTo(y int) {
	m.scrollY = min(max(y, 0), maxScroll(m.rows, m.winHeight))
}

// copyColor shows a toast and starts copying the color number to the
// clipboard. The native write runs as a command so a slow clipboard backend
// cannot freeze the UI; its result arrives as a clipboardResultMsg.
func (m *model) copyColor(id int) tea.Cmd {
	text := strconv.Itoa(id)
	m.toastID++
	m.toastColor = id
	m.toastErr = false
	m.toastText = "copied to clipboard"

	seq, clip := m.toastID, m.clip
	write := func() tea.Msg {
		skipped, err := clip.Write(seq, text)
		return clipboardResultMsg{id: seq, text: text, err: err, skipped: skipped}
	}
	return tea.Batch(clearToastAfter(seq), write)
}

// handleClipboardResult falls back to the OSC 52 terminal clipboard when the
// native write failed (e.g. over SSH or without xclip). Results from clicks
// that a newer click has superseded are ignored, so only the latest color
// can end up on the clipboard.
func (m *model) handleClipboardResult(msg clipboardResultMsg) tea.Cmd {
	if msg.id != m.toastID || msg.skipped || msg.err == nil {
		return nil
	}
	m.toastErr = true
	m.toastText = fmt.Sprintf("native clipboard unavailable (%v); sent via OSC 52", msg.err)
	return tea.SetClipboard(msg.text)
}

// clearToastAfter returns a command that hides toast id after toastDuration.
func clearToastAfter(id int) tea.Cmd {
	return tea.Tick(toastDuration, func(time.Time) tea.Msg { return toastClearMsg{id: id} })
}

// View implements tea.Model. It does not mutate the model.
func (m *model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true // mouse coordinates are only reliable in the alt screen
	v.MouseMode = tea.MouseModeCellMotion
	v.WindowTitle = "ANSI Cube"
	return v
}

// render produces the full screen as a string.
func (m *model) render() string {
	lines := m.contentLines()
	if m.winHeight > 0 {
		start := min(m.scrollY, len(lines))
		end := min(start+viewportLines(m.winHeight), len(lines))
		lines = lines[start:end]
		// Pad so the status line and footer stay pinned to the bottom.
		for len(lines) < viewportLines(m.winHeight) {
			lines = append(lines, "")
		}
	}
	lines = append(lines, m.statusLine(), m.footerLine())
	return strings.Join(lines, "\n")
}

// contentLines renders the header and every palette row as individual lines.
func (m *model) contentLines() []string {
	lines := make([]string, 0, contentLines(m.rows))
	title := titleStyle.Render("ANSI Colors")
	lines = append(lines, lipgloss.PlaceHorizontal(m.winWidth, lipgloss.Center, title), "")

	for _, row := range m.rows {
		if row.isGap() {
			for range tileHeight {
				lines = append(lines, "")
			}
			continue
		}
		widths := calcRowWidths(m.winWidth, len(row.ids))
		tiles := make([]string, len(row.ids))
		for c, id := range row.ids {
			tiles[c] = renderTile(id, widths[c], tileHeight)
		}
		joined := lipgloss.JoinHorizontal(lipgloss.Top, tiles...)
		lines = append(lines, strings.Split(joined, "\n")...)
	}
	return lines
}

// tileStyle returns the background style for a palette color with a
// foreground chosen for maximum contrast.
func tileStyle(id int) lipgloss.Style {
	// Use true black/white rather than ANSI 0/15: terminal themes often remap
	// those to mid-grays, which ruins the contrast useLightText computed.
	fg := lipgloss.Color("#000000")
	if useLightText(id) {
		fg = lipgloss.Color("#ffffff")
	}
	return lipgloss.NewStyle().Background(lipgloss.Color(strconv.Itoa(id))).Foreground(fg)
}

// renderTile draws one color tile labelled with its zero-padded index.
func renderTile(id, width, height int) string {
	if width <= 0 {
		return ""
	}
	label := fmt.Sprintf("%03d", id)
	return tileStyle(id).Width(width).Height(height).
		AlignHorizontal(lipgloss.Center).AlignVertical(lipgloss.Center).
		Render(label)
}

// statusLine shows the quit prompt, the latest toast, or nothing.
func (m *model) statusLine() string {
	switch {
	case m.confirmQ:
		return fitWidth(infoStyle.Render("Quit? Press y to confirm, n to cancel."), m.winWidth)
	case m.toastText == "":
		return ""
	}
	style := toastStyle
	if m.toastErr {
		style = errorStyle
	}
	swatch := ""
	if m.toastColor >= 0 {
		swatch = tileStyle(m.toastColor).Render(fmt.Sprintf(" %03d ", m.toastColor)) + " "
	}
	return fitWidth(swatch+style.Render(m.toastText), m.winWidth)
}

// fitWidth truncates s to width cells (ANSI-aware). A non-positive width,
// meaning the window size is not yet known, leaves s unchanged.
func fitWidth(s string, width int) string {
	if width <= 0 {
		return s
	}
	return ansi.Truncate(s, width, "…")
}

// footerLine renders the quit hint on the left and the author credit,
// a gradient hyperlink, on the right.
func (m *model) footerLine() string {
	left := footerStyle.Render(quitHint)
	right := gradientLink(authorName, authorURL)
	gap := m.winWidth - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		// Not enough room for both: keep the quit hint, drop the credit.
		return fitWidth(left, m.winWidth)
	}
	return left + strings.Repeat(" ", gap) + right
}

// gradientLink renders text with a per-character color gradient, wrapped
// in an OSC 8 hyperlink to url.
func gradientLink(text, url string) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return ""
	}
	colors := lipgloss.Blend1D(len(runes), gradientStart, gradientEnd)
	var b strings.Builder
	for i, r := range runes {
		// A shared id lets terminals treat every cell as a single link.
		s := lipgloss.NewStyle().Foreground(colors[i]).Hyperlink(url, "id=author")
		b.WriteString(s.Render(string(r)))
	}
	return b.String()
}
