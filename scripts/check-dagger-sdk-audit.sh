#!/usr/bin/env bash
set -euo pipefail

npm --prefix .dagger audit --package-lock-only --omit=dev --audit-level=high
