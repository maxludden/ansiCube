# Changelog

All notable changes to this project are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and the project uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.0] - 2026-11-24

### Added

- Keyboard scrolling: `↑`/`↓`, `j`/`k`, `PgUp`/`PgDn`/`space`, and `Home`/`End`.
- `--version` flag. Release builds report the tag version; `go install` builds
  report the module version.
- OSC 52 clipboard fallback when the native clipboard is unavailable (for
  example over SSH or on Linux without `xclip`), with an error toast.
- A color swatch in the "copied" toast, so dark colors stay readable.
- Documentation site built with Zensical and deployed to GitHub Pages, with
  screenshots rendered by VHS in CI.
- CI workflow running vet, race-enabled tests on Linux, macOS, and Windows,
  golangci-lint, govulncheck, and a `go mod tidy` check.
- Unit tests for the palette layout, hit-testing, rendering, and input handling.
- This changelog.

### Changed

- `q` now asks for confirmation before quitting. `Ctrl+C` still quits at once.
- Upgraded to Bubble Tea v2 and Lip Gloss v2. Requires Go 1.26 or newer.
- Tile label color is now chosen by WCAG contrast against each color's real
  xterm RGB value, so some labels switch between black and white.
  Labels use true black and white, so themes that remap colors 0 and 15
  no longer wash them out.
- The toast and quit prompt share one fixed status line, so the grid no longer
  shifts when they appear.
- The footer hint uses gray (color 8) instead of black, so it is visible on
  dark terminals.
- Rendering no longer mutates state; layout and click hit-testing share a
  single source of geometry.
- Release builds are reproducible (`-trimpath`, commit timestamps), and release
  notes come from this changelog.
- The Homebrew cask removes the macOS quarantine flag after install.

### Removed

- The `g` key, which previously opened the quit prompt.
- The compiled binary that was committed to the repository.

### Fixed

- The release workflow passed tokens as literal strings instead of secrets and
  pinned a Go version too old to build the module.
- Clipboard write errors were silently ignored.
- A slow clipboard backend could freeze the UI; writes now run asynchronously
  and are serialized so an older click never overwrites a newer one.
- The status line and footer are truncated in narrow windows instead of
  wrapping and shifting the palette.
- Errors on exit are printed to stderr with a trailing newline.
- The README referenced a missing image.

## [0.1.5] - 2026-01-08

### Fixed

- `.gitignore` entry for the `dist` directory.

### Changed

- Removed stale distribution files and added a VS Code task configuration.

## [0.1.0] - 2026-01-08

Releases 0.1.1 through 0.1.4 (same day) only adjusted the release pipeline.

### Added

- Initial release: an interactive ANSI 256-color explorer showing the system
  colors, the 6×6×6 color cube, and the grayscale ramp as clickable tiles.
- Click a tile to copy its color number to the clipboard.
- Responsive layout with mouse-wheel scrolling.
- GoReleaser builds for macOS, Linux, and Windows, and a Homebrew tap.

[Unreleased]: https://github.com/maxludden/ansiCube/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/maxludden/ansiCube/compare/v0.1.5...v0.2.0
[0.1.5]: https://github.com/maxludden/ansiCube/compare/v0.1.4...v0.1.5
[0.1.0]: https://github.com/maxludden/ansiCube/releases/tag/v0.1.0
