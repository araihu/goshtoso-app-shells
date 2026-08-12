#!/usr/bin/env bash
set -euo pipefail

payload=${ASSETS_PAYLOAD_PATH:-/run/assets-payload.json}
context=${ASSETS_CONTEXT_DIR:-/run/assets-context}
jq --exit-status '
  keys == [
    "assets_repository",
    "assets_revision",
    "release",
    "release_json_sha256",
    "release_sha256",
    "release_url"
  ]
' "$payload" >/dev/null

assets_repository=$(jq -er '.assets_repository | select(type == "string")' "$payload")
assets_revision=$(jq -er '.assets_revision | select(type == "string")' "$payload")
release=$(jq -er '.release | select(type == "string")' "$payload")
release_url=$(jq -er '.release_url | select(type == "string")' "$payload")
release_sha256=$(jq -er '.release_sha256 | select(type == "string")' "$payload")
release_json_sha256=$(jq -er '.release_json_sha256 | select(type == "string")' "$payload")

test "$assets_repository" = araihu/assets
[[ "$assets_revision" =~ ^[0-9a-f]{40}$ ]]
[[ "$release" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]
[[ "$release_sha256" =~ ^[0-9a-f]{64}$ ]]
[[ "$release_json_sha256" =~ ^[0-9a-f]{64}$ ]]
expected_url="https://github.com/araihu/assets/releases/download/${release}/araihu-assets-${release}.tar.gz"
test "$release_url" = "$expected_url"

mkdir "$context"
printf '%s\n' "$assets_repository" > "$context/repository"
printf '%s\n' "$assets_revision" > "$context/revision"
printf '%s\n' "$release" > "$context/release"
printf '%s\n' "$release_url" > "$context/url"
printf '%s\n' "$release_sha256" > "$context/archive-sha256"
printf '%s\n' "$release_json_sha256" > "$context/release-json-sha256"
