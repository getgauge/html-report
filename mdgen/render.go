/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

// Package mdgen render.go: turns the domain model into GitHub-Flavored
// Markdown. One function per node type, each writing to an io.Writer.
//
// The format is described in PLAN.md §3 (templates as code, not files).
// Renderers do not perform I/O beyond writing to the supplied writer; file
// layout and screenshot copying live in generate.go.
package mdgen

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

// fwrite is a thin Fprintf wrapper that swallows the byte count and surfaces
// only the error. Renderers chain many writes so dropping the count is purely
// for caller readability.
func fwrite(w io.Writer, format string, a ...any) error {
	_, err := fmt.Fprintf(w, format, a...)
	return err
}

// RenderIndex writes the top-level index.md for a SuiteResult.
func RenderIndex(w io.Writer, r *SuiteResult) error {
	if r == nil {
		return nil
	}
	if err := fwrite(w, "# Gauge Report — %s\n\n", escapeMD(r.ProjectName)); err != nil {
		return err
	}
	if err := fwrite(w, "_Generated %s · Environment: %s · Tags: %s_\n\n",
		escapeMD(r.Timestamp), escapeMD(r.Environment), escapeMD(r.Tags)); err != nil {
		return err
	}

	// Suite-level summary table covers Specs + Scenarios scopes.
	if err := fwrite(w, "## Summary\n\n"); err != nil {
		return err
	}
	if err := fwrite(w, "| Scope | Total | %s Passed | %s Failed | %s Skipped | Success rate |\n",
		statusGlyph(pass), statusGlyph(fail), statusGlyph(skip)); err != nil {
		return err
	}
	if err := fwrite(w, "| --- | --- | --- | --- | --- | --- |\n"); err != nil {
		return err
	}
	specTotal := r.PassedSpecsCount + r.FailedSpecsCount + r.SkippedSpecsCount
	scnTotal := r.PassedScenarioCount + r.FailedScenarioCount + r.SkippedScenarioCount
	specRate := percent(r.PassedSpecsCount, specTotal)
	scnRate := percent(r.PassedScenarioCount, scnTotal)
	if err := fwrite(w, "| Specs | %d | %d | %d | %d | %s |\n",
		specTotal, r.PassedSpecsCount, r.FailedSpecsCount, r.SkippedSpecsCount, formatPercent(specRate)); err != nil {
		return err
	}
	if err := fwrite(w, "| Scenarios | %d | %d | %d | %d | %s |\n\n",
		scnTotal, r.PassedScenarioCount, r.FailedScenarioCount, r.SkippedScenarioCount, formatPercent(scnRate)); err != nil {
		return err
	}
	if err := fwrite(w, "**Total time:** %s\n\n", formatDuration(r.ExecutionTime)); err != nil {
		return err
	}

	// Hook failures bracket the spec list because they affect the whole suite.
	if r.BeforeSuiteHookFailure != nil {
		if err := fwrite(w, "## Pre-suite hook failure\n\n"); err != nil {
			return err
		}
		if err := renderHookFailure(w, r.BeforeSuiteHookFailure, "Before Suite"); err != nil {
			return err
		}
	}
	if r.AfterSuiteHookFailure != nil {
		if err := fwrite(w, "## Post-suite hook failure\n\n"); err != nil {
			return err
		}
		if err := renderHookFailure(w, r.AfterSuiteHookFailure, "After Suite"); err != nil {
			return err
		}
	}

	if len(r.SpecResults) > 0 {
		if err := fwrite(w, "## Specs\n\n"); err != nil {
			return err
		}
		if err := fwrite(w, "| Status | Spec | Time | Tags |\n| --- | --- | --- | --- |\n"); err != nil {
			return err
		}
		for _, s := range r.SpecResults {
			href := indexLinkHref(s)
			tags := strings.Join(s.Tags, ", ")
			if err := fwrite(w, "| %s | %s | %s | %s |\n",
				statusGlyph(s.ExecutionStatus),
				mdLink(specDisplayName(s), href),
				formatDuration(s.ExecutionTime),
				escapeMD(tags),
			); err != nil {
				return err
			}
		}
		if err := fwrite(w, "\n"); err != nil {
			return err
		}
	}
	return nil
}

// indexLinkHref builds the relative href from index.md to the spec's .md
// page. Mirrors toMDFileName but anchored at the report root rather than a
// configurable basePath, since index.md lives there.
func indexLinkHref(s *spec) string {
	base := filepath.Base(s.FileName)
	dir := filepath.Dir(s.FileName)
	// Try to derive the relative dir under "specs/" by stripping anything up
	// to and including the project root. Without projectRoot we fall back to
	// just placing files under specs/<basename>.
	rel := ""
	if projectRoot != "" {
		if r, err := filepath.Rel(projectRoot, dir); err == nil && !strings.HasPrefix(r, "..") {
			rel = r
		}
	}
	if rel == "" || rel == "." {
		rel = "specs"
	}
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	return filepath.ToSlash(filepath.Join(rel, stem+".md"))
}

// specDisplayName picks a short, readable label for a spec link. SpecHeading
// is preferred; the basename is the fallback for unparseable specs.
func specDisplayName(s *spec) string {
	if strings.TrimSpace(s.SpecHeading) != "" {
		return s.SpecHeading
	}
	return filepath.Base(s.FileName)
}

// percent returns the pass rate as a float in 0–100. Guards against the empty
// suite case so 0/0 renders as 0% rather than NaN.
func percent(pass, total int) float32 {
	if total == 0 {
		return 0
	}
	return float32(pass) * 100.0 / float32(total)
}

// RenderSpec writes a single spec's markdown report.
func RenderSpec(w io.Writer, _ *SuiteResult, s *spec) error {
	if s == nil {
		return nil
	}
	if err := renderSpecHeader(w, s); err != nil {
		return err
	}

	// Parse errors short-circuit the rest: there are no scenarios to render.
	if len(s.Errors) > 0 {
		if err := fwrite(w, "## Errors\n\n"); err != nil {
			return err
		}
		for _, e := range s.Errors {
			if err := fwrite(w, "- %s (line %d): %s\n", errorTypeLabel(e.ErrorType), e.LineNumber, escapeMD(e.Message)); err != nil {
				return err
			}
		}
		if err := fwrite(w, "\n"); err != nil {
			return err
		}
		return nil
	}

	if err := fwrite(w, "## Summary\n\n"); err != nil {
		return err
	}
	if err := renderSummary(w, toScenarioSummary(s), "Scenarios"); err != nil {
		return err
	}
	if err := fwrite(w, "\n"); err != nil {
		return err
	}

	for _, hf := range s.BeforeSpecHookFailures {
		if err := renderHookFailure(w, hf, "Before Spec"); err != nil {
			return err
		}
	}

	if s.IsTableDriven && s.Datatable != nil {
		if err := fwrite(w, "## Data table\n\n"); err != nil {
			return err
		}
		if err := renderTable(w, s.Datatable); err != nil {
			return err
		}
		if err := fwrite(w, "\n"); err != nil {
			return err
		}
	}

	if len(s.Scenarios) > 0 {
		if err := fwrite(w, "## Scenarios\n\n"); err != nil {
			return err
		}
		for _, sc := range s.Scenarios {
			if err := renderScenario(w, sc); err != nil {
				return err
			}
		}
	}

	for _, hf := range s.AfterSpecHookFailures {
		if err := renderHookFailure(w, hf, "After Spec"); err != nil {
			return err
		}
	}
	return nil
}

func errorTypeLabel(t errorType) string {
	switch t {
	case parseErrorType:
		return "Parse Error"
	case validationErrorType:
		return "Validation Error"
	default:
		return "Error"
	}
}

// renderSpecHeader writes the H1 title and the metadata line beneath it.
func renderSpecHeader(w io.Writer, s *spec) error {
	if s == nil {
		return nil
	}
	heading := s.SpecHeading
	if strings.TrimSpace(heading) == "" {
		heading = filepath.Base(s.FileName)
	}
	if err := fwrite(w, "# %s\n\n", escapeMD(heading)); err != nil {
		return err
	}
	tags := strings.Join(s.Tags, ", ")
	parts := []string{}
	if s.FileName != "" {
		parts = append(parts, fmt.Sprintf("File: `%s`", filepath.Base(s.FileName)))
	}
	parts = append(parts, fmt.Sprintf("Time: %s", formatDuration(s.ExecutionTime)))
	if tags != "" {
		parts = append(parts, fmt.Sprintf("Tags: %s", escapeMD(tags)))
	}
	return fwrite(w, "_%s_\n\n", strings.Join(parts, " · "))
}

// renderScenario writes a single scenario block (heading, hooks, items).
func renderScenario(w io.Writer, sc *scenario) error {
	if sc == nil {
		return nil
	}
	if err := fwrite(w, "### %s %s — %s\n\n",
		statusGlyph(sc.ExecutionStatus), escapeMD(sc.Heading), sc.ExecutionTime); err != nil {
		return err
	}
	if len(sc.Tags) > 0 {
		if err := fwrite(w, "_Tags: %s_\n\n", escapeMD(strings.Join(sc.Tags, ", "))); err != nil {
			return err
		}
	}
	if sc.BeforeScenarioHookFailure != nil {
		if err := renderHookFailure(w, sc.BeforeScenarioHookFailure, "Before Scenario"); err != nil {
			return err
		}
	}

	if err := fwrite(w, "#### Steps\n\n"); err != nil {
		return err
	}
	for _, it := range sc.Contexts {
		if err := renderItem(w, it, 0); err != nil {
			return err
		}
	}
	for _, it := range sc.Items {
		if err := renderItem(w, it, 0); err != nil {
			return err
		}
	}
	for _, it := range sc.Teardowns {
		if err := renderItem(w, it, 0); err != nil {
			return err
		}
	}
	if err := fwrite(w, "\n"); err != nil {
		return err
	}

	if sc.AfterScenarioHookFailure != nil {
		if err := renderHookFailure(w, sc.AfterScenarioHookFailure, "After Scenario"); err != nil {
			return err
		}
	}
	return nil
}

// maxNestingDepth is the cutoff before substep concept items collapse into a
// fenced block — see PLAN.md §3 ("flatten depth >3 into a fenced block").
const maxNestingDepth = 3

// renderItem dispatches step / concept / comment items. depth tracks bullet
// nesting so concept substeps render as nested bullets.
func renderItem(w io.Writer, it item, depth int) error {
	switch it.Kind {
	case stepKind:
		if it.Step == nil {
			return nil
		}
		return renderStepBullet(w, it.Step, depth)
	case conceptKind:
		if it.Concept == nil {
			return nil
		}
		return renderConcept(w, it.Concept, depth)
	case commentKind:
		if it.Comment == nil {
			return nil
		}
		// Comments are scenario prose between steps; render as a plain
		// paragraph (italicised) so they don't get glued to the bullet
		// above them.
		return fwrite(w, "\n_%s_\n\n", escapeMD(strings.TrimSpace(it.Comment.Text)))
	}
	return nil
}

// renderConcept renders a concept item as a parent bullet plus nested
// children. Past maxNestingDepth, children flatten into a fenced block to
// avoid GFM indent ambiguity.
func renderConcept(w io.Writer, c *concept, depth int) error {
	if c.ConceptStep != nil {
		if err := renderStepBullet(w, c.ConceptStep, depth); err != nil {
			return err
		}
	}
	if depth >= maxNestingDepth {
		// Collapse remaining children into a fenced block so deeply nested
		// concepts stay readable.
		if err := fwrite(w, "%s```\n", indent(depth+1)); err != nil {
			return err
		}
		for _, sub := range c.Items {
			if err := writeFlatItem(w, sub); err != nil {
				return err
			}
		}
		return fwrite(w, "%s```\n", indent(depth+1))
	}
	for _, sub := range c.Items {
		if err := renderItem(w, sub, depth+1); err != nil {
			return err
		}
	}
	return nil
}

// writeFlatItem writes the textual representation of an item without
// markdown bullet structure — used inside the fenced-block fallback.
func writeFlatItem(w io.Writer, it item) error {
	switch it.Kind {
	case stepKind:
		if it.Step == nil {
			return nil
		}
		return fwrite(w, "%s\n", stepText(it.Step))
	case conceptKind:
		if it.Concept == nil || it.Concept.ConceptStep == nil {
			return nil
		}
		if err := fwrite(w, "%s\n", stepText(it.Concept.ConceptStep)); err != nil {
			return err
		}
		for _, sub := range it.Concept.Items {
			if err := writeFlatItem(w, sub); err != nil {
				return err
			}
		}
	case commentKind:
		if it.Comment != nil {
			return fwrite(w, "# %s\n", strings.TrimSpace(it.Comment.Text))
		}
	}
	return nil
}

// renderStepBullet writes a step as a markdown list item, then any error /
// stack / screenshot content beneath it indented appropriately.
func renderStepBullet(w io.Writer, st *step, depth int) error {
	pad := indent(depth)
	g := "▫️"
	if st.Result != nil {
		g = statusGlyph(st.Result.Status)
	}
	if err := fwrite(w, "%s- %s %s", pad, g, stepText(st)); err != nil {
		return err
	}
	if st.Result != nil && st.Result.ExecutionTime != "" {
		if err := fwrite(w, " _(%s)_", st.Result.ExecutionTime); err != nil {
			return err
		}
	}
	if err := fwrite(w, "\n"); err != nil {
		return err
	}

	// Multiline / table fragments live below the bullet so they're not
	// crammed into a single line.
	if err := renderStepBlockFragments(w, st, depth); err != nil {
		return err
	}

	if st.BeforeStepHookFailure != nil {
		if err := renderIndentedHookFailure(w, st.BeforeStepHookFailure, "Before Step", depth+1); err != nil {
			return err
		}
	}
	if st.AfterStepHookFailure != nil {
		if err := renderIndentedHookFailure(w, st.AfterStepHookFailure, "After Step", depth+1); err != nil {
			return err
		}
	}
	if st.Result != nil {
		if err := renderIndentedResult(w, st.Result, depth+1); err != nil {
			return err
		}
	}
	return nil
}

// indent returns 2*depth spaces, the GFM convention for nested list items.
func indent(depth int) string {
	if depth <= 0 {
		return ""
	}
	return strings.Repeat("  ", depth)
}

// stepText flattens a step's fragments into a single line of inline markdown
// (parameters become code-spans). Multiline + table fragments are placeholders
// here and rendered in a block beneath the bullet.
func stepText(st *step) string {
	if st == nil {
		return ""
	}
	if len(st.Fragments) == 0 {
		return escapeMD(st.StepText)
	}
	var b strings.Builder
	for _, f := range st.Fragments {
		switch f.FragmentKind {
		case textFragmentKind:
			b.WriteString(escapeMD(f.Text))
		case staticFragmentKind, dynamicFragmentKind, specialStringFragmentKind:
			b.WriteString("`")
			b.WriteString(strings.ReplaceAll(f.Text, "`", "'"))
			b.WriteString("`")
		case multilineFragmentKind:
			b.WriteString("(see block below)")
		case tableFragmentKind, specialTableFragmentKind:
			b.WriteString("(see table below)")
		}
	}
	return b.String()
}

// renderStepBlockFragments writes block-level fragments (multiline strings,
// tables, special tables) as separate elements below the bullet line.
func renderStepBlockFragments(w io.Writer, st *step, depth int) error {
	pad := indent(depth + 1)
	for _, f := range st.Fragments {
		switch f.FragmentKind {
		case multilineFragmentKind:
			if err := fwrite(w, "\n%s```\n", pad); err != nil {
				return err
			}
			for _, line := range strings.Split(f.Text, "\n") {
				if err := fwrite(w, "%s%s\n", pad, line); err != nil {
					return err
				}
			}
			if err := fwrite(w, "%s```\n\n", pad); err != nil {
				return err
			}
		case tableFragmentKind:
			if err := fwrite(w, "\n"); err != nil {
				return err
			}
			if err := renderTable(w, f.Table); err != nil {
				return err
			}
			if err := fwrite(w, "\n"); err != nil {
				return err
			}
		case specialTableFragmentKind:
			if err := fwrite(w, "\n%s```csv\n%s\n%s```\n\n", pad, f.Text, pad); err != nil {
				return err
			}
		}
	}
	return nil
}

// renderStep is exposed for tests; it wraps renderStepBullet at depth 0.
func renderStep(w io.Writer, st *step) error {
	if st == nil {
		return nil
	}
	return renderStepBullet(w, st, 0)
}

// renderResult writes the error / stack / screenshot block for a step result
// without any leading indent. Pass-status results render nothing.
func renderResult(w io.Writer, r *result) error {
	if r == nil {
		return nil
	}
	return renderIndentedResult(w, r, 0)
}

func renderIndentedResult(w io.Writer, r *result, depth int) error {
	if r == nil {
		return nil
	}
	pad := indent(depth)
	switch r.Status {
	case fail:
		if r.ErrorMessage != "" {
			if err := fwrite(w, "\n%s**Error:** `%s`\n", pad, oneLine(r.ErrorMessage)); err != nil {
				return err
			}
		}
		if r.StackTrace != "" {
			if err := writeDetails(w, "Stack trace", r.StackTrace, depth); err != nil {
				return err
			}
		}
		if r.FailureScreenshotFile != "" {
			if err := fwrite(w, "\n%s![Failure screenshot](%s)\n", pad, screenshotHref(r.FailureScreenshotFile)); err != nil {
				return err
			}
		}
	case skip:
		if r.SkippedReason != "" {
			if err := fwrite(w, "\n%s**Skipped:** %s\n", pad, escapeMD(r.SkippedReason)); err != nil {
				return err
			}
		}
	}
	return nil
}

// oneLine flattens whitespace so an error message renders cleanly in a
// backticked inline span.
func oneLine(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}

// screenshotHref renders the relative href to a screenshot. Renderers don't
// know the report root, so we anchor at images/<basename>; integration code
// copies files into images/ to match.
func screenshotHref(path string) string {
	return "../images/" + filepath.Base(path)
}

// writeDetails wraps text in a collapsible <details> block, the GFM idiom for
// keeping stack traces out of the way.
func writeDetails(w io.Writer, summary, body string, depth int) error {
	pad := indent(depth)
	if err := fwrite(w, "\n%s<details><summary>%s</summary>\n\n", pad, summary); err != nil {
		return err
	}
	if err := fwrite(w, "%s```\n", pad); err != nil {
		return err
	}
	for _, line := range strings.Split(body, "\n") {
		if err := fwrite(w, "%s%s\n", pad, line); err != nil {
			return err
		}
	}
	if err := fwrite(w, "%s```\n\n", pad); err != nil {
		return err
	}
	return fwrite(w, "%s</details>\n", pad)
}

// renderHookFailure writes a hook failure section header + body. Caller
// supplies the label so the section name matches the hook scope (Before
// Suite, Before Spec, etc.).
func renderHookFailure(w io.Writer, h *hookFailure, label string) error {
	if h == nil {
		return nil
	}
	return renderIndentedHookFailure(w, h, label, 0)
}

func renderIndentedHookFailure(w io.Writer, h *hookFailure, label string, depth int) error {
	if h == nil {
		return nil
	}
	pad := indent(depth)
	if err := fwrite(w, "%s**%s failed:** %s\n", pad, label, escapeMD(h.ErrMsg)); err != nil {
		return err
	}
	if h.StackTrace != "" {
		if err := writeDetails(w, "Stack trace", h.StackTrace, depth); err != nil {
			return err
		}
	}
	if h.FailureScreenshotFile != "" {
		if err := fwrite(w, "%s![Hook failure screenshot](%s)\n", pad, screenshotHref(h.FailureScreenshotFile)); err != nil {
			return err
		}
	}
	return fwrite(w, "\n")
}

// renderTable writes a GFM table. Cells with markdown specials are escaped so
// pipes don't break the column structure.
func renderTable(w io.Writer, t *table) error {
	if t == nil {
		return nil
	}
	if err := fwrite(w, "| %s |\n", strings.Join(escapeCells(t.Headers), " | ")); err != nil {
		return err
	}
	sep := make([]string, len(t.Headers))
	for i := range sep {
		sep[i] = "---"
	}
	if err := fwrite(w, "| %s |\n", strings.Join(sep, " | ")); err != nil {
		return err
	}
	for _, r := range t.Rows {
		if err := fwrite(w, "| %s |\n", strings.Join(escapeCells(r.Cells), " | ")); err != nil {
			return err
		}
	}
	return nil
}

func escapeCells(cells []string) []string {
	out := make([]string, len(cells))
	for i, c := range cells {
		out[i] = escapeMD(c)
	}
	return out
}

// renderSummary writes a one-row summary table for a single scope (e.g.
// "Specs" or "Scenarios"). Used inside per-spec pages.
func renderSummary(w io.Writer, s *summary, scope string) error {
	if s == nil {
		return nil
	}
	if err := fwrite(w, "| Scope | Total | %s Passed | %s Failed | %s Skipped | Success rate |\n",
		statusGlyph(pass), statusGlyph(fail), statusGlyph(skip)); err != nil {
		return err
	}
	if err := fwrite(w, "| --- | --- | --- | --- | --- | --- |\n"); err != nil {
		return err
	}
	rate := percent(s.Passed, s.Total)
	return fwrite(w, "| %s | %d | %d | %d | %d | %s |\n",
		escapeMD(scope), s.Total, s.Passed, s.Failed, s.Skipped, formatPercent(rate))
}
