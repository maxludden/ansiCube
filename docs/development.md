# Development

Requires Go 1.26 or newer.

```bash
git clone https://github.com/maxludden/ansiCube
cd ansiCube
make run
```

## Make targets

| Target | What it does |
| --- | --- |
| `make run` | `go run .` |
| `make build` | Builds `./ansicube` |
| `make test` | `go test -race ./...` |
| `make lint` | `go vet` and golangci-lint |
| `make vuln` | govulncheck |
| `make check` | `test`, `lint`, and `vuln` (the default) |
| `make screenshots` | Renders the documentation screenshots with VHS |
| `make docs` | Renders screenshots and builds the site into `site/` |
| `make docs-serve` | Serves the docs locally with live reload |
| `make release-dry` | Local GoReleaser snapshot build; publishes nothing |

`make screenshots` needs [VHS](https://github.com/charmbracelet/vhs)
(`brew install vhs`). The docs targets need [uv](https://docs.astral.sh/uv/).

## Code layout

| File | Purpose |
| --- | --- |
| `main.go` | Flag parsing, version, and program start-up |
| `palette.go` | The 256-color layout, xterm RGB values, and label contrast |
| `layout.go` | Tile widths, scrolling bounds, and mouse hit-testing |
| `model.go` | Bubble Tea model: input handling and rendering |

Rendering and hit-testing share the same geometry functions, so a click always
lands on the tile that was drawn there. `View` never mutates the model.

## Documentation

The site is built with [Zensical](https://zensical.org) from `docs/` and
`zensical.toml`. The **docs** GitHub Actions workflow renders the screenshots
with VHS, builds the site, and deploys it to GitHub Pages on every push to
`main` that touches the code or docs.

## Releasing

1. Move the `Unreleased` entries in `CHANGELOG.md` under a new version heading,
   such as `## [0.3.0] - YYYY-MM-DD`, and update the comparison links at the
   bottom.
2. Commit, then tag and push:

    ```bash
    git tag -a v0.3.0 -m "v0.3.0"
    git push origin main v0.3.0
    ```

The **release** workflow runs the tests, takes the release notes from that
version's section of `CHANGELOG.md`, and publishes the binaries and the
Homebrew cask with GoReleaser. It fails if the changelog has no section for
the tag.
