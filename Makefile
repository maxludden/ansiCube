SHELL := /bin/bash

.PHONY: release
release:
	@set -a; source .env; set +a; goreleaser release --clean

.PHONY: release-dry
release-dry:
	@set -a; source .env; set +a; goreleaser release --clean --skip=publish
