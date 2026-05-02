/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package mdgen

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// statusGlyph returns the emoji shown next to a status in tables and bullet
// lists. The glyphs were chosen to be unambiguous in monospace and accessible
// readers; statusWord is rendered alongside for the same reason.
func statusGlyph(s status) string {
	switch s {
	case pass:
		return "✅"
	case fail:
		return "❌"
	case skip:
		return "⏭️"
	case notExecuted:
		return "▫️"
	default:
		return "❓"
	}
}

// statusWord returns the human-readable status name.
func statusWord(s status) string {
	switch s {
	case pass:
		return "Passed"
	case fail:
		return "Failed"
	case skip:
		return "Skipped"
	case notExecuted:
		return "Not executed"
	default:
		return "Unknown"
	}
}

// formatDuration renders a millisecond count as a short human string. The
// HTML report used a HH:MM:SS clock format which read poorly for sub-second
// durations and overflowed past 24 hours; this version picks a unit based on
// magnitude.
func formatDuration(ms int64) string {
	if ms < 0 {
		ms = 0
	}
	switch {
	case ms < 1000:
		return fmt.Sprintf("%dms", ms)
	case ms < 60_000:
		return fmt.Sprintf("%.2fs", float64(ms)/1000)
	case ms < 3_600_000:
		m := ms / 60_000
		s := (ms % 60_000) / 1000
		return fmt.Sprintf("%dm %ds", m, s)
	default:
		h := ms / 3_600_000
		m := (ms % 3_600_000) / 60_000
		return fmt.Sprintf("%dh %dm", h, m)
	}
}

// formatPercent renders a percentage with at most one decimal, trimming a
// trailing ".0" so common values read as "100%" not "100.0%".
func formatPercent(f float32) string {
	s := fmt.Sprintf("%.1f", f)
	s = strings.TrimSuffix(s, ".0")
	return s + "%"
}

// mdSpecialRe matches the Markdown special characters we escape. The optional
// leading backslash lets the replacement skip already-escaped occurrences,
// which keeps escapeMD idempotent.
var mdSpecialRe = regexp.MustCompile("(\\\\?)([|`*_\\[\\]<>])")

// escapeMD escapes characters that would otherwise be interpreted as Markdown
// formatting when inlined into prose, table cells, or link text. It is
// idempotent: escapeMD(escapeMD(s)) == escapeMD(s).
//
// We deliberately do NOT escape backslashes themselves — escaping `\` would
// break idempotence and the special chars we care about (|, *, _, etc.) are
// the only ones that cause structural problems in our output.
func escapeMD(s string) string {
	return mdSpecialRe.ReplaceAllStringFunc(s, func(m string) string {
		if m[0] == '\\' {
			return m
		}
		return "\\" + m
	})
}

// mdLink renders a Markdown link with the text portion escaped for inline
// safety. Hrefs are passed through unchanged because callers produce them
// from filepath.Rel and they should not contain Markdown specials.
func mdLink(text, href string) string {
	return "[" + escapeMD(text) + "](" + href + ")"
}

// relPath returns to relative to from with forward slashes, suitable for use
// inside Markdown links regardless of host OS.
func relPath(from, to string) string {
	r, err := filepath.Rel(from, to)
	if err != nil {
		return filepath.ToSlash(to)
	}
	return filepath.ToSlash(r)
}
