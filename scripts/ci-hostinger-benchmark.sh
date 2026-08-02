#!/usr/bin/env bash

set -euo pipefail

readonly marker="$(go env GOCACHE)/.araihu-goshtoso-app-shells-benchmark-v1"

if [[ -e "$marker" ]]; then
  echo "BENCHMARK_CACHE_STATE=warm"
else
  echo "BENCHMARK_CACHE_STATE=cold"
fi

go install github.com/a-h/templ/cmd/templ@v0.3.1020
"$(go env GOPATH)/bin/templ" generate
git diff --exit-code
GOWORK=off go test ./... -count=1
GOWORK=off go vet ./...
GOWORK=off go build ./...

mkdir -p "$(dirname "$marker")"
touch "$marker"
