#!/usr/bin/env bash
set -euo pipefail

tmp=$(mktemp -d)
trap 'rm -rf -- "$tmp"' EXIT

valid="$tmp/valid.json"
cat > "$valid" <<'JSON'
{
  "assets_repository": "araihu/assets",
  "assets_revision": "0123456789abcdef0123456789abcdef01234567",
  "release": "v1.2.3",
  "release_url": "https://github.com/araihu/assets/releases/download/v1.2.3/araihu-assets-v1.2.3.tar.gz",
  "release_sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "release_json_sha256": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
}
JSON

ASSETS_PAYLOAD_PATH="$valid" ASSETS_CONTEXT_DIR="$tmp/context" \
  bash scripts/dagger/assets-validate.sh
test "$(<"$tmp/context/release")" = v1.2.3

jq '.unexpected = true' "$valid" > "$tmp/extra.json"
if ASSETS_PAYLOAD_PATH="$tmp/extra.json" ASSETS_CONTEXT_DIR="$tmp/extra-context" \
  bash scripts/dagger/assets-validate.sh; then
  echo 'payload with an unknown key was accepted' >&2
  exit 1
fi

jq '.release_url = "https://example.invalid/archive.tar.gz"' "$valid" > "$tmp/url.json"
if ASSETS_PAYLOAD_PATH="$tmp/url.json" ASSETS_CONTEXT_DIR="$tmp/url-context" \
  bash scripts/dagger/assets-validate.sh; then
  echo 'payload with a noncanonical URL was accepted' >&2
  exit 1
fi

echo 'assets payload validation regressions passed'
