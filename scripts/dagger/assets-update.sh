#!/usr/bin/env bash
set -euo pipefail

: "${GH_TOKEN:?GH_TOKEN is required}"
: "${CI_RUN_NONCE:?CI_RUN_NONCE is required}"

context=${ASSETS_CONTEXT_DIR:-/run/assets-context}
assets_repository=$(<"$context/repository")
assets_revision=$(<"$context/revision")
release=$(<"$context/release")
release_url=$(<"$context/url")
release_sha256=$(<"$context/archive-sha256")
release_json_sha256=$(<"$context/release-json-sha256")

github_api() {
  curl --fail --silent --show-error --location \
    --proto '=https' \
    --header 'Accept: application/vnd.github+json' \
    --header "Authorization: Bearer ${GH_TOKEN}" \
    --header 'X-GitHub-Api-Version: 2022-11-28' \
    "$1"
}

object=$(github_api "https://api.github.com/repos/araihu/assets/git/ref/tags/${release}" |
  jq -er '.object.type + " " + .object.sha')
read -r object_type object_sha <<< "$object"
if [[ "$object_type" = tag ]]; then
  object=$(github_api "https://api.github.com/repos/araihu/assets/git/tags/${object_sha}" |
    jq -er '.object.type + " " + .object.sha')
  read -r object_type object_sha <<< "$object"
fi
test "$object_type" = commit
test "$object_sha" = "$assets_revision"

archive="/tmp/araihu-assets-${release}.tar.gz"
extracted="/tmp/araihu-assets-${release}"
curl --fail --location --silent --show-error --retry 3 --proto '=https' \
  --output "$archive" "$release_url"
printf '%s  %s\n' "$release_sha256" "$archive" | sha256sum --check --strict
while IFS= read -r member; do
  case "$member" in
    /*|../*|*/../*|*/..) echo "unsafe archive member: $member" >&2; exit 1 ;;
  esac
done < <(tar --list --gzip --file "$archive")
if tar --list --verbose --gzip --file "$archive" | grep --extended-regexp --quiet '^[lh]'; then
  echo 'release archive contains a link' >&2
  exit 1
fi
mkdir "$extracted"
tar --extract --gzip --file "$archive" --directory "$extracted" \
  --no-same-owner --no-same-permissions

update_args=(
  -release-dir "$extracted"
  -assets-repository "$assets_repository"
  -assets-revision "$assets_revision"
  -release "$release"
  -release-url "$release_url"
  -release-sha256 "$release_sha256"
  -release-json-sha256 "$release_json_sha256"
)
go run ./cmd/araihu-assets-update "${update_args[@]}"

allowlisted=(
  araihu-assets.json
  componentdocshell/assets/araihu.css
  componentdocshell/assets/goshtoso-favicon.svg
  componentdocshell/assets/goshtoso-logo.svg
  componentdocshell/assets/goshtoso-mark-reverse.svg
  componentdocshell/assets/goshtoso-mark.svg
)
first=$(mktemp -d)
for path in "${allowlisted[@]}"; do
  mkdir -p "$first/$(dirname "$path")"
  cp "$path" "$first/$path"
done
go run ./cmd/araihu-assets-update "${update_args[@]}"
for path in "${allowlisted[@]}"; do
  cmp "$first/$path" "$path"
done

go test ./internal/araihuassets ./cmd/araihu-assets-update -count=1
mkdir /out
for path in "${allowlisted[@]}"; do
  mkdir -p "/out/$(dirname "$path")"
  cp "$path" "/out/$path"
done
