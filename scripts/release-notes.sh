#!/usr/bin/env bash
# Prints the CHANGELOG.md section for one version, without its heading.
# Usage: scripts/release-notes.sh 0.2.0   (a leading "v" is ignored)
# Exits non-zero if the changelog has no section for that version.
set -euo pipefail

version="${1:?usage: $0 VERSION}"
version="${version#v}"
changelog="${2:-CHANGELOG.md}"

notes="$(awk -v ver="$version" '
  # A version heading looks like: ## [0.2.0] - 2026-11-24
  /^## \[/ {
    if (found) exit
    if (index($0, "## [" ver "]") == 1) { found = 1; next }
  }
  # Stop at the link reference definitions at the bottom of the file.
  found && /^\[[^]]+\]: / { exit }
  found { print }
' "$changelog")"

if [[ -z "${notes//[[:space:]]/}" ]]; then
  echo "error: no entries for version $version in $changelog" >&2
  exit 1
fi
printf '%s\n' "$notes"
