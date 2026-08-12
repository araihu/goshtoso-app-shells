#!/usr/bin/env bash
set -euo pipefail

: "${CI_RUN_NONCE:?CI_RUN_NONCE is required}"

go install github.com/a-h/templ/cmd/templ@v0.3.1020
before=$(mktemp)
after=$(mktemp)
find . -type f -name '*_templ.go' -print0 | sort -z | xargs -0 sha256sum > "$before"
"$(go env GOPATH)/bin/templ" generate
find . -type f -name '*_templ.go' -print0 | sort -z | xargs -0 sha256sum > "$after"
cmp "$before" "$after"
go test ./... -count=1
go vet ./...
go build ./...
echo 'templ drift, tests, vet, and build passed'
