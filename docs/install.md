# Installation

=== "Homebrew"

    ```bash
    brew install maxludden/tap/ansicube
    ```

=== "Go"

    Requires Go 1.26 or newer.

    ```bash
    go install github.com/maxludden/ansiCube@latest
    ```

    `go install` names the binary after the module path, so it is installed as
    `ansiCube` in `$(go env GOPATH)/bin`.

=== "Prebuilt binary"

    Download the archive for your platform from the
    [GitHub Releases](https://github.com/maxludden/ansiCube/releases) page,
    extract it, and move `ansicube` somewhere on your `PATH`.

    Builds are published for macOS, Linux, and Windows on `amd64` and `arm64`.
    Each release includes a `checksums.txt` file.

## Check the install

```bash
ansicube --version
```

## Requirements

- A terminal with 256-color support. Nearly every modern terminal qualifies.
- Mouse reporting, to click tiles. Keyboard scrolling works without it.
- On Linux, `xclip`, `xsel`, or `wl-clipboard` for the native clipboard.
  Without one, ANSI Cube falls back to OSC 52, which most modern terminals
  support.
