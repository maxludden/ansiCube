# AnsiCube

An interactive ANSI 256-color explorer for the terminal.

AnsiCube renders system colors, the full 6×6×6 color cube, and grayscale ramps as clickable tiles that automatically adapt to your terminal size. Click any color to copy its ANSI color code directly to the clipboard, with instant visual feedback. It supports mouse interaction, smooth scrolling, dynamic resizing, and high-contrast text for readability. Built with Bubble Tea and Lip Gloss.

## Features

- Click any tile to copy its ANSI color number to the clipboard
- Responsive layout with scrolling
- High-contrast text for readability
- Mouse support

## Install

### Homebrew

```bash
brew install maxludden/tap/ansicube
```

### Go install

```bash
go install github.com/maxludden/ansiCube@latest
```

### Binary release

Download a prebuilt binary from the GitHub Releases page.

## Usage

```bash
ansicube
```

## Controls

- `q` or `ctrl+c` to quit
- Click a color tile to copy its ANSI color number
- Scroll to move through the palette

## Development

```bash
go run .
```

## License

MIT
