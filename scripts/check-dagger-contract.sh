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
grep -F '    - hostinger-vps-pr' .github/actionlint.yaml
grep -F '    - hostinger-vps-trusted' .github/actionlint.yaml
if grep -Fx '    - hostinger-vps' .github/actionlint.yaml; then
  echo 'Generic Hostinger runner label remains in actionlint' >&2
  exit 1
fi
grep -F 'golang:1.27.0-bookworm@sha256:484ef6066fa69acb059fdfeda7ba2b8f7391f2ef6abc6f9b8411e669ebd56466' .dagger/src/index.ts
grep -F 'ghcr.io/jqlang/jq:1.8.2@sha256:b9c68867e5766576263a222e91db3de422d802069c7af70440e667a95344e486' .dagger/src/index.ts
grep -F 'const PLAYWRIGHT_VERSION = "v0.6201.1"' .dagger/src/index.ts
grep -F '@func({ cache: "never" })' .dagger/src/index.ts
grep -F 'dag.cacheVolume(`araihu-ci-v1-goshtoso-app-shells-${cacheNamespace}-gomod`)' .dagger/src/index.ts
grep -F 'dag.cacheVolume(`araihu-ci-v1-goshtoso-app-shells-${cacheNamespace}-gobuild`)' .dagger/src/index.ts
grep -F 'araihu-ci-v1-goshtoso-app-shells-${cacheNamespace}-playwright-${PLAYWRIGHT_VERSION}' .dagger/src/index.ts
grep -F 'COMPONENTDOCSHELL_E2E", "1"' .dagger/src/index.ts
grep -F '["go", "test", "./example/e2e", "-count=1", "-v"]' .dagger/src/index.ts
grep -F 'github.com/a-h/templ/cmd/templ@v0.3.1020' .dagger/src/index.ts
grep -F 'github.com/mxschmitt/playwright-go/cmd/playwright@${PLAYWRIGHT_VERSION}' .dagger/src/index.ts
grep -F '"install-deps",' .dagger/src/index.ts
grep -F '"install",' .dagger/src/index.ts
grep -F '"chromium",' .dagger/src/index.ts
test "$(grep -cF 'withMountedCache(' .dagger/src/index.ts)" -eq 3
grep -F 'const CACHE_NAMESPACES = [' .dagger/src/index.ts
grep -F '  "pr",' .dagger/src/index.ts
grep -F '  "trusted",' .dagger/src/index.ts
grep -F '  "branch-hosted",' .dagger/src/index.ts
if grep -E 'TRUST_DOMAINS|trustDomain|trust-domain|"fork"|"internal"|isPullRequestTrustDomain' .dagger/src/index.ts .github/workflows/*.yml; then
  echo 'Cache isolation still depends on workflow trust arguments or fork/internal guards' >&2
  exit 1
fi
grep -F 'providerEvent: File' .dagger/src/index.ts
grep -F 'eventName: string' .dagger/src/index.ts
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
grep -F -- '--provider-event="$GITHUB_EVENT_PATH"' .github/workflows/araihu-assets.yml
grep -F -- '--event-name="$EVENT_NAME"' .github/workflows/araihu-assets.yml
grep -F 'runs-on: [self-hosted, Linux, X64, hostinger-vps-trusted]' .github/workflows/araihu-assets.yml
grep -F -- '--cache-namespace=trusted' .github/workflows/araihu-assets.yml
if grep -E -- '--(assets-repository|assets-revision|release|release-url|release-sha256|release-json-sha256)=' .github/workflows/araihu-assets.yml; then
  echo 'External asset identity must travel only through the payload file' >&2
  exit 1
fi
grep -F 'This workflow never auto-merges.' .github/workflows/araihu-assets.yml
bash scripts/check-araihu-assets-workflow.sh
bash scripts/check-araihu-assets-workflow_test.sh
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
grep -F 'timeout-minutes: 15' .github/workflows/ci.yml
test "$(grep -cF "fromJSON('[\"self-hosted\",\"Linux\",\"X64\",\"hostinger-vps-pr\"]')" .github/workflows/ci.yml)" -eq 2
test "$(grep -cF "fromJSON('[\"self-hosted\",\"Linux\",\"X64\",\"hostinger-vps-trusted\"]')" .github/workflows/ci.yml)" -eq 2
test "$(grep -cF "github.ref == 'refs/heads/main'" .github/workflows/ci.yml)" -eq 4
test "$(grep -cF "'ubuntu-24.04'" .github/workflows/ci.yml)" -eq 2
test "$(grep -cF -- '--cache-namespace="$CACHE_NAMESPACE"' .github/workflows/ci.yml)" -eq 2
test "$(grep -cF "github.event_name == 'pull_request' && 'pr' ||" .github/workflows/ci.yml)" -eq 2
test "$(grep -cF "github.ref == 'refs/heads/main' && 'trusted' ||" .github/workflows/ci.yml)" -eq 2
test "$(grep -cF "'branch-hosted'" .github/workflows/ci.yml)" -eq 2
if grep -F 'actions/cache@' .github/workflows/ci.yml; then
  echo 'Playwright and Go caching must remain inside trust-aware Dagger functions' >&2
  exit 1
fi
grep -F 'name: CI' .github/workflows/ci.yml

test "$(grep -cF "$action" .github/workflows/hostinger-ci-benchmark.yml)" -eq 1
grep -F 'GitHub-hosted identical workload' .github/workflows/hostinger-ci-benchmark.yml
grep -F 'Hostinger identical workload' .github/workflows/hostinger-ci-benchmark.yml
test "$(grep -cF 'dagger call benchmark' .github/workflows/hostinger-ci-benchmark.yml)" -eq 2
test "$(grep -cF 'actual="$(dagger version' .github/workflows/hostinger-ci-benchmark.yml)" -eq 2
grep -F 'runs-on: [self-hosted, Linux, X64, hostinger-vps-trusted]' .github/workflows/hostinger-ci-benchmark.yml
grep -F -- '--cache-namespace=benchmark-hosted' .github/workflows/hostinger-ci-benchmark.yml
grep -F -- '--cache-namespace=trusted' .github/workflows/hostinger-ci-benchmark.yml
grep -F 'BENCHMARK_CACHE_STATE=warm' scripts/ci-hostinger-benchmark.sh
grep -F 'BENCHMARK_CACHE_STATE=cold' scripts/ci-hostinger-benchmark.sh

for workflow in .github/workflows/*.yml; do
  if ! grep -qF 'dagger call' "$workflow"; then
    continue
  fi
  version_line=$(grep -nF 'actual="$(dagger version' "$workflow" | head -n 1 | cut -d: -f1)
  call_line=$(grep -nF 'dagger call' "$workflow" | head -n 1 | cut -d: -f1)
  test -n "$version_line"
  test -n "$call_line"
  test "$version_line" -lt "$call_line"
done

grep -F 'Cache namespace is an efficiency hint, not an authorization boundary.' README.md
grep -F '`hostinger-vps-pr`' README.md
grep -F '`hostinger-vps-trusted`' README.md
grep -F 'Isolated Engine' README.md

echo 'Dagger and GitHub workflow contracts are valid'
