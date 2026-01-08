SHELL := /bin/bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c


.PHONY: release
release:
	@ : "$${GITHUB_TOKEN:?GITHUB_TOKEN missing}"
	@ : "$${HOMEBREW_TAP_GITHUB_TOKEN:?HOMEBREW_TAP_GITHUB_TOKEN missing}"
	@ goreleaser release --clean

.PHONY: release-dry
release-dry:
	@ : "$${GITHUB_TOKEN:?GITHUB_TOKEN missing}"
	@ : "$${HOMEBREW_TAP_GITHUB_TOKEN:?HOMEBREW_TAP_GITHUB_TOKEN missing}"
	@ goreleaser release --clean --skip=publish
