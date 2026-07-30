# Task 3 report — theme preference source

Status: complete.

RED: `GOWORK=off go test ./componentdocshell ./consoleshell -count=1` failed because both bootstrap scripts lacked the `default`/`preference` source marker and component docs lacked locked-theme persistence configuration.

GREEN: both bootstrap scripts set `data-theme-source` beside theme and dark state. A non-empty existing theme key marks `preference`; missing, empty, disabled, and storage-failure cases retain `default`. Deferred shell runtimes consume that marker instead of rereading theme storage. Component docs suppresses theme restore/write when its selector is disabled while retaining dark-mode persistence.

Focused verification: `GOWORK=off go test ./componentdocshell/... ./consoleshell/... -count=1` and `git diff --check` pass.

Nonblocking gap: no browser-engine execution matrix added for storage exceptions or Alpine watcher timing. Focused Go package tests inspect generated bootstrap/runtime bytes, as requested; broader browser coverage is deferred.
