package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// testModel returns a sized model whose clipboard writes are recorded.
func testModel(t *testing.T, copyErr error) (*model, *[]string) {
	t.Helper()
	var copied []string
	m := newModel(func(s string) error {
		copied = append(copied, s)
		return copyErr
	})
	m.Update(tea.WindowSizeMsg{Width: 96, Height: 30})
	return m, &copied
}

// isQuit reports whether cmd produces tea.QuitMsg.
func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	_, ok := cmd().(tea.QuitMsg)
	return ok
}

// TestKeyStrings pins the v2 key strings the model matches on.
func TestKeyStrings(t *testing.T) {
	tests := map[string]tea.KeyPressMsg{
		"q":      {Code: 'q', Text: "q"},
		"y":      {Code: 'y', Text: "y"},
		"Y":      {Code: 'y', ShiftedCode: 'Y', Text: "Y", Mod: tea.ModShift},
		"n":      {Code: 'n', Text: "n"},
		"esc":    {Code: tea.KeyEscape},
		"enter":  {Code: tea.KeyEnter},
		"ctrl+c": {Code: 'c', Mod: tea.ModCtrl},
		"pgdown": {Code: tea.KeyPgDown},
	}
	for want, msg := range tests {
		if got := msg.String(); got != want {
			t.Errorf("String() = %q, want %q", got, want)
		}
	}
}

// TestQuitKeys checks Ctrl+C quits at once and q asks for confirmation.
func TestQuitKeys(t *testing.T) {
	m, _ := testModel(t, nil)
	if _, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}); !isQuit(cmd) {
		t.Error("ctrl+c should quit immediately")
	}

	m, _ = testModel(t, nil)
	if _, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"}); cmd != nil || !m.confirmQ {
		t.Fatal("q should open the quit prompt, not quit")
	}
	if !strings.Contains(m.render(), "Quit?") {
		t.Error("prompt not rendered")
	}
	if _, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape}); isQuit(cmd) || m.confirmQ {
		t.Error("esc should cancel the prompt")
	}
	m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if _, cmd := m.Update(tea.KeyPressMsg{Code: 'n', Text: "n"}); isQuit(cmd) || m.confirmQ {
		t.Error("n should cancel the prompt")
	}
	for _, confirm := range []tea.KeyPressMsg{
		{Code: 'y', Text: "y"},
		{Code: 'y', ShiftedCode: 'Y', Text: "Y", Mod: tea.ModShift},
		{Code: tea.KeyEnter},
	} {
		m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
		if _, cmd := m.Update(confirm); !isQuit(cmd) {
			t.Errorf("%q should confirm quit", confirm.String())
		}
	}
}

// TestClickCopiesColor checks a left click copies the tile's number.
func TestClickCopiesColor(t *testing.T) {
	m, copied := testModel(t, nil)
	_, cmd := m.Update(tea.MouseClickMsg{X: 0, Y: headerLines + 3*tileHeight, Button: tea.MouseLeft})
	result, ok := runBatch(t, cmd)
	if !ok {
		t.Fatal("click did not start a clipboard write")
	}
	if len(*copied) != 1 || (*copied)[0] != "16" {
		t.Fatalf("copied = %v, want [16]", *copied)
	}
	if _, cmd := m.Update(result); cmd != nil {
		t.Error("successful write should not fall back to OSC 52")
	}
	if m.toastErr || m.toastColor != 16 || !strings.Contains(m.render(), "copied to clipboard") {
		t.Errorf("unexpected toast state: %+v", m)
	}

	// A stale clear message must not hide a newer toast.
	m.Update(toastClearMsg{id: m.toastID - 1})
	if m.toastText == "" {
		t.Error("stale clear hid the toast")
	}
	m.Update(toastClearMsg{id: m.toastID})
	if m.toastText != "" {
		t.Error("current clear did not hide the toast")
	}
}

// TestClickIgnoresNonTiles checks right clicks and spacer clicks do nothing.
func TestClickIgnoresNonTiles(t *testing.T) {
	m, copied := testModel(t, nil)
	m.Update(tea.MouseClickMsg{X: 0, Y: headerLines, Button: tea.MouseRight})
	m.Update(tea.MouseClickMsg{X: 0, Y: headerLines + 2*tileHeight, Button: tea.MouseLeft})
	m.Update(tea.MouseClickMsg{X: 0, Y: 0, Button: tea.MouseLeft})
	if len(*copied) != 0 {
		t.Errorf("copied = %v, want none", *copied)
	}
}

// TestClipboardFailureFallsBack checks a failed native write shows an error
// toast and issues an OSC 52 write of the same color.
func TestClipboardFailureFallsBack(t *testing.T) {
	m, _ := testModel(t, errors.New("no xclip"))
	_, cmd := m.Update(tea.MouseClickMsg{X: 0, Y: headerLines, Button: tea.MouseLeft})
	result, ok := runBatch(t, cmd)
	if !ok || result.err == nil {
		t.Fatalf("expected a failed clipboard result, got %+v", result)
	}
	_, cmd = m.Update(result)
	if !m.toastErr || !strings.Contains(m.toastText, "OSC 52") {
		t.Errorf("expected OSC 52 fallback toast, got %q", m.toastText)
	}
	if cmd == nil {
		t.Fatal("expected an OSC 52 clipboard command")
	}
	// tea.SetClipboard's message type is unexported; check its name and payload.
	msg := cmd()
	if typ := fmt.Sprintf("%T", msg); !strings.Contains(typ, "setClipboardMsg") || fmt.Sprint(msg) != "0" {
		t.Errorf("fallback message = %s(%v), want setClipboardMsg(0)", typ, msg)
	}
}

// TestStaleClipboardResultIgnored checks a superseded click cannot trigger
// an OSC 52 write that would overwrite the newer color.
func TestStaleClipboardResultIgnored(t *testing.T) {
	m, _ := testModel(t, errors.New("no xclip"))
	m.Update(tea.MouseClickMsg{X: 0, Y: headerLines, Button: tea.MouseLeft})
	m.Update(tea.MouseClickMsg{X: 12, Y: headerLines, Button: tea.MouseLeft})
	stale := clipboardResultMsg{id: m.toastID - 1, text: "0", err: errors.New("no xclip")}
	if _, cmd := m.Update(stale); cmd != nil || m.toastErr {
		t.Error("stale result should be ignored")
	}
}

// TestClipboardWriterSkipsStale checks older writes never land after newer ones.
func TestClipboardWriterSkipsStale(t *testing.T) {
	var got []string
	w := newClipboardWriter(func(s string) error { got = append(got, s); return nil })
	if skipped, _ := w.Write(2, "new"); skipped {
		t.Fatal("first write skipped")
	}
	if skipped, _ := w.Write(1, "old"); !skipped {
		t.Error("older write should be skipped")
	}
	if len(got) != 1 || got[0] != "new" {
		t.Errorf("clipboard writes = %v, want [new]", got)
	}
}

// runBatch executes cmd (possibly a tea.Batch) and returns the first
// clipboardResultMsg it produces. Tick commands are skipped, not awaited.
func runBatch(t *testing.T, cmd tea.Cmd) (clipboardResultMsg, bool) {
	t.Helper()
	if cmd == nil {
		return clipboardResultMsg{}, false
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c == nil {
				continue
			}
			// Each Batch entry is either the clipboard write or a toast tick;
			// run only non-tick commands so the test does not sleep.
			if r, ok := runIfClipboard(c); ok {
				return r, true
			}
		}
		return clipboardResultMsg{}, false
	}
	r, ok := msg.(clipboardResultMsg)
	return r, ok
}

// runIfClipboard runs c in a goroutine and returns its result if it is a
// clipboardResultMsg delivered promptly; tick commands time out and are ignored.
func runIfClipboard(c tea.Cmd) (clipboardResultMsg, bool) {
	ch := make(chan tea.Msg, 1)
	go func() { ch <- c() }()
	select {
	case msg := <-ch:
		r, ok := msg.(clipboardResultMsg)
		return r, ok
	case <-time.After(200 * time.Millisecond):
		return clipboardResultMsg{}, false
	}
}

// TestScrollClamps checks wheel and key scrolling stay in range.
func TestScrollClamps(t *testing.T) {
	m, _ := testModel(t, nil)
	m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if m.scrollY != 0 {
		t.Errorf("scrollY = %d after wheel up at top", m.scrollY)
	}
	limit := maxScroll(m.rows, m.winHeight)
	for range 100 {
		m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	}
	if m.scrollY != limit {
		t.Errorf("scrollY = %d, want %d", m.scrollY, limit)
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyHome})
	if m.scrollY != 0 {
		t.Errorf("home: scrollY = %d", m.scrollY)
	}
	m.Update(tea.KeyPressMsg{Code: tea.KeyEnd})
	if m.scrollY != limit {
		t.Errorf("end: scrollY = %d", m.scrollY)
	}
	// Growing the window re-clamps the offset.
	m.Update(tea.WindowSizeMsg{Width: 96, Height: 500})
	if m.scrollY != 0 {
		t.Errorf("after resize: scrollY = %d", m.scrollY)
	}
}

// TestRenderFillsWindow checks the frame is exactly the window size.
func TestRenderFillsWindow(t *testing.T) {
	m, _ := testModel(t, nil)
	lines := strings.Split(m.render(), "\n")
	if len(lines) != m.winHeight {
		t.Fatalf("rendered %d lines, want %d", len(lines), m.winHeight)
	}
	for i, l := range lines {
		if w := lipgloss.Width(l); w > m.winWidth {
			t.Errorf("line %d width %d exceeds %d", i, w, m.winWidth)
		}
	}
	// Tile rows must span the full window width.
	tileLine := lines[headerLines]
	if w := lipgloss.Width(tileLine); w != m.winWidth {
		t.Errorf("tile row width %d, want %d", w, m.winWidth)
	}
	if !strings.Contains(lines[len(lines)-1], quitHint) {
		t.Error("footer missing quit hint")
	}
}

// TestRenderedTilesMatchHitTest checks rendered tile labels sit where
// tileAt says they are.
func TestRenderedTilesMatchHitTest(t *testing.T) {
	m := newModel(func(string) error { return nil })
	m.Update(tea.WindowSizeMsg{Width: 96, Height: contentLines(m.rows) + chromeLines})
	lines := m.contentLines()
	if len(lines) != contentLines(m.rows) {
		t.Fatalf("contentLines rendered %d, geometry says %d", len(lines), contentLines(m.rows))
	}
	// Middle line of the first cube row, first tile, should read "016".
	y := headerLines + 3*tileHeight + 1
	plain := stripANSI(lines[y])
	if id, ok := tileAt(m.rows, 0, y, 0, m.winWidth, m.winHeight); !ok || id != 16 || !strings.HasPrefix(strings.TrimSpace(plain), "016") {
		t.Errorf("line %q / tileAt=(%d,%v)", plain, id, ok)
	}
}

// stripANSI removes escape sequences for plain-text assertions.
func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		switch {
		case r == 0x1b:
			inEsc = true
		case inEsc && (r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'):
			inEsc = false
		case !inEsc:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// TestGradientLink checks the author credit keeps its visible width.
func TestGradientLink(t *testing.T) {
	got := gradientLink(authorName, authorURL)
	if w := lipgloss.Width(got); w != len(authorName) {
		t.Errorf("width = %d, want %d", w, len(authorName))
	}
	if !strings.Contains(got, authorURL) {
		t.Error("hyperlink URL missing")
	}
	if gradientLink("", authorURL) != "" {
		t.Error("empty text should render empty")
	}
}

// TestNarrowWindowChromeFits checks the status and footer rows are truncated
// rather than wrapping in a narrow window.
func TestNarrowWindowChromeFits(t *testing.T) {
	m, _ := testModel(t, errors.New("a very long clipboard error message"))
	m.Update(tea.WindowSizeMsg{Width: 20, Height: 10})
	_, cmd := m.Update(tea.MouseClickMsg{X: 0, Y: headerLines, Button: tea.MouseLeft})
	if r, ok := runBatch(t, cmd); ok {
		m.Update(r)
	}
	for _, line := range []string{m.statusLine(), m.footerLine()} {
		if w := lipgloss.Width(line); w > 20 {
			t.Errorf("chrome line width %d exceeds 20: %q", w, stripANSI(line))
		}
	}
	m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if w := lipgloss.Width(m.statusLine()); w > 20 {
		t.Errorf("quit prompt width %d exceeds 20", w)
	}
}
