# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project context

This is the Gauge **markdown-report** plugin (formerly `html-report`, currently mid-rewrite as of 2026-05). It's a long-lived gRPC reporter that Gauge invokes after a test suite finishes; the plugin transforms a `ProtoSuiteResult` into a directory of `.md` files. The HTML rewrite is tracked in **`PLAN.md`** — read it before starting work; its **Progress** section is the durable resume point across sessions and must be updated before and after every commit.

## Common commands

```sh
# Build the plugin binary into ./bin/<os>_<arch>/markdown-report
go run build/make.go

# Cross-compile for all supported platforms
go run build/make.go --all-platforms

# Install locally so `gauge` picks it up
go run build/make.go --install

# Build a release zip
go run build/make.go --distro

# Run the full test suite (CI gate)
go test ./...

# Race + coverage on the renderer package only
go test -race -cover ./mdgen/...

# Regenerate the mdgen golden fixtures after a deliberate format change
go test -update ./mdgen/...

# Lint (CI uses golangci-lint via .github/workflows/golangci-lint.yml)
go vet ./...
```

`go test` is the only test entry point — `make.go` is for building, not testing. CI (`.github/workflows/test.yml`) runs `go test ./...` on Linux, macOS, and Windows.

## Architecture

The runtime has two entry modes, both in `main.go`:

1. **gRPC server mode** (when `markdown-report_action=execution` env is set by Gauge). Starts a gRPC server, registers `handler.go` as a `ReporterServer`, and waits for Gauge to send a `SuiteExecutionResult`. On receipt, `mdReport.go::createReport` orchestrates: `mdgen.ToSuiteResult` (proto → domain) → `mdgen.GenerateReports` (writes the report tree).
2. **Regenerate-from-saved-proto mode** (when `--input` is passed on the CLI). `regenerate/regenerate.go::Report` reads `last_run_result` (proto-serialized `SuiteResult`), unmarshals, and calls the same `mdgen.GenerateReports`. This lets users rebuild a report without rerunning the suite.

The `mdgen/` package is the heart of the rewrite:

- `types.go` — domain model (`SuiteResult`, `spec`, `scenario`, `step`, `result`, `hookFailure`, `concept`, `table`). Lifted verbatim from the old `generator/` package; **JSON tags must stay stable** because `regenerate/_testdata/last_run_result.json` is checked in and round-trips through them.
- `transform.go` — `ToSuiteResult` walks a `ProtoSuiteResult` and populates the domain types. Sets the package-level `projectRoot` and appends to package-level `screenshotFiles`. These globals are deliberate carry-over from the HTML generator and the rendering-side code reads them. Tests must stamp them via `withProjectRoot(t, ...)` / `withScreenshotFiles(t, ...)` (see `integration_test.go`).
- `render.go` — one function per node type, each writing to an `io.Writer`. Output format is GFM; the contract is pinned by goldens in `mdgen/_testdata/`. `RenderIndex` and `RenderSpec` are the entry points; smaller helpers (`renderScenario`, `renderStep`, `renderResult`, `renderHookFailure`, `renderTable`, `renderSummary`) are package-private.
- `generate.go` — filesystem-side. `GenerateReports(res, reportsDir)` writes `index.md` at the root, fans out per-spec `.md` writes across goroutines, optionally emits per-directory `index.md` when `env.ShouldUseNestedSpecs()`, and copies screenshots into `images/`.
- `format.go` — pure helpers: status glyphs, `formatDuration`, `escapeMD` (idempotent regex-based escaping), `mdLink`, `relPath`.

### Test strategy in `mdgen/`

The package has four layers of tests, in increasing scope:

1. `format_test.go` / `transform_test.go` — pure-function unit tests.
2. `render_test.go` — per-render-function unit tests using `bytes.Buffer` + substring assertions. Don't pin exact output here; that's the goldens' job.
3. `golden_test.go` + `_testdata/*.md` — six representative `SuiteResult` shapes rendered end-to-end and diffed against checked-in expected files. The `-update` flag rewrites them; CI never passes `-update`. Goldens are review-able fixtures, not just snapshots.
4. `integration_test.go` + `parse_test.go` — call `GenerateReports` against a temp dir, then assert (a) the produced file tree, (b) every link in `index.md` resolves to a real file on disk, (c) every produced `.md` parses cleanly with goldmark + GFM and every link/image target resolves. The parse test is the highest-leverage single test: it catches "looks fine in goldens, breaks in real renderers" bugs.

Fixtures with side-effect dependencies (e.g. `with_screenshots` needs a backing PNG and a populated `screenshotFiles` slice) carry a `setup` hook on the `fixture` struct. Don't bypass it.

## Conventions specific to this repo

- **`make.go` constants matter.** `htmlReport`/`markdownReport` constants in `build/make.go` drive the binary name, the deploy zip name, and the install path. The same string also appears in `plugin.json`'s `id` field and `.github/workflows/deploy.yml`. Change all four together or release tooling breaks.
- **`plugin.json`'s `id` is the plugin identity.** `pluginActionEnv` in `mdReport.go` (`"markdown-report_action"`) must match `<id>_action` — this is the env var Gauge uses to dispatch actions to the binary.
- **Module path is `github.com/getgauge/html-report` despite the rewrite.** The Go module has not been renamed; only the plugin identity changed. New imports inside the repo use this path.
- **Don't reintroduce a global `htmlFiles`/`mdFiles` slice for parallel writes.** The old generator had a race here. Per-spec writes go straight to disk via `os.WriteFile` after rendering into a per-goroutine `bytes.Buffer`.
- **Screenshot href is hard-coded `../images/<basename>`** and assumes specs live one level deep under the report root. If `use_nested_specs` produces deeper trees, the link will be slightly wrong — flagged in PLAN.md, fix is not yet implemented.

## Guardrails the user expects

- **Update `PLAN.md`'s Progress section before and after every commit.** It's the canonical resume point. Folding the update into the same commit is fine.
- **Don't push without explicit confirmation.** Local commits on `master` (or feature branches) are the default; push only when asked.
- **Dependency adoption requires a 7-day quarantine.** Before bumping or adding any dep, verify the version's release date is ≥ 7 days old (e.g. `npm view <pkg>@<version> time`, `gh api repos/<owner>/<repo>/releases/tags/<tag> --jq '.published_at'`). Pin to specific patch versions, never floating majors.
- **`htmlReport_test.go` no longer exists** — its successor is `mdReport_test.go`. Don't grep for the old name expecting to find the entry-point tests.
