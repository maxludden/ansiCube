SHELL := /bin/bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c

.DEFAULT_GOAL := check

.PHONY: build
build:
	@ go build -trimpath -o ansicube .

.PHONY: run
run:
	@ go run .

.PHONY: test
test:
	@ go test -race ./...

.PHONY: lint
lint:
	@ go vet ./...
	@ go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest run

.PHONY: vuln
vuln:
	@ go run golang.org/x/vuln/cmd/govulncheck@latest ./...

.PHONY: check
check: test lint vuln

.PHONY: screenshots
screenshots: build
	@ mkdir -p docs/assets/screenshots
	@ for tape in docs/tapes/*.tape; do vhs "$$tape"; done

.PHONY: docs
docs: screenshots
	@ uvx zensical@0.1.3 build --clean --strict

.PHONY: docs-serve
docs-serve:
	@ uvx zensical@0.1.3 serve

# Print the release notes for VERSION, e.g. `make release-notes VERSION=0.2.0`.
.PHONY: release-notes
release-notes:
	@ scripts/release-notes.sh "$${VERSION:?set VERSION, e.g. VERSION=0.2.0}"

.PHONY: release
release:
	@ : "$${GITHUB_TOKEN:?GITHUB_TOKEN missing}"
	@ : "$${HOMEBREW_TAP_GITHUB_TOKEN:?HOMEBREW_TAP_GITHUB_TOKEN missing}"
	@ goreleaser release --clean

# Local snapshot build; needs no tokens and publishes nothing.
.PHONY: release-dry
release-dry:
	@ goreleaser release --clean --snapshot --skip=publish
