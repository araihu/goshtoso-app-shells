# Arai Hu fallback assets

Goshtoso App Shells keeps five embedded fallback files synchronized with
immutable [`araihu/assets`](https://github.com/araihu/assets) releases. Root
[`araihu-assets.json`](../araihu-assets.json) pins release tag, release commit,
archive URL and SHA-256, exact `release.json` SHA-256, and every allowed
source-to-destination mapping.

Brand mappings additionally pin catalog canonical name and semantic roles. The
updater rejects a release when a canonical name resolves to a different path,
role, or checksum. Theme CSS is a release-inventory file without brand catalog
identity.

## Local update

Download and verify immutable tar archive separately, then extract it into a
new directory. Go updater deliberately has no download mode:

```bash
go run ./cmd/araihu-assets-update -release-dir /path/to/verified-release
```

To advance manifest, provide complete identity as one unit:

```bash
go run ./cmd/araihu-assets-update \
  -release-dir /path/to/verified-release \
  -assets-repository araihu/assets \
  -assets-revision 0123456789abcdef0123456789abcdef01234567 \
  -release v1.2.3 \
  -release-url https://github.com/araihu/assets/releases/download/v1.2.3/araihu-assets-v1.2.3.tar.gz \
  -release-sha256 0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef \
  -release-json-sha256 abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789
```

Only stable `vX.Y.Z` tags and exact `araihu/assets` release URL shape are
accepted. Source and destination traversal, symlinks, duplicate destinations,
catalog collisions, and checksum mismatches fail before any copy. Run same
command twice; second run must report fallbacks current. All changed outputs
are staged before replacement. Mid-apply failure restores every
already-replaced file, and manifest is replaced last. Transaction errors report
failed path, paths applied before failure, and incomplete rollback operations.

Focused verification:

```bash
GOWORK=off go test ./internal/araihuassets ./cmd/araihu-assets-update -count=1
```

## Automation

`.github/workflows/araihu-assets.yml` accepts `repository_dispatch` event
`araihu-assets-released` and guarded manual dispatch. Both use exact fields:

- `assets_repository`
- `assets_revision`
- `release`
- `release_url`
- `release_sha256`
- `release_json_sha256`

Workflow validates all fields and resolves release tag to dispatched commit
before download, downloads and verifies archive once, runs offline updater
twice to prove idempotence, and opens or updates
`automation/araihu-assets-vX.Y.Z`. It uses selected-repository GitHub App
secrets `ARAIHU_ASSETS_APP_ID` and `ARAIHU_ASSETS_APP_PRIVATE_KEY`. Existing
`dependencies` and `assets` labels are applied when present. No label is
created, and no PR is auto-merged.

## Known hardening debt

Archive SHA-256 verification makes published archive immutable, and workflow
rejects traversal and link members before extraction. Future focused hardening
should also reject duplicate member names, case-folded member collisions, and
every non-regular archive member type before extraction. Offline updater still
enforces release inventory, catalog, checksum, path, symlink, and transactional
replacement checks.
