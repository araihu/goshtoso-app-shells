#!/usr/bin/env bash
set -euo pipefail

checker=scripts/check-araihu-assets-workflow.sh
source_workflow=.github/workflows/araihu-assets.yml
tmp=$(mktemp -d)
trap 'rm -rf -- "$tmp"' EXIT

bash "$checker" "$source_workflow" >/dev/null

mutation=0
expect_reject() {
  mutation=$((mutation + 1))
  if bash "$checker" "$1" "${2:-.dagger/src/index.ts}" \
    "${3:-scripts/dagger/assets-validate.sh}" >/dev/null 2>&1; then
    echo "asset workflow mutation $mutation was accepted" >&2
    exit 1
  fi
}

sed 's#--provider-event="$GITHUB_EVENT_PATH"#--payload="$GITHUB_EVENT_PATH"#' \
  "$source_workflow" > "$tmp/no-provider-event.yml"
expect_reject "$tmp/no-provider-event.yml"

sed 's#actions/github-script@3a2844b7e9c422d3c10d287c895573f7108da1b3#actions/github-script@v9#' \
  "$source_workflow" > "$tmp/mutable-action.yml"
expect_reject "$tmp/mutable-action.yml"

sed '/github\.paginate(github\.rest\.issues\.listLabelsForRepo/d' \
  "$source_workflow" > "$tmp/no-pagination.yml"
expect_reject "$tmp/no-pagination.yml"

sed '/--event-name="\$EVENT_NAME"/d' "$source_workflow" > "$tmp/no-event-name.yml"
expect_reject "$tmp/no-event-name.yml"

sed 's/This workflow never auto-merges\./Manual review required./' \
  "$source_workflow" > "$tmp/no-merge-policy.yml"
expect_reject "$tmp/no-merge-policy.yml"

cp "$source_workflow" "$tmp/host-jq.yml"
printf '%s\n' '          jq --version' >> "$tmp/host-jq.yml"
expect_reject "$tmp/host-jq.yml"

cp "$source_workflow" "$tmp/host-gh.yml"
printf '%s\n' '          gh label list' >> "$tmp/host-gh.yml"
expect_reject "$tmp/host-gh.yml"

cp "$source_workflow" "$tmp/host-absolute-jq.yml"
printf '%s\n' '          /usr/local/bin/jq --version' >> "$tmp/host-absolute-jq.yml"
expect_reject "$tmp/host-absolute-jq.yml"

cp "$source_workflow" "$tmp/auto-merge.yml"
printf '%s\n' '          gh pr merge --auto' >> "$tmp/auto-merge.yml"
expect_reject "$tmp/auto-merge.yml"

sed 's/\.withExec(\["bash", "scripts\/dagger\/assets-validate.sh"\])/\.withSecretVariable("EARLY_SECRET", githubToken)/' \
  .dagger/src/index.ts > "$tmp/no-validation.ts"
expect_reject "$source_workflow" "$tmp/no-validation.ts"

sed 's/\.withFile("\/run\/assets-provider-event.json", providerEvent)/\.withFile("\/run\/assets-payload.json", providerEvent)/' \
  .dagger/src/index.ts > "$tmp/no-provider-file.ts"
expect_reject "$source_workflow" "$tmp/no-provider-file.ts"

awk '!changed && index($0, "araihu/goshtoso-app-shells") {
       sub("araihu/goshtoso-app-shells", "araihu/other"); changed=1
     } { print }' scripts/dagger/assets-validate.sh > "$tmp/wrong-repository.sh"
expect_reject "$source_workflow" .dagger/src/index.ts "$tmp/wrong-repository.sh"

awk '!removed && index($0, "araihu/goshtoso-app-shells") { removed=1; next }
     { print }' scripts/dagger/assets-validate.sh > "$tmp/missing-repository-check.sh"
expect_reject "$source_workflow" .dagger/src/index.ts "$tmp/missing-repository-check.sh"

echo "$mutation asset workflow mutations passed"
