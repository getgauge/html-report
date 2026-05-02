/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package mdgen

import (
	"path/filepath"
	"testing"
)

func TestStatusGlyph(t *testing.T) {
	tests := []struct {
		name string
		in   status
		want string
	}{
		{"pass", pass, "✅"},
		{"fail", fail, "❌"},
		{"skip", skip, "⏭️"},
		{"not executed", notExecuted, "▫️"},
		{"unknown", status("bogus"), "❓"},
		{"empty", status(""), "❓"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusGlyph(tt.in); got != tt.want {
				t.Errorf("statusGlyph(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestStatusWord(t *testing.T) {
	tests := []struct {
		name string
		in   status
		want string
	}{
		{"pass", pass, "Passed"},
		{"fail", fail, "Failed"},
		{"skip", skip, "Skipped"},
		{"not executed", notExecuted, "Not executed"},
		{"unknown", status("bogus"), "Unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := statusWord(tt.in); got != tt.want {
				t.Errorf("statusWord(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		in   int64
		want string
	}{
		{"zero", 0, "0ms"},
		{"sub-second", 250, "250ms"},
		{"just-under-1s", 999, "999ms"},
		{"exactly-1s", 1000, "1.00s"},
		{"a-few-seconds", 1234, "1.23s"},
		{"just-under-a-minute", 59_999, "60.00s"},
		{"one-minute", 60_000, "1m 0s"},
		{"minutes-and-seconds", 75_000, "1m 15s"},
		{"just-under-an-hour", 3_599_000, "59m 59s"},
		{"one-hour", 3_600_000, "1h 0m"},
		{"hours-and-minutes", 5_400_000, "1h 30m"},
		{"negative-clamps-to-zero", -500, "0ms"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDuration(tt.in); got != tt.want {
				t.Errorf("formatDuration(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestFormatPercent(t *testing.T) {
	tests := []struct {
		name string
		in   float32
		want string
	}{
		{"zero", 0, "0%"},
		{"hundred", 100, "100%"},
		{"one-decimal", 83.3, "83.3%"},
		{"rounds-up", 83.46, "83.5%"},
		{"trims-trailing-zero", 50.0, "50%"},
		{"small", 0.1, "0.1%"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatPercent(tt.in); got != tt.want {
				t.Errorf("formatPercent(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestEscapeMD(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "hello world", "hello world"},
		{"empty", "", ""},
		{"pipe", "a|b", "a\\|b"},
		{"asterisk", "*emphasis*", "\\*emphasis\\*"},
		{"underscore", "snake_case", "snake\\_case"},
		{"backtick", "use `code`", "use \\`code\\`"},
		{"square-brackets", "[link]", "\\[link\\]"},
		{"angle-brackets", "<tag>", "\\<tag\\>"},
		{"mixed", "foo*bar|baz", "foo\\*bar\\|baz"},
		{"unicode-untouched", "café ☕", "café ☕"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeMD(tt.in); got != tt.want {
				t.Errorf("escapeMD(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestEscapeMDIdempotent guards the most important property: applying
// escapeMD repeatedly must not keep adding backslashes. Without this the
// renderer could double-escape values that were already cleaned upstream.
func TestEscapeMDIdempotent(t *testing.T) {
	inputs := []string{
		"",
		"plain",
		"a|b",
		"*emphasis*",
		"foo*bar|baz_qux`run`<x>[y]",
		"already \\| escaped",
		"mix \\| of escaped \\* and unescaped |",
	}
	for _, in := range inputs {
		once := escapeMD(in)
		twice := escapeMD(once)
		if once != twice {
			t.Errorf("escapeMD not idempotent for %q: once=%q twice=%q", in, once, twice)
		}
	}
}

func TestMDLink(t *testing.T) {
	tests := []struct {
		name, text, href, want string
	}{
		{"plain", "Home", "index.md", "[Home](index.md)"},
		{"escapes-text", "a|b", "x.md", "[a\\|b](x.md)"},
		{"empty-text", "", "x.md", "[](x.md)"},
		{"href-with-spaces", "Spec", "specs/some file.md", "[Spec](specs/some file.md)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mdLink(tt.text, tt.href); got != tt.want {
				t.Errorf("mdLink(%q,%q) = %q, want %q", tt.text, tt.href, got, tt.want)
			}
		})
	}
}

func TestRelPath(t *testing.T) {
	tests := []struct {
		name, from, to, want string
	}{
		{"same-dir", filepath.Join("a", "b"), filepath.Join("a", "b", "c.md"), "c.md"},
		{"parent", filepath.Join("a", "b", "c"), filepath.Join("a", "d.md"), "../../d.md"},
		{"sibling", filepath.Join("a", "b"), filepath.Join("a", "c", "d.md"), "../c/d.md"},
		{"identity", "a", "a", "."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := relPath(tt.from, tt.to); got != tt.want {
				t.Errorf("relPath(%q,%q) = %q, want %q", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

// TestRelPathForwardSlashesOnWindowsStyle exercises the ToSlash conversion
// using a forward-slash input on every platform. We can't truly simulate
// Windows path separators on darwin/linux without faking filepath, so this
// just confirms we normalize the output rather than passing through.
func TestRelPathReturnsForwardSlashes(t *testing.T) {
	got := relPath(filepath.Join("a", "b"), filepath.Join("a", "b", "deep", "c.md"))
	want := "deep/c.md"
	if got != want {
		t.Errorf("relPath = %q, want %q", got, want)
	}
}
