# Usage

Run the explorer:

```bash
ansicube
```

If you installed with `go install`, the command is `ansiCube` instead.

It opens full screen. Your terminal is restored when you quit.

## Controls

| Input | Action |
| --- | --- |
| Left click | Copy the tile's color number to the clipboard |
| Mouse wheel, ++up++ / ++down++, ++j++ / ++k++ | Scroll |
| ++page-up++ / ++page-down++ / ++space++ | Scroll one page |
| ++home++ / ++end++ | Jump to the top or bottom |
| ++q++ | Quit, after confirmation |
| ++ctrl+c++ | Quit immediately |

When you press ++q++, the status line asks you to confirm. Press ++y++ or
++enter++ to quit, or ++n++ or ++esc++ to keep going.

![The quit confirmation prompt](assets/screenshots/quit-prompt.png)

## Copying colors

Clicking a tile copies its number, for example `208`, and the status line shows
a swatch of the color you copied.

If the native clipboard is unavailable, as on a remote machine over SSH or on
Linux without a clipboard tool, ANSI Cube sends the text with the OSC 52 escape
sequence instead. The toast turns red to tell you. Your terminal must allow
OSC 52 clipboard writes for this to work.

## Flags

| Flag | Description |
| --- | --- |
| `--version` | Print the version and exit |
| `-h`, `--help` | Show help |

## Colors 0–15

The first 16 colors are defined by your terminal theme, so they look different
from one terminal to the next. Colors 16–255 are fixed by the xterm standard and
look the same everywhere.
