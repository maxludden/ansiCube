package main

import "sync"

// clipboardWriter serializes native clipboard writes, which run off the UI
// goroutine because backends such as xclip can block. Each write carries a
// sequence number; a write older than one already performed is skipped, so a
// slow earlier click can never overwrite a later one.
type clipboardWriter struct {
	mu     sync.Mutex
	latest int
	write  func(string) error
}

// newClipboardWriter wraps write, the underlying clipboard backend.
func newClipboardWriter(write func(string) error) *clipboardWriter {
	return &clipboardWriter{write: write}
}

// Write copies text unless a write with a higher seq has already happened.
// It reports whether the write was skipped as stale.
func (c *clipboardWriter) Write(seq int, text string) (skipped bool, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if seq < c.latest {
		return true, nil
	}
	c.latest = seq
	return false, c.write(text)
}
