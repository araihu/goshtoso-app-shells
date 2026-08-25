#!/usr/bin/env bash
set -euo pipefail

workflow=${1:-.github/workflows/araihu-assets.yml}
dagger_source=${2:-.dagger/src/index.ts}
validator=${3:-scripts/dagger/assets-validate.sh}
github_script='actions/github-script@3a2844b7e9c422d3c10d287c895573f7108da1b3 # v9.0.0'

test -f "$workflow"
test -f "$dagger_source"
test -f "$validator"
grep -F -- '--provider-event="$GITHUB_EVENT_PATH"' "$workflow" >/dev/null
grep -F -- '--event-name="$EVENT_NAME"' "$workflow" >/dev/null

if grep -E '(^|[^[:alnum:]_.-])(jq|gh)([^[:alnum:]_.-]|$)' "$workflow"; then
  echo 'Asset workflow must not invoke jq or gh on the runner host' >&2
  exit 1
fi

test "$(grep -cF "$github_script" "$workflow")" -eq 1
grep -F 'github.paginate(github.rest.issues.listLabelsForRepo' "$workflow" >/dev/null
grep -F 'const selected = wanted.filter((label) => available.has(label))' "$workflow" >/dev/null
grep -F 'core.setOutput("labels", selected.join("\n"))' "$workflow" >/dev/null

if grep -E 'issues\.createLabel|enable-pull-request-automerge|pulls\.merge|gh pr merge' "$workflow"; then
  echo 'Asset workflow must neither create labels nor merge pull requests' >&2
  exit 1
fi

dagger_line=$(grep -nF 'dagger call assets-update' "$workflow" | cut -d: -f1)
app_token_line=$(grep -nF 'name: Create selected-repository App token' "$workflow" | cut -d: -f1)
labels_line=$(grep -nF "$github_script" "$workflow" | cut -d: -f1)
test -n "$dagger_line"
test -n "$app_token_line"
test -n "$labels_line"
test "$dagger_line" -lt "$app_token_line"
test "$app_token_line" -lt "$labels_line"

grep -F '.withFile("/run/assets-provider-event.json", providerEvent)' "$dagger_source" >/dev/null
validate_line=$(grep -nF '.withExec(["bash", "scripts/dagger/assets-validate.sh"])' "$dagger_source" | cut -d: -f1)
secret_line=$(grep -nF '.withSecretVariable("GH_TOKEN", githubToken)' "$dagger_source" | cut -d: -f1)
test -n "$validate_line"
test -n "$secret_line"
test "$validate_line" -lt "$secret_line"
test "$(grep -cF '.repository.full_name == "araihu/goshtoso-app-shells"' "$validator")" -eq 2

grep -F 'This workflow never auto-merges.' "$workflow" >/dev/null
grep -F 'peter-evans/create-pull-request@5f6978faf089d4d20b00c7766989d076bb2fc7f1 # v8' "$workflow" >/dev/null

echo 'Arai Hu asset workflow host contract is valid'
