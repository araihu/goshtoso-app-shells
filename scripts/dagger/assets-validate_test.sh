#!/usr/bin/env bash
set -euo pipefail

tmp=$(mktemp -d)
trap 'rm -rf -- "$tmp"' EXIT

payload="$tmp/payload.json"
cat > "$payload" <<'JSON'
{
  "assets_repository": "araihu/assets",
  "assets_revision": "0123456789abcdef0123456789abcdef01234567",
  "release": "v1.2.3",
  "release_url": "https://github.com/araihu/assets/releases/download/v1.2.3/araihu-assets-v1.2.3.tar.gz",
  "release_sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "release_json_sha256": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
}
JSON

repository_event="$tmp/repository-event.json"
jq --null-input --slurpfile payload "$payload" \
  '{action: "araihu-assets-released", client_payload: $payload[0],
    repository: {full_name: "araihu/goshtoso-app-shells"}}' > "$repository_event"

ASSETS_PROVIDER_EVENT_PATH="$repository_event" GITHUB_EVENT_NAME=repository_dispatch \
  ASSETS_CONTEXT_DIR="$tmp/context" \
  bash scripts/dagger/assets-validate.sh
test "$(<"$tmp/context/release")" = v1.2.3

workflow_event="$tmp/workflow-event.json"
jq --null-input --slurpfile payload "$payload" \
  '{inputs: $payload[0], repository: {full_name: "araihu/goshtoso-app-shells"}}' > "$workflow_event"
ASSETS_PROVIDER_EVENT_PATH="$workflow_event" GITHUB_EVENT_NAME=workflow_dispatch \
  ASSETS_CONTEXT_DIR="$tmp/workflow-context" \
  bash scripts/dagger/assets-validate.sh
test "$(<"$tmp/workflow-context/release")" = v1.2.3

jq '.client_payload.unexpected = true' "$repository_event" > "$tmp/extra.json"
if ASSETS_PROVIDER_EVENT_PATH="$tmp/extra.json" GITHUB_EVENT_NAME=repository_dispatch \
  ASSETS_CONTEXT_DIR="$tmp/extra-context" \
  bash scripts/dagger/assets-validate.sh; then
  echo 'payload with an unknown key was accepted' >&2
  exit 1
fi

jq '.inputs.release_url = "https://example.invalid/archive.tar.gz"' "$workflow_event" > "$tmp/url.json"
if ASSETS_PROVIDER_EVENT_PATH="$tmp/url.json" GITHUB_EVENT_NAME=workflow_dispatch \
  ASSETS_CONTEXT_DIR="$tmp/url-context" \
  bash scripts/dagger/assets-validate.sh; then
  echo 'payload with a noncanonical URL was accepted' >&2
  exit 1
fi

jq '.action = "untrusted-event"' "$repository_event" > "$tmp/action.json"
if ASSETS_PROVIDER_EVENT_PATH="$tmp/action.json" GITHUB_EVENT_NAME=repository_dispatch \
  ASSETS_CONTEXT_DIR="$tmp/action-context" \
  bash scripts/dagger/assets-validate.sh; then
  echo 'repository dispatch with wrong action was accepted' >&2
  exit 1
fi

if ASSETS_PROVIDER_EVENT_PATH="$repository_event" GITHUB_EVENT_NAME=pull_request \
  ASSETS_CONTEXT_DIR="$tmp/unsupported-context" \
  bash scripts/dagger/assets-validate.sh; then
  echo 'unsupported provider event was accepted' >&2
  exit 1
fi

for event_name in repository_dispatch workflow_dispatch; do
  case "$event_name" in
    repository_dispatch) valid_event=$repository_event ;;
    workflow_dispatch) valid_event=$workflow_event ;;
  esac

  jq 'del(.repository)' "$valid_event" > "$tmp/$event_name-repository-missing.json"
  if ASSETS_PROVIDER_EVENT_PATH="$tmp/$event_name-repository-missing.json" \
    GITHUB_EVENT_NAME="$event_name" ASSETS_CONTEXT_DIR="$tmp/$event_name-missing-context" \
    bash scripts/dagger/assets-validate.sh; then
    echo "$event_name without repository identity was accepted" >&2
    exit 1
  fi

  jq '.repository.full_name = "araihu/other"' "$valid_event" \
    > "$tmp/$event_name-repository-wrong.json"
  if ASSETS_PROVIDER_EVENT_PATH="$tmp/$event_name-repository-wrong.json" \
    GITHUB_EVENT_NAME="$event_name" ASSETS_CONTEXT_DIR="$tmp/$event_name-wrong-context" \
    bash scripts/dagger/assets-validate.sh; then
    echo "$event_name with wrong repository identity was accepted" >&2
    exit 1
  fi

  jq '.repository.full_name = 42' "$valid_event" \
    > "$tmp/$event_name-repository-type.json"
  if ASSETS_PROVIDER_EVENT_PATH="$tmp/$event_name-repository-type.json" \
    GITHUB_EVENT_NAME="$event_name" ASSETS_CONTEXT_DIR="$tmp/$event_name-type-context" \
    bash scripts/dagger/assets-validate.sh; then
    echo "$event_name with non-string repository identity was accepted" >&2
    exit 1
  fi
done

echo '8 assets payload validation regressions passed'
