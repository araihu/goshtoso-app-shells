#!/usr/bin/env bash
set -euo pipefail

repo_root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd -P)"
fixture_source="$repo_root/testdata/external-consumer"
tmp_parent="$(CDPATH= cd -- "${TMPDIR:-/tmp}" && pwd -P)"

if [[ ! -f "$repo_root/go.mod" || ! -d "$repo_root/componentdocshell" ]]; then
	echo "repository root validation failed: $repo_root" >&2
	exit 1
fi
if [[ ! -f "$fixture_source/go.mod" || ! -f "$fixture_source/main_test.go" ]]; then
	echo "external consumer fixture is incomplete: $fixture_source" >&2
	exit 1
fi
if grep -n '/internal/' "$fixture_source/main_test.go" >/dev/null; then
	echo "external consumer fixture imports an internal package" >&2
	exit 1
fi

fixture_work="$(mktemp -d "$tmp_parent/componentdocshell-consumer.XXXXXX")"
cleanup() {
	case "$fixture_work" in
		"$tmp_parent"/componentdocshell-consumer.*)
			if [[ -d "$fixture_work" ]]; then
				rm -rf -- "$fixture_work"
			fi
			;;
		*)
			echo "refusing cleanup of unexpected path: $fixture_work" >&2
			return 1
			;;
	esac
}
trap cleanup EXIT

case "$fixture_work" in
	"$repo_root"|"$repo_root"/*)
		echo "temporary consumer target unexpectedly inside repository: $fixture_work" >&2
		exit 1
		;;
esac

cp -- "$fixture_source/go.mod" "$fixture_source/main_test.go" "$fixture_work/"
if [[ ! -d "$fixture_work" || ! -f "$fixture_work/go.mod" || ! -f "$fixture_work/main_test.go" ]]; then
	echo "temporary consumer target validation failed: $fixture_work" >&2
	exit 1
fi

cd -- "$fixture_work"
GOWORK=off go mod edit "-replace=github.com/araihu/goshtoso-app-shells=$repo_root"
GOWORK=off go mod tidy
GOWORK=off go mod verify
GOWORK=off go test ./... -count=1
