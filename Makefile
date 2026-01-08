SHELL := /bin/bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
ENV_FILE := .env
LOAD_ENV := set -a; [ -f "$(ENV_FILE)" ] && source "$(ENV_FILE)"; set +a

.PHONY: release
release:
	@$(LOAD_ENV)
	@ : "$${GITHUB_TOKEN:?GITHUB_TOKEN missing}"
	@ : "$${HOMEBREW_TAP_GITHUB_TOKEN:?HOMEBREW_TAP_GITHUB_TOKEN missing}"
	@ goreleaser release --clean

.PHONY: release-dry
release-dry:
	@$(LOAD_ENV)
	@ : "$${GITHUB_TOKEN:?GITHUB_TOKEN missing}"
	@ : "$${HOMEBREW_TAP_GITHUB_TOKEN:?HOMEBREW_TAP_GITHUB_TOKEN missing}"
	@ goreleaser release --clean --skip=publish
