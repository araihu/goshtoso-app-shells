#!/usr/bin/env bash
set -euo pipefail

action='dagger/dagger-for-github@456fc3af63a2ba6f9789af9c55045b459115541b # v8.3.0'

grep -F '"engineVersion": "v0.21.8"' dagger.json
grep -F '"@dagger.io/dagger": "./sdk"' .dagger/package.json
grep -F '"typescript": "6.0.3"' .dagger/package.json
if grep -F '"devDependencies"' .dagger/package.json; then
  echo 'TypeScript must remain available to the production-only Dagger runtime install' >&2
  exit 1
fi
grep -F '"@opentelemetry/core": "2.10.0"' .dagger/package.json
grep -F '"@opentelemetry/propagator-jaeger": "2.9.0"' .dagger/package.json
grep -F '"adm-zip": "0.6.0"' .dagger/package.json
grep -F '"uuid": "14.0.1"' .dagger/package.json
grep -A4 -F '"node_modules/@dagger.io/dagger"' .dagger/package-lock.json | grep -F '"resolved": "sdk"'
grep -A4 -F '"node_modules/typescript"' .dagger/package-lock.json | grep -F '"version": "6.0.3"'
grep -F '/.dagger/sdk/' .gitignore
grep -F 'npm --prefix .dagger audit --package-lock-only --omit=dev --audit-level=high' scripts/check-dagger-sdk-audit.sh
if grep -E '(npm .* install|tsc|node .*\.dagger/sdk)' scripts/check-dagger-sdk-audit.sh; then
  echo 'SDK audit must not install or typecheck a handwritten generated-SDK substitute' >&2
  exit 1
fi
if git rev-parse --is-inside-work-tree >/dev/null 2>&1 &&
  test -n "$(git ls-files '.dagger/sdk/**')"; then
  echo 'Generated Dagger SDK files must not be committed' >&2
  exit 1
fi
grep -F 'golang:1.26.5-bookworm@sha256:53eeac89074db483fdf0ab3be1df32bf6e47562263d2d0d6baa7f26acb4957dd' .dagger/src/index.ts
grep -F 'ghcr.io/jqlang/jq:1.8.2@sha256:b9c68867e5766576263a222e91db3de422d802069c7af70440e667a95344e486' .dagger/src/index.ts
grep -F 'const PLAYWRIGHT_VERSION = "v0.6100.0"' .dagger/src/index.ts
grep -F '@func({ cache: "never" })' .dagger/src/index.ts
grep -F 'dag.cacheVolume(`araihu-ci-v1-goshtoso-app-shells-${trustDomain}-gomod`)' .dagger/src/index.ts
grep -F 'COMPONENTDOCSHELL_E2E", "1"' .dagger/src/index.ts
grep -F '["go", "test", "./example/e2e", "-count=1", "-v"]' .dagger/src/index.ts
grep -F 'github.com/a-h/templ/cmd/templ@v0.3.1020' .dagger/src/index.ts
grep -F 'github.com/mxschmitt/playwright-go/cmd/playwright@${PLAYWRIGHT_VERSION}' .dagger/src/index.ts
grep -F '"install-deps",' .dagger/src/index.ts
grep -F '"install",' .dagger/src/index.ts
grep -F '"chromium",' .dagger/src/index.ts
test "$(grep -cF 'withMountedCache(' .dagger/src/index.ts)" -eq 3
go_container=$(sed -n '/^  private goContainer(/,/^  private browserContainer/p' .dagger/src/index.ts)
browser_container=$(sed -n '/^  private browserContainer(/,/^  private isPullRequestTrustDomain/p' .dagger/src/index.ts)
for block in "$go_container" "$browser_container"; do
  trust_line=$(printf '%s\n' "$block" | grep -nF 'if (this.isPullRequestTrustDomain(trustDomain))' | cut -d: -f1)
  cache_line=$(printf '%s\n' "$block" | grep -nF 'withMountedCache(' | head -n 1 | cut -d: -f1)
  test -n "$trust_line"
  test -n "$cache_line"
  test "$trust_line" -lt "$cache_line"
  printf '%s\n' "$block" | grep -A2 -F 'if (this.isPullRequestTrustDomain(trustDomain))' | grep -F 'return container'
done
grep -F 'return trustDomain === "fork" || trustDomain === "internal"' .dagger/src/index.ts
grep -F 'payload: File' .dagger/src/index.ts
grep -F 'githubToken: Secret' .dagger/src/index.ts
assets_block=$(sed -n '/^  assetsUpdate(/,/^  private goContainer/p' .dagger/src/index.ts)
validate_line=$(printf '%s\n' "$assets_block" | grep -nF 'scripts/dagger/assets-validate.sh' | cut -d: -f1)
secret_line=$(printf '%s\n' "$assets_block" | grep -nF 'withSecretVariable("GH_TOKEN"' | cut -d: -f1)
nonce_line=$(printf '%s\n' "$assets_block" | grep -nF 'withEnvVariable("CI_RUN_NONCE"' | cut -d: -f1)
test "$validate_line" -lt "$secret_line"
test "$validate_line" -lt "$nonce_line"

for workflow in .github/workflows/*.yml; do
  if grep -i -E 'coderabbit|code rabbit' "$workflow"; then
    echo "CodeRabbit must remain absent: $workflow" >&2
    exit 1
  fi
  if grep -F 'verb: call' "$workflow"; then
    echo "Dagger action must be installer-only: $workflow" >&2
    exit 1
  fi
done

grep -F 'types: [araihu-assets-released]' .github/workflows/araihu-assets.yml
grep -F -- '--payload=.dagger-input/assets-payload.json' .github/workflows/araihu-assets.yml
if grep -E -- '--(assets-repository|assets-revision|release|release-url|release-sha256|release-json-sha256)=' .github/workflows/araihu-assets.yml; then
  echo 'External asset identity must travel only through the payload file' >&2
  exit 1
fi
grep -F 'This workflow never auto-merges.' .github/workflows/araihu-assets.yml
test "$(grep -cF "$action" .github/workflows/araihu-assets.yml)" -eq 1
grep -B2 -F "$action" .github/workflows/araihu-assets.yml | grep -F "if: runner.environment == 'github-hosted'"

test "$(grep -cF "$action" .github/workflows/ci.yml)" -eq 2
grep -B2 -F "$action" .github/workflows/ci.yml | grep -F "if: runner.environment == 'github-hosted'"
grep -F 'dagger call ci' .github/workflows/ci.yml
grep -F 'dagger call browser' .github/workflows/ci.yml
grep -F '  push:' .github/workflows/ci.yml
grep -F '  pull_request:' .github/workflows/ci.yml
grep -F '  verify:' .github/workflows/ci.yml
grep -F '  browser:' .github/workflows/ci.yml
grep -A2 -F '  browser:' .github/workflows/ci.yml | grep -F 'runs-on: ubuntu-latest'
grep -A3 -F '  browser:' .github/workflows/ci.yml | grep -F 'timeout-minutes: 15'
test "$(grep -cF -- '--trust-domain="$TRUST_DOMAIN"' .github/workflows/ci.yml)" -eq 2
if grep -F 'actions/cache@' .github/workflows/ci.yml; then
  echo 'Playwright and Go caching must remain inside trust-aware Dagger functions' >&2
  exit 1
fi
grep -F 'name: CI' .github/workflows/ci.yml
grep -F "github.ref == 'refs/heads/main' && 'main'" .github/workflows/ci.yml
grep -F "'branch'" .github/workflows/ci.yml

test "$(grep -cF "$action" .github/workflows/hostinger-ci-benchmark.yml)" -eq 1
grep -F 'GitHub-hosted identical workload' .github/workflows/hostinger-ci-benchmark.yml
grep -F 'Hostinger identical workload' .github/workflows/hostinger-ci-benchmark.yml
test "$(grep -cF 'dagger call benchmark' .github/workflows/hostinger-ci-benchmark.yml)" -eq 2
test "$(grep -cF 'actual="$(dagger version' .github/workflows/hostinger-ci-benchmark.yml)" -eq 2
grep -F 'BENCHMARK_CACHE_STATE=warm' scripts/ci-hostinger-benchmark.sh
grep -F 'BENCHMARK_CACHE_STATE=cold' scripts/ci-hostinger-benchmark.sh

for workflow in .github/workflows/*.yml; do
  version_line=$(grep -nF 'actual="$(dagger version' "$workflow" | head -n 1 | cut -d: -f1)
  call_line=$(grep -nF 'dagger call' "$workflow" | head -n 1 | cut -d: -f1)
  test -n "$version_line"
  test -n "$call_line"
  test "$version_line" -lt "$call_line"
done

echo 'Dagger and GitHub workflow contracts are valid'
