/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package mdgen

import (
	"bytes"
	"strings"
	"testing"
)

// renderTo runs fn against a fresh buffer and returns the rendered string.
// Renderer-under-test errors fail the calling test.
func renderTo(t *testing.T, fn func(b *bytes.Buffer) error) string {
	t.Helper()
	var buf bytes.Buffer
	if err := fn(&buf); err != nil {
		t.Fatalf("renderer returned error: %v", err)
	}
	return buf.String()
}

func mustContain(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Errorf("output missing %q\n--- got ---\n%s", want, got)
	}
}

func mustNotContain(t *testing.T, got, unwanted string) {
	t.Helper()
	if strings.Contains(got, unwanted) {
		t.Errorf("output contains forbidden %q\n--- got ---\n%s", unwanted, got)
	}
}

func TestRenderSummary(t *testing.T) {
	t.Run("typical-counts", func(t *testing.T) {
		s := &summary{Total: 12, Passed: 10, Failed: 1, Skipped: 1}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderSummary(b, s, "Specs") })
		mustContain(t, got, "| Specs |")
		mustContain(t, got, "12")
		mustContain(t, got, "10")
		mustContain(t, got, "✅")
		mustContain(t, got, "❌")
		mustContain(t, got, "⏭️")
	})

	t.Run("zero-total-no-divide-by-zero", func(t *testing.T) {
		s := &summary{}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderSummary(b, s, "Specs") })
		mustContain(t, got, "0")
		mustNotContain(t, got, "NaN")
		mustNotContain(t, got, "+Inf")
	})
}

func TestRenderTable(t *testing.T) {
	t.Run("renders-headers-and-rows", func(t *testing.T) {
		tbl := &table{
			Headers: []string{"a", "b"},
			Rows: []*row{
				{Cells: []string{"1", "2"}, Result: pass},
				{Cells: []string{"3", "4"}, Result: fail},
			},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderTable(b, tbl) })
		mustContain(t, got, "| a | b |")
		mustContain(t, got, "| 1 | 2 |")
		mustContain(t, got, "| 3 | 4 |")
		mustContain(t, got, "| --- | --- |")
	})

	t.Run("escapes-pipe-in-cell", func(t *testing.T) {
		tbl := &table{
			Headers: []string{"col"},
			Rows:    []*row{{Cells: []string{"a|b"}, Result: pass}},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderTable(b, tbl) })
		mustContain(t, got, "a\\|b")
	})

	t.Run("nil-table-no-panic", func(t *testing.T) {
		got := renderTo(t, func(b *bytes.Buffer) error { return renderTable(b, nil) })
		if got != "" {
			t.Errorf("nil table should render empty, got %q", got)
		}
	})

	t.Run("empty-rows", func(t *testing.T) {
		tbl := &table{Headers: []string{"a", "b"}}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderTable(b, tbl) })
		mustContain(t, got, "| a | b |")
		mustContain(t, got, "| --- | --- |")
	})
}

func TestRenderHookFailure(t *testing.T) {
	hf := &hookFailure{
		HookName:   "Before Suite",
		ErrMsg:     "boom",
		StackTrace: "stack\nlines",
	}
	got := renderTo(t, func(b *bytes.Buffer) error { return renderHookFailure(b, hf, "Before Suite") })
	mustContain(t, got, "Before Suite")
	mustContain(t, got, "boom")
	mustContain(t, got, "<details>")
	mustContain(t, got, "stack")

	t.Run("nil-no-panic-no-output", func(t *testing.T) {
		got := renderTo(t, func(b *bytes.Buffer) error { return renderHookFailure(b, nil, "X") })
		if got != "" {
			t.Errorf("nil failure should render empty, got %q", got)
		}
	})

	t.Run("escapes-special-chars-in-msg", func(t *testing.T) {
		hf := &hookFailure{HookName: "Before Spec", ErrMsg: "got *star* and |pipe|"}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderHookFailure(b, hf, "Before Spec") })
		mustContain(t, got, "\\*star\\*")
		mustContain(t, got, "\\|pipe\\|")
	})

	t.Run("with-screenshot", func(t *testing.T) {
		hf := &hookFailure{HookName: "Before Spec", ErrMsg: "x", FailureScreenshotFile: "shot.png"}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderHookFailure(b, hf, "Before Spec") })
		mustContain(t, got, "shot.png")
		mustContain(t, got, "![")
	})
}

func TestRenderResult(t *testing.T) {
	t.Run("pass-no-error-output", func(t *testing.T) {
		r := &result{Status: pass}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderResult(b, r) })
		if strings.Contains(got, "Error:") || strings.Contains(got, "Stack trace") {
			t.Errorf("passing result should not render error block, got %q", got)
		}
	})

	t.Run("fail-renders-error-and-stack", func(t *testing.T) {
		r := &result{
			Status:       fail,
			ErrorMessage: "expected 200 got 401",
			StackTrace:   "at line 1\nat line 2",
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderResult(b, r) })
		mustContain(t, got, "Error:")
		mustContain(t, got, "expected 200 got 401")
		mustContain(t, got, "<details>")
		mustContain(t, got, "Stack trace")
		mustContain(t, got, "at line 1")
	})

	t.Run("skip-renders-reason", func(t *testing.T) {
		r := &result{Status: skip, SkippedReason: "requires-internet"}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderResult(b, r) })
		mustContain(t, got, "requires-internet")
	})

	t.Run("with-screenshot", func(t *testing.T) {
		r := &result{Status: fail, ErrorMessage: "x", FailureScreenshotFile: "scn-3-failure.png"}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderResult(b, r) })
		mustContain(t, got, "scn-3-failure.png")
		mustContain(t, got, "![")
	})

	t.Run("nil-no-panic", func(t *testing.T) {
		got := renderTo(t, func(b *bytes.Buffer) error { return renderResult(b, nil) })
		if got != "" {
			t.Errorf("nil result should render empty, got %q", got)
		}
	})
}

// fragment helpers reduce noise in step tests.
func textFrag(s string) *fragment   { return &fragment{FragmentKind: textFragmentKind, Text: s} }
func staticFrag(s string) *fragment { return &fragment{FragmentKind: staticFragmentKind, Text: s} }
func dynFrag(s string) *fragment    { return &fragment{FragmentKind: dynamicFragmentKind, Text: s} }

func TestRenderStep(t *testing.T) {
	t.Run("passing-step", func(t *testing.T) {
		st := &step{
			Fragments: []*fragment{textFrag("the user is on the login page")},
			Result:    &result{Status: pass, ExecutionTime: "00:00:00"},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderStep(b, st) })
		mustContain(t, got, "✅")
		mustContain(t, got, "the user is on the login page")
	})

	t.Run("failing-step-with-error", func(t *testing.T) {
		st := &step{
			Fragments: []*fragment{textFrag("when the user submits invalid creds")},
			Result: &result{
				Status:       fail,
				ErrorMessage: "expected 200 got 401",
				StackTrace:   "at line 1",
			},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderStep(b, st) })
		mustContain(t, got, "❌")
		mustContain(t, got, "when the user submits invalid creds")
		mustContain(t, got, "expected 200 got 401")
		mustContain(t, got, "<details>")
	})

	t.Run("skipped-step", func(t *testing.T) {
		st := &step{
			Fragments: []*fragment{textFrag("a step")},
			Result:    &result{Status: skip},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderStep(b, st) })
		mustContain(t, got, "⏭️")
	})

	t.Run("renders-static-param-as-code", func(t *testing.T) {
		st := &step{
			Fragments: []*fragment{
				textFrag("login as "),
				staticFrag("alice"),
			},
			Result: &result{Status: pass},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderStep(b, st) })
		mustContain(t, got, "login as")
		mustContain(t, got, "`alice`")
	})

	t.Run("renders-dynamic-param-as-code", func(t *testing.T) {
		st := &step{
			Fragments: []*fragment{
				textFrag("count is "),
				dynFrag("3"),
			},
			Result: &result{Status: pass},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderStep(b, st) })
		mustContain(t, got, "`3`")
	})

	t.Run("multiline-param-as-fenced-block", func(t *testing.T) {
		st := &step{
			Fragments: []*fragment{
				textFrag("with body "),
				{FragmentKind: multilineFragmentKind, Text: "line1\nline2\nline3"},
			},
			Result: &result{Status: pass},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderStep(b, st) })
		mustContain(t, got, "```")
		mustContain(t, got, "line1")
		mustContain(t, got, "line2")
	})

	t.Run("inline-table-param", func(t *testing.T) {
		st := &step{
			Fragments: []*fragment{
				textFrag("with rows "),
				{FragmentKind: tableFragmentKind, Table: &table{
					Headers: []string{"k", "v"},
					Rows:    []*row{{Cells: []string{"a", "1"}, Result: pass}},
				}},
			},
			Result: &result{Status: pass},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderStep(b, st) })
		mustContain(t, got, "| k | v |")
		mustContain(t, got, "| a | 1 |")
	})

	t.Run("nil-result-no-panic", func(t *testing.T) {
		st := &step{Fragments: []*fragment{textFrag("a step")}}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderStep(b, st) })
		mustContain(t, got, "a step")
	})

	t.Run("before-step-hook-failure-rendered", func(t *testing.T) {
		st := &step{
			Fragments:             []*fragment{textFrag("the step")},
			BeforeStepHookFailure: &hookFailure{HookName: "Before Step", ErrMsg: "setup blew up"},
			Result:                &result{Status: fail},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderStep(b, st) })
		mustContain(t, got, "Before Step")
		mustContain(t, got, "setup blew up")
	})
}

func TestRenderScenario(t *testing.T) {
	t.Run("passing-scenario", func(t *testing.T) {
		sc := &scenario{
			Heading:         "User logs in",
			ExecutionStatus: pass,
			ExecutionTime:   "00:00:01",
			Tags:            []string{"smoke"},
			Items: []item{
				{Kind: stepKind, Step: &step{
					Fragments: []*fragment{textFrag("login")},
					Result:    &result{Status: pass},
				}},
			},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderScenario(b, sc) })
		mustContain(t, got, "✅")
		mustContain(t, got, "User logs in")
		mustContain(t, got, "smoke")
		mustContain(t, got, "login")
	})

	t.Run("failing-scenario", func(t *testing.T) {
		sc := &scenario{
			Heading:         "Bad path",
			ExecutionStatus: fail,
			ExecutionTime:   "00:00:00",
			Items: []item{
				{Kind: stepKind, Step: &step{
					Fragments: []*fragment{textFrag("a step")},
					Result:    &result{Status: fail, ErrorMessage: "bad"},
				}},
			},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderScenario(b, sc) })
		mustContain(t, got, "❌")
		mustContain(t, got, "Bad path")
		mustContain(t, got, "bad")
	})

	t.Run("scenario-with-special-chars-in-heading", func(t *testing.T) {
		sc := &scenario{
			Heading:         "User|admin enters *creds*",
			ExecutionStatus: pass,
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderScenario(b, sc) })
		mustContain(t, got, "User\\|admin enters \\*creds\\*")
	})

	t.Run("with-before-and-after-hook-failures", func(t *testing.T) {
		sc := &scenario{
			Heading:                   "x",
			ExecutionStatus:           fail,
			BeforeScenarioHookFailure: &hookFailure{HookName: "Before Scenario", ErrMsg: "pre-hook err"},
			AfterScenarioHookFailure:  &hookFailure{HookName: "After Scenario", ErrMsg: "post-hook err"},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderScenario(b, sc) })
		mustContain(t, got, "pre-hook err")
		mustContain(t, got, "post-hook err")
	})

	t.Run("with-context-and-teardown-steps", func(t *testing.T) {
		sc := &scenario{
			Heading:         "x",
			ExecutionStatus: pass,
			Contexts: []item{{Kind: stepKind, Step: &step{
				Fragments: []*fragment{textFrag("setup")}, Result: &result{Status: pass},
			}}},
			Items: []item{{Kind: stepKind, Step: &step{
				Fragments: []*fragment{textFrag("the step")}, Result: &result{Status: pass},
			}}},
			Teardowns: []item{{Kind: stepKind, Step: &step{
				Fragments: []*fragment{textFrag("teardown")}, Result: &result{Status: pass},
			}}},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderScenario(b, sc) })
		mustContain(t, got, "setup")
		mustContain(t, got, "the step")
		mustContain(t, got, "teardown")
	})

	t.Run("with-comment-item", func(t *testing.T) {
		sc := &scenario{
			Heading:         "x",
			ExecutionStatus: pass,
			Items: []item{
				{Kind: commentKind, Comment: &comment{Text: "this is a remark"}},
				{Kind: stepKind, Step: &step{
					Fragments: []*fragment{textFrag("a step")}, Result: &result{Status: pass},
				}},
			},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderScenario(b, sc) })
		mustContain(t, got, "this is a remark")
	})

	t.Run("with-concept-item-shows-nesting", func(t *testing.T) {
		conceptStep := &step{
			Fragments: []*fragment{textFrag("perform login flow")},
			Result:    &result{Status: pass},
		}
		conceptItem := &concept{
			ConceptStep: conceptStep,
			Items: []item{
				{Kind: stepKind, Step: &step{
					Fragments: []*fragment{textFrag("inner step")}, Result: &result{Status: pass},
				}},
			},
		}
		sc := &scenario{
			Heading:         "x",
			ExecutionStatus: pass,
			Items:           []item{{Kind: conceptKind, Concept: conceptItem}},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderScenario(b, sc) })
		mustContain(t, got, "perform login flow")
		mustContain(t, got, "inner step")
	})
}

func TestRenderSpecHeader(t *testing.T) {
	s := &spec{
		SpecHeading:   "Login flow",
		FileName:      "/proj/specs/login.spec",
		Tags:          []string{"smoke", "auth"},
		ExecutionTime: 4800,
	}
	got := renderTo(t, func(b *bytes.Buffer) error { return renderSpecHeader(b, s) })
	mustContain(t, got, "# Login flow")
	mustContain(t, got, "login.spec")
	mustContain(t, got, "smoke")
	mustContain(t, got, "auth")
	mustContain(t, got, "4.80s")

	t.Run("special-chars-in-heading-escaped", func(t *testing.T) {
		s := &spec{SpecHeading: "Foo|Bar*Baz"}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderSpecHeader(b, s) })
		mustContain(t, got, "Foo\\|Bar\\*Baz")
	})

	t.Run("no-tags-renders-cleanly", func(t *testing.T) {
		s := &spec{SpecHeading: "x", FileName: "x.spec"}
		got := renderTo(t, func(b *bytes.Buffer) error { return renderSpecHeader(b, s) })
		mustContain(t, got, "# x")
	})
}

func TestRenderSpec(t *testing.T) {
	suite := &SuiteResult{
		ProjectName: "demo",
		Timestamp:   "Jan 2, 2026 at 3:04pm",
		Environment: "default",
	}
	s := &spec{
		SpecHeading:         "Login flow",
		FileName:            "/proj/specs/login.spec",
		ExecutionTime:       4800,
		ExecutionStatus:     fail,
		PassedScenarioCount: 1,
		FailedScenarioCount: 1,
		Scenarios: []*scenario{
			{Heading: "Pass case", ExecutionStatus: pass, Items: []item{{Kind: stepKind, Step: &step{
				Fragments: []*fragment{textFrag("login")}, Result: &result{Status: pass},
			}}}},
			{Heading: "Fail case", ExecutionStatus: fail, Items: []item{{Kind: stepKind, Step: &step{
				Fragments: []*fragment{textFrag("login")}, Result: &result{Status: fail, ErrorMessage: "bad"},
			}}}},
		},
	}
	got := renderTo(t, func(b *bytes.Buffer) error { return RenderSpec(b, suite, s) })
	mustContain(t, got, "# Login flow")
	mustContain(t, got, "## Summary")
	mustContain(t, got, "## Scenarios")
	mustContain(t, got, "Pass case")
	mustContain(t, got, "Fail case")
	mustContain(t, got, "bad")

	t.Run("with-spec-hook-failures", func(t *testing.T) {
		s := &spec{
			SpecHeading: "x", FileName: "x.spec",
			BeforeSpecHookFailures: []*hookFailure{{HookName: "Before Spec", ErrMsg: "pre"}},
			AfterSpecHookFailures:  []*hookFailure{{HookName: "After Spec", ErrMsg: "post"}},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return RenderSpec(b, suite, s) })
		mustContain(t, got, "pre")
		mustContain(t, got, "post")
	})

	t.Run("with-data-table", func(t *testing.T) {
		s := &spec{
			SpecHeading: "x", FileName: "x.spec", IsTableDriven: true,
			Datatable: &table{
				Headers: []string{"name", "age"},
				Rows: []*row{
					{Cells: []string{"alice", "30"}, Result: pass},
					{Cells: []string{"bob", "25"}, Result: fail},
				},
			},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return RenderSpec(b, suite, s) })
		mustContain(t, got, "| name | age |")
		mustContain(t, got, "alice")
		mustContain(t, got, "bob")
	})

	t.Run("parse-errors-rendered", func(t *testing.T) {
		s := &spec{
			SpecHeading: "x", FileName: "x.spec",
			Errors: []buildError{
				{ErrorType: parseErrorType, FileName: "x.spec", LineNumber: 3, Message: "unexpected token"},
			},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return RenderSpec(b, suite, s) })
		mustContain(t, got, "Parse Error")
		mustContain(t, got, "unexpected token")
	})
}

func TestRenderIndex(t *testing.T) {
	suite := &SuiteResult{
		ProjectName:          "demo",
		Timestamp:            "Jan 2, 2026 at 3:04pm",
		Environment:          "default",
		Tags:                 "smoke",
		ExecutionTime:        134_000,
		SuccessRate:          83.3,
		PassedSpecsCount:     10,
		FailedSpecsCount:     1,
		SkippedSpecsCount:    1,
		PassedScenarioCount:  42,
		FailedScenarioCount:  3,
		SkippedScenarioCount: 2,
		SpecResults: []*spec{
			{SpecHeading: "Login flow", FileName: "/proj/specs/login.spec", ExecutionTime: 1200, ExecutionStatus: pass, Tags: []string{"smoke"}},
			{SpecHeading: "Checkout", FileName: "/proj/specs/checkout.spec", ExecutionTime: 4800, ExecutionStatus: fail, Tags: []string{"regression"}},
		},
	}
	got := renderTo(t, func(b *bytes.Buffer) error { return RenderIndex(b, suite) })
	mustContain(t, got, "# Gauge Report")
	mustContain(t, got, "demo")
	mustContain(t, got, "## Summary")
	mustContain(t, got, "## Specs")
	mustContain(t, got, "Login flow")
	mustContain(t, got, "Checkout")
	mustContain(t, got, "specs/login.md")
	mustContain(t, got, "specs/checkout.md")
	mustContain(t, got, "2m 14s")

	t.Run("renders-pre-suite-hook-failure", func(t *testing.T) {
		suite := &SuiteResult{
			ProjectName:            "demo",
			BeforeSuiteHookFailure: &hookFailure{HookName: "Before Suite", ErrMsg: "init failed"},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return RenderIndex(b, suite) })
		mustContain(t, got, "init failed")
		mustContain(t, got, "Before Suite")
	})

	t.Run("renders-post-suite-hook-failure", func(t *testing.T) {
		suite := &SuiteResult{
			ProjectName:           "demo",
			AfterSuiteHookFailure: &hookFailure{HookName: "After Suite", ErrMsg: "cleanup failed"},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return RenderIndex(b, suite) })
		mustContain(t, got, "cleanup failed")
	})

	t.Run("empty-spec-list", func(t *testing.T) {
		suite := &SuiteResult{ProjectName: "demo"}
		got := renderTo(t, func(b *bytes.Buffer) error { return RenderIndex(b, suite) })
		mustContain(t, got, "# Gauge Report")
		// no panic, no `## Specs` content rows is fine
	})

	t.Run("link-uses-md-extension", func(t *testing.T) {
		// projectRoot is set by ToSuiteResult in production; tests must
		// stamp it so relative spec paths resolve.
		saved := projectRoot
		projectRoot = "/proj"
		t.Cleanup(func() { projectRoot = saved })

		suite := &SuiteResult{
			ProjectName: "demo",
			SpecResults: []*spec{
				{SpecHeading: "x", FileName: "/proj/specs/sub/x.spec", ExecutionStatus: pass},
			},
		}
		got := renderTo(t, func(b *bytes.Buffer) error { return RenderIndex(b, suite) })
		mustContain(t, got, "specs/sub/x.md")
		mustNotContain(t, got, ".spec)")
	})
}
