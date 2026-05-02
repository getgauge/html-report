# Plan: Convert gauge HTML report plugin to Markdown reports

## Progress

> **Update this section before and after every commit.** It is the single durable record of where this multi-PR effort stands; a fresh session should be able to read it and resume work without re-deriving context.

**Status as of 2026-05-02:** PR 1 committed locally on `master` (not yet pushed). Ready to start PR 2.

| PR | Scope | Status | Commit |
| --- | --- | --- | --- |
| 1 | `mdgen/` package: types, transform, fragments, format helpers, tests for transform + format | 🟢 Done (local) | `d557725` |
| 2 | Renderers + golden + integration + parse tests (steps 3, 4.2–4.5) | ⏳ Not started | — |
| 3 | Switch entry point, delete HTML pipeline, rename plugin, update build + `deploy.yml` (step 5) | ⏳ Not started | — |
| 4 | README, schema, examples, migration note | ⏳ Not started | — |

**Uncommitted in working tree right now:** none (this PLAN.md edit will become a small follow-up commit, or fold into the first PR 2 commit).

**Next concrete action:** start PR 2 step 4.2 — write `mdgen/render_test.go` with one test per planned render function, using `bytes.Buffer` and substring assertions. The render functions don't exist yet; write tests against the API sketched in §3 of this plan, then implement the renderers to satisfy them (TDD). When the test file is in place but renderers are stubs, expect failing tests — that's the checkpoint to commit "tests in, renderers stubbed" before implementing.

**Push status:** `master` is 1 commit ahead of `origin/master`. User has not asked for a push yet — do not push without explicit confirmation, especially given this is a feature scaffold rather than a complete change.

**Status legend:** 🟢 Done · 🟡 In progress / ready to commit · ⏳ Not started · 🔴 Blocked

## 1. Goal

Replace the HTML rendering layer with a Markdown rendering layer while keeping the protobuf-to-domain transform, the gRPC handler, and the env/config plumbing. Output a folder of `.md` files (GitHub-Flavored Markdown) that renders well in GitHub, GitLab, IDE previewers, and standard viewers.

Out of scope: changing what data is captured, the gRPC contract with Gauge, or the regenerate-from-protobuf workflow.

## 2. Decisions to confirm before coding

These shape the rest of the plan; flag any you want to change before step 3.

1. **File layout** — mirror the current HTML structure: `index.md` at the root, one `.md` per spec under `specs/`. Alternative: a single combined `report.md`. **Recommendation: mirror the current structure** so the existing nesting/`use_nested_specs` logic carries over and large suites stay readable.
2. **Markdown flavor** — target **GitHub-Flavored Markdown** (tables, fenced code, task lists, autolinks). Do not depend on Mermaid by default; offer it as an opt-in for the summary chart.
3. **Status display** — use `✅ / ❌ / ⏭️` glyphs plus a text status word (`Passed` / `Failed` / `Skipped`). Configurable via env var if anyone objects to emoji.
4. **Screenshots** — copy files to `images/` (same as today) and reference them with `![alt](images/<file>.png)`. No base64 inlining (kills grep-ability and bloats files).
5. **Drop the following HTML-only features** — search index (`search.go` and the JS index), theme system (`theme/`, `themes/`), HTML minifier, sidebar partial, `htmlPageStartTag` / `htmlPageEndWithJS` chrome, modal/lightbox screenshot UI. Replace pie chart with a small results table (and optional Mermaid pie behind a flag).
6. **Plugin identity** — rename `id` in `plugin.json` from `html-report` to `markdown-report`; rename the env-var prefix `GAUGE_HTML_REPORT_*` to `GAUGE_MARKDOWN_REPORT_*`. Bump major version.

## 3. Implementation steps

### Step 1 — Carve out a new `mdgen/` package alongside `generator/`

- Create `mdgen/` (do **not** edit `generator/` until step 5; keeping both side-by-side lets us cross-check output during development).
- Copy `generator/transform.go`, `generator/fragments.go`, and the domain types from `generator/generate.go` (lines ~31–235) into `mdgen/types.go` and `mdgen/transform.go`. The domain model stays identical.
- In `mdgen/`, do **not** copy `search.go`, the minifier code, or theme-path resolution.

**Why a new package, not in-place edits:** lets us run the existing HTML golden tests during development as a regression net, then delete `generator/` in step 5 once `mdgen/` is at parity.

### Step 2 — Build the Markdown renderer

Replace Go `text/template` with direct rendering functions per node type. Templates were valuable for HTML because of nested structural tags; for Markdown, a small set of pure functions is clearer and easier to test.

Create `mdgen/render.go` with one function per domain type. Each takes the value and writes to an `io.Writer` (or returns a string for leaf renderers). Suggested signatures:

```go
func RenderIndex(w io.Writer, r *SuiteResult) error
func RenderSpec(w io.Writer, r *SuiteResult, s *spec) error
func renderSpecHeader(w io.Writer, s *spec) error
func renderScenario(w io.Writer, sc *scenario) error
func renderStep(w io.Writer, st *step) error
func renderResult(w io.Writer, res *result) error  // stack trace, error msg, screenshot
func renderHookFailure(w io.Writer, h *hookFailure, label string) error
func renderTable(w io.Writer, t *table) error
func renderSummary(w io.Writer, s *summary, scope string) error
```

Helpers in `mdgen/format.go` (already created in PR 1):

- `statusGlyph(status) string` — `✅` / `❌` / `⏭️`
- `statusWord(status) string`
- `formatDuration(ms int64) string` — `1m 23s` style
- `formatPercent(f float32) string`
- `escapeMD(s string) string` — escape pipes/backticks/asterisks/underscores in user-supplied text that gets inlined; **never** escape multiline step bodies that are themselves Markdown (preserve via fenced blocks where needed)
- `mdLink(text, href string) string`
- `relPath(from, to string) string` — replaces the `BasePath` template plumbing

Drop the `BasePath` propagation tree (`propogateBasePath` and friends). Compute relative paths at render time from spec file path → report root. This removes a whole class of template-context bugs.

### Step 3 — Define the output format (templates as code, not files)

Sketch the per-page markdown so reviewers can eyeball it before we wire it up. Examples below; align on these in PR review.

**`index.md`:**

```markdown
# Gauge Report — <ProjectName>

_Generated <Timestamp> · Environment: <Env> · Tags: <Tags>_

## Summary

| Scope | Total | ✅ Passed | ❌ Failed | ⏭️ Skipped | Success rate |
| --- | --- | --- | --- | --- | --- |
| Specs | 12 | 10 | 1 | 1 | 83.3% |
| Scenarios | 47 | 42 | 3 | 2 | 89.4% |

**Total time:** 2m 14s

<!-- Optional: ```mermaid ... ``` block when GAUGE_MARKDOWN_REPORT_MERMAID=true -->

## Pre-suite hook failures
<rendered via renderHookFailure if present>

## Specs

| Status | Spec | Time | Tags |
| --- | --- | --- | --- |
| ✅ | [Login flow](specs/login.md) | 1.2s | smoke |
| ❌ | [Checkout](specs/checkout.md) | 4.8s | regression |
```

**`specs/<name>.md`:**

```markdown
# <SpecHeading>

_File: `<FileName>` · Time: 4.8s · Tags: regression_

## Summary
<scenario summary table>

## Scenarios

### ❌ <Scenario heading> — 1.2s

#### Steps
- ✅ Given the user is on the login page _(120ms)_
- ❌ When the user enters invalid credentials _(80ms)_
  
  **Error:** `expected 200 got 401`
  
  <details><summary>Stack trace</summary>
  
  ```
  ...
  ```
  </details>
  
  ![Failure screenshot](../images/scn-3-failure.png)
```

Notes:

- Use `<details>` for stack traces and long error blocks — GFM renders these collapsed, keeps files scannable.
- Step substep concepts get rendered as nested bullets; flatten depth >3 into a fenced block to avoid Markdown indent ambiguity.
- For data-table-driven specs, render the data table once and per-row results in a sub-table.
- Multi-line step text and pre-formatted output go inside fenced ```` ``` ```` blocks to avoid being mangled by the surrounding Markdown.

### Step 4 — Wire up `GenerateReports`

In `mdgen/generate.go` write the new `GenerateReports(res *SuiteResult, reportsDir string) error`:

- No theme path argument (delete from signature).
- No `searchIndex` argument (drop the search feature).
- Create `reportsDir/index.md`, then iterate specs and write `reportsDir/<relpath>.md`. Reuse the `env.ShouldUseNestedSpecs()` decision and the `toHTMLFileName`-style relpath logic, renamed `toMDFileName` (already added in PR 1).
- Keep the parallel goroutine fan-out (it was useful; markdown is still I/O bound on large suites). Confirm via benchmarks in step 6.
- Call `copyScreenshotFiles` (lift unchanged from `generator/`).
- Drop `theme.CopyReportTemplateFiles` and `minifyHTMLFiles` calls entirely.

### Step 5 — Switch the entry point and delete the HTML pipeline

- Update `htmlReport.go` (rename to `mdReport.go`) to call `mdgen.GenerateReports` instead of `generator.GenerateReport`.
- Update `handler.go` to import the new package.
- Delete: `generator/` (entire dir), `theme/`, `themes/`, `htmlReport_test.go` (replaced by `mdReport_test.go`), all references to `bluemonday`, `blackfriday`, `tdewolff/minify`, `text/template` from `go.mod`. Run `go mod tidy`.
- Update `plugin.json`: id → `markdown-report`, version bump to next major (e.g. `5.0.0`).
- Update `env/env.go`: rename env-var prefix; remove `gauge_minify_reports` and the theme-path env var; keep `gauge_reports_dir`, `overwrite_reports`, `use_nested_specs`, `save_execution_result`, `gauge_screenshots_dir`.
- Update `regenerate/` to call the markdown generator.
- Update `build/make.go`: remove theme copy steps; rename release zip prefix.
- Update `.github/workflows/deploy.yml`:
  - line 33: `ls html-report*` → `ls markdown-report*`
  - line 40: release title text from "Gauge Html Report" to "Gauge Markdown Report"
  - line 49: `update_metadata.py html-report` → `update_metadata.py markdown-report`
- Update `README.md` and `schema.json` to match the new behavior.

(The `test.yml` and `golangci-lint.yml` workflows need no changes — they already glob the whole module via `go test ./...` and lint everything.)

### Step 6 — Run end-to-end against a real Gauge project

The test suite proves correctness; an actual run proves it ships. From a Gauge example project:

1. Build the plugin (`go run build/make.go`), install locally.
2. Run a known suite with passing, failing, and skipped specs, and at least one before/after hook failure.
3. Diff the output tree against the spec from step 3. Verify links work in GitHub preview (push a branch with the report committed) and in `glow` / VS Code preview.
4. Capture timings vs. the old HTML report on a large suite (>500 specs). Document in PR.

## 4. Testing strategy

The current suite is ~5.5k lines of generator tests — preserve that level of coverage. Markdown is easier to assert on than HTML (no whitespace/attribute-order noise), so prefer string assertions over diffing.

### 4.1 Unit tests — `mdgen/format_test.go` (DONE in PR 1)

Table-driven tests for every helper in `format.go`. These are the cheapest, fastest signal.

- `statusGlyph` / `statusWord` — every status enum value, including unknown.
- `formatDuration` — 0, sub-ms, seconds, minutes, hours; verify rounding.
- `formatPercent` — 0, 100, NaN-guard for 0/0 spec count.
- `escapeMD` — pipes, backticks, asterisks, underscores, backslashes, mixed; verify idempotence (escaping twice ≠ double-escape).
- `relPath` — same dir, parent dir, deeply nested, Windows-style paths (use `filepath.ToSlash`).
- `mdLink` — text containing `]`, `(`, etc.

### 4.2 Renderer unit tests — `mdgen/render_test.go`

One `Test<Func>` per render function. For each, build a small fixture struct, call the renderer, assert on the produced string. Patterns:

- **Happy path** — typical input, assert exact substring matches for headings, status glyphs, link targets.
- **Empty / nil** — empty Tags slice, nil hook failures, zero scenarios. Renderer must not panic; should produce sensible "no items" output where applicable.
- **Special characters** — spec heading with `|`, `*`, backticks; assert escaping kicks in.
- **Long content** — multi-line step text, large stack traces; assert fenced block wrapping and `<details>` collapse.
- **Status variants** — pass/fail/skip for `renderStep`, `renderScenario`, `renderResult`.

Use `bytes.Buffer` in tests, assert on `buf.String()`. Where exact match is brittle, assert with `strings.Contains` on the load-bearing pieces.

### 4.3 Golden-file tests — `mdgen/golden_test.go`

For end-to-end render shape. Mirror the existing `generator/_testdata/` approach.

- Create `mdgen/_testdata/` with hand-written `expected_*.md` fixtures for ~6 representative `SuiteResult` shapes:
  - all-pass minimal, all-pass with concepts, mixed pass/fail, all-skipped, before-suite-hook-failure (only `index.md` rendered), data-table-driven spec, spec with screenshots.
- Add a `-update` flag (`flag.Bool("update", false, ...)`) that rewrites the goldens on demand; CI runs without it.
- Assert on full file contents with a unified-diff helper for readable failures (`go-cmp` with line-by-line diff is fine; don't pull in a big dep just for this).
- Goldens are *checked-in fixtures*, not just snapshots — review changes deliberately in PRs.

### 4.4 Integration tests — `mdgen/integration_test.go`

End-to-end through `GenerateReports`:

- Build a `SuiteResult` with N specs in nested directories, call `GenerateReports(res, tmpDir)`, then assert the produced file tree:
  - `index.md` exists at root.
  - Each spec produces a `.md` file at the expected nested path.
  - Links in `index.md` resolve to the produced spec files (parse links out and `os.Stat` each target).
  - Screenshots referenced by render output exist under `images/`.
- Test `use_nested_specs=true` produces per-directory `index.md` files.
- Test re-running into the same dir with `overwrite_reports=true` vs `false`.

### 4.5 Round-trip parse test — `mdgen/parse_test.go`

Catches the entire class of "looks fine, breaks in real renderers" bugs.

- For each golden, parse the produced markdown with `github.com/yuin/goldmark` (GFM extensions enabled).
- Assert: no parse errors; all heading levels present; all links resolve to extant files; all image refs resolve; no stray HTML that GFM doesn't whitelist.

This is the highest-leverage single test we can add; HTML reports never had it because parsing HTML is harder.

### 4.6 Property tests — `mdgen/property_test.go`

Use `testing/quick` (no new dep) for two properties:

- `escapeMD` is idempotent on its own output for any input.
- For any `SuiteResult` randomly assembled from a small generator, `RenderIndex` produces output that parses successfully with goldmark.

Keep these short — property tests on Markdown rendering tend to flake on edge cases the user will never hit; pin the seed and cap input size.

### 4.7 Plugin-level test — `mdReport_test.go`

Lift the existing `htmlReport_test.go` cases (report dir creation, overwrite behavior, executable symlink) and update them for the new package and file extensions. These exercise the gRPC-handler-side wiring rather than rendering.

### 4.8 Coverage targets and CI

- `go test ./mdgen/... -cover`: aim for ≥85% line coverage (the renderers are pure functions; this is achievable). Don't chase 100% — error paths on `io.Writer.Write` failure aren't worth the test machinery.
- Run `go vet`, `staticcheck`, and `gofmt -d` in CI.
- Add a CI step that runs the integration test under `-race`.

## 5. Risks and mitigations

| Risk | Mitigation |
| --- | --- |
| Existing users have CI that scrapes `index.html` | Major version bump + clear README migration note. Don't try to ship both formats from one binary. |
| Large stack traces / data tables blow up file size | Use `<details>` collapse; document a future option to truncate. Don't pre-optimize. |
| GFM behaves differently in GitLab / Bitbucket / IDE previewers | Round-trip parse test (4.5) plus manual verification on GitHub + `glow` + VS Code (step 6). Don't try to support every viewer. |
| Spec heading containing markdown chars breaks tables in `index.md` | `escapeMD` covers this; the property test (4.6) is the safety net. |
| Parallel write fan-out reorders nothing visible but races on shared `screenshotFiles`/`htmlFiles` slices | Already a race in current code. Replace global slice append with a `sync.Mutex` or per-goroutine result channel during the rewrite. |

## 6. Suggested PR sequencing

1. **PR 1 (DONE)** — `mdgen/` package with types, transform, fragments, format helpers, and tests for transform + format. No wiring; existing HTML pipeline untouched. 79% coverage; passes under `-race`.
2. PR 2 — add renderers + golden + integration + parse tests (steps 3, 4.2–4.5). Still no wiring.
3. PR 3 — switch entry point, delete HTML pipeline, rename plugin, update build, update `deploy.yml` (step 5). This is the breaking change.
4. PR 4 — README, schema, examples, migration note.

Three smaller PRs are easier to review than one giant one, and PRs 1+2 can land with full test coverage before PR 3 makes the user-visible change.
