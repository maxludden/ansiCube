# ANSI Cube

An interactive ANSI 256-color explorer for the terminal.

![ANSI Cube showing the full 256-color palette](assets/screenshots/palette.png)

ANSI Cube shows every color a 256-color terminal can display as a grid of
labelled tiles:

- **System colors** (0–15), which your terminal theme usually remaps.
- **The 6×6×6 color cube** (16–231).
- **The grayscale ramp** (232–255).

Click any tile and its color number is copied to your clipboard, ready to paste
into a Lip Gloss style, a shell prompt, a `tmux` config, or anywhere else that
takes an ANSI 256 color index.

## Highlights

- **Click to copy.** Uses the native clipboard and falls back to the OSC 52
  terminal clipboard over SSH.
- **Readable labels.** Each tile picks black or white text by WCAG contrast.
- **Fits your terminal.** Tiles stretch to the window width and the palette
  scrolls with the mouse wheel or keyboard.
- **Cross-platform.** Prebuilt for macOS, Linux, and Windows.

## Quick start

```bash
brew install maxludden/tap/ansicube
ansicube
```

See [Installation](install.md) for other options and [Usage](usage.md) for the
full list of controls.

![Animated demo of scrolling through ANSI Cube](assets/screenshots/demo.gif)
