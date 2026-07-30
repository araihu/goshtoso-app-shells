# Task 4 report — campaign integration hooks

Status: complete.

RED: focused component-docs and console layout contracts failed because neither
shell emitted managed asset hooks, the campaign toggle, or the dedicated
deferred runtime.

GREEN: both layouts now render managed fixed-size logo and optional favicon
hooks, a hidden accessible header toggle with localized-label data, and a
dedicated integrity-pinned deferred runtime after the first-paint bootstrap.
`Brand.HideName` suppresses only the displayed name; unset configuration keeps
the existing brand name and emits no campaign hooks. CSS only reserves managed
asset/toggle geometry, focus, and no-preference transitions.

Verification:

- `templ generate`
- `GOWORK=off go test ./componentdocshell ./consoleshell -count=1`
- `git diff --check`

Nonblocking debt: focused server-rendered contracts cover markup and ordering;
browser-engine coverage for runtime icon insertion and reduced-motion behavior
remains deferred.
