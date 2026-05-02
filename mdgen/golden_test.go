/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package mdgen

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// updateGolden, when -update is passed, rewrites every golden file from the
// renderer's current output. CI never runs with this flag; reviewers diff the
// updated files in the PR.
var updateGolden = flag.Bool("update", false, "regenerate golden files in mdgen/_testdata")

// fixture pairs a name with a builder function. setup, when present, runs
// before GenerateReports — used by fixtures that depend on side-effect state
// (screenshotFiles, environment variables) that production transform.go
// would normally populate.
type fixture struct {
	name    string
	specDir string // relative dir under projectRoot for the per-spec golden
	build   func() *SuiteResult
	setup   func(t *testing.T)
}

// fixtureProjectRoot is the synthetic project root all fixtures use, so the
// renderer's link-href logic produces deterministic paths.
const fixtureProjectRoot = "/proj"

// allFixtures enumerates the suite shapes covered by golden tests. The set
// matches PLAN.md §4.3: minimal pass, mixed pass/fail, all-skipped,
// before-suite-hook-failure, data-table-driven, with-screenshots.
func allFixtures() []fixture {
	return []fixture{
		{name: "all_pass_minimal", specDir: "specs/login", build: buildAllPassMinimal},
		{name: "mixed_pass_fail", specDir: "specs/checkout", build: buildMixedPassFail},
		{name: "all_skipped", specDir: "specs/skipped", build: buildAllSkipped},
		{name: "before_suite_hook_failure", build: buildBeforeSuiteHookFailure},
		{name: "data_table_driven", specDir: "specs/users", build: buildDataTableDriven},
		{name: "with_screenshots", specDir: "specs/checkout", build: buildWithScreenshots, setup: setupWithScreenshots},
	}
}

func buildAllPassMinimal() *SuiteResult {
	return &SuiteResult{
		ProjectName:         "demo",
		Timestamp:           "Jan 2, 2026 at 3:04pm",
		Environment:         "default",
		Tags:                "smoke",
		ExecutionTime:       1500,
		SuccessRate:         100,
		PassedSpecsCount:    1,
		PassedScenarioCount: 1,
		ExecutionStatus:     pass,
		SpecResults: []*spec{
			{
				SpecHeading:         "Login flow",
				FileName:            "/proj/specs/login.spec",
				Tags:                []string{"smoke"},
				ExecutionTime:       1500,
				ExecutionStatus:     pass,
				PassedScenarioCount: 1,
				Scenarios: []*scenario{
					{
						Heading:         "User logs in",
						ExecutionStatus: pass,
						ExecutionTime:   "00:00:01",
						Items: []item{
							{Kind: stepKind, Step: &step{
								Fragments: []*fragment{textFrag("the user is on the login page")},
								Result:    &result{Status: pass, ExecutionTime: "00:00:00"},
							}},
							{Kind: stepKind, Step: &step{
								Fragments: []*fragment{
									textFrag("the user logs in as "),
									staticFrag("alice"),
								},
								Result: &result{Status: pass, ExecutionTime: "00:00:01"},
							}},
						},
					},
				},
			},
		},
	}
}

func buildMixedPassFail() *SuiteResult {
	return &SuiteResult{
		ProjectName:         "demo",
		Timestamp:           "Jan 2, 2026 at 3:04pm",
		Environment:         "default",
		Tags:                "regression",
		ExecutionTime:       6000,
		SuccessRate:         50,
		PassedSpecsCount:    0,
		FailedSpecsCount:    1,
		PassedScenarioCount: 1,
		FailedScenarioCount: 1,
		ExecutionStatus:     fail,
		SpecResults: []*spec{
			{
				SpecHeading:         "Checkout",
				FileName:            "/proj/specs/checkout.spec",
				Tags:                []string{"regression"},
				ExecutionTime:       6000,
				ExecutionStatus:     fail,
				PassedScenarioCount: 1,
				FailedScenarioCount: 1,
				Scenarios: []*scenario{
					{
						Heading:         "Happy path",
						ExecutionStatus: pass,
						ExecutionTime:   "00:00:01",
						Items: []item{{Kind: stepKind, Step: &step{
							Fragments: []*fragment{textFrag("checkout completes")},
							Result:    &result{Status: pass, ExecutionTime: "00:00:01"},
						}}},
					},
					{
						Heading:         "Bad path",
						ExecutionStatus: fail,
						ExecutionTime:   "00:00:05",
						Items: []item{{Kind: stepKind, Step: &step{
							Fragments: []*fragment{textFrag("checkout breaks")},
							Result: &result{
								Status:        fail,
								ErrorMessage:  "expected 200 got 500",
								StackTrace:    "at handler.go:42\nat router.go:11",
								ExecutionTime: "00:00:00",
							},
						}}},
					},
				},
			},
		},
	}
}

func buildAllSkipped() *SuiteResult {
	return &SuiteResult{
		ProjectName:          "demo",
		Timestamp:            "Jan 2, 2026 at 3:04pm",
		Environment:          "default",
		Tags:                 "wip",
		SuccessRate:          0,
		SkippedSpecsCount:    1,
		SkippedScenarioCount: 1,
		ExecutionStatus:      skip,
		SpecResults: []*spec{
			{
				SpecHeading:          "Skipped flows",
				FileName:             "/proj/specs/skipped.spec",
				Tags:                 []string{"wip"},
				ExecutionStatus:      skip,
				SkippedScenarioCount: 1,
				Scenarios: []*scenario{
					{
						Heading:         "Some scenario",
						ExecutionStatus: skip,
						Items: []item{{Kind: stepKind, Step: &step{
							Fragments: []*fragment{textFrag("a step")},
							Result:    &result{Status: skip, SkippedReason: "blocked-by-INFRA-12"},
						}}},
					},
				},
			},
		},
	}
}

func buildBeforeSuiteHookFailure() *SuiteResult {
	return &SuiteResult{
		ProjectName:     "demo",
		Timestamp:       "Jan 2, 2026 at 3:04pm",
		Environment:     "default",
		ExecutionStatus: fail,
		BeforeSuiteHookFailure: &hookFailure{
			HookName:   "Before Suite",
			ErrMsg:     "could not connect to db",
			StackTrace: "at db.go:12\nat suite.go:3",
		},
	}
}

func buildDataTableDriven() *SuiteResult {
	dt := &table{
		Headers: []string{"name", "age"},
		Rows: []*row{
			{Cells: []string{"alice", "30"}, Result: pass},
			{Cells: []string{"bob", "25"}, Result: fail},
		},
	}
	return &SuiteResult{
		ProjectName:         "demo",
		Timestamp:           "Jan 2, 2026 at 3:04pm",
		Environment:         "default",
		ExecutionTime:       2000,
		FailedSpecsCount:    1,
		PassedScenarioCount: 1,
		FailedScenarioCount: 1,
		ExecutionStatus:     fail,
		SpecResults: []*spec{
			{
				SpecHeading:         "Users table",
				FileName:            "/proj/specs/users.spec",
				ExecutionTime:       2000,
				ExecutionStatus:     fail,
				IsTableDriven:       true,
				Datatable:           dt,
				PassedScenarioCount: 1,
				FailedScenarioCount: 1,
				Scenarios: []*scenario{
					{
						Heading:         "Row passes",
						ExecutionStatus: pass,
						ExecutionTime:   "00:00:01",
						TableRowIndex:   0,
						Items: []item{{Kind: stepKind, Step: &step{
							Fragments: []*fragment{textFrag("a user is created")},
							Result:    &result{Status: pass, ExecutionTime: "00:00:01"},
						}}},
					},
					{
						Heading:         "Row fails",
						ExecutionStatus: fail,
						ExecutionTime:   "00:00:00",
						TableRowIndex:   1,
						Items: []item{{Kind: stepKind, Step: &step{
							Fragments: []*fragment{textFrag("a user is created")},
							Result: &result{
								Status:       fail,
								ErrorMessage: "validation failed",
							},
						}}},
					},
				},
			},
		},
	}
}

// setupWithScreenshots seeds the package-level screenshot slice and writes
// a backing file for the screenshot the fixture references. In production
// transform.go does both jobs as it walks the proto tree.
func setupWithScreenshots(t *testing.T) {
	t.Helper()
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "scn-fail.png"), []byte("PNG"), 0o644); err != nil {
		t.Fatalf("seed screenshot: %v", err)
	}
	t.Setenv("gauge_screenshots_dir", srcDir)
	saved := screenshotFiles
	screenshotFiles = []string{"scn-fail.png"}
	t.Cleanup(func() { screenshotFiles = saved })
}

func buildWithScreenshots() *SuiteResult {
	return &SuiteResult{
		ProjectName:         "demo",
		Timestamp:           "Jan 2, 2026 at 3:04pm",
		Environment:         "default",
		ExecutionTime:       3000,
		FailedSpecsCount:    1,
		FailedScenarioCount: 1,
		ExecutionStatus:     fail,
		SpecResults: []*spec{
			{
				SpecHeading:         "Checkout",
				FileName:            "/proj/specs/checkout.spec",
				ExecutionTime:       3000,
				ExecutionStatus:     fail,
				FailedScenarioCount: 1,
				Scenarios: []*scenario{
					{
						Heading:         "Bad path",
						ExecutionStatus: fail,
						ExecutionTime:   "00:00:03",
						Items: []item{{Kind: stepKind, Step: &step{
							Fragments: []*fragment{textFrag("checkout breaks")},
							Result: &result{
								Status:                fail,
								ErrorMessage:          "boom",
								FailureScreenshotFile: "scn-fail.png",
								ExecutionTime:         "00:00:00",
							},
						}}},
					},
				},
			},
		},
	}
}

// TestGoldenIndex compares each fixture's RenderIndex output against a
// committed _testdata/<name>_index.md file. -update rewrites them.
func TestGoldenIndex(t *testing.T) {
	saved := projectRoot
	projectRoot = fixtureProjectRoot
	t.Cleanup(func() { projectRoot = saved })

	for _, f := range allFixtures() {
		t.Run(f.name, func(t *testing.T) {
			suite := f.build()
			var buf bytes.Buffer
			if err := RenderIndex(&buf, suite); err != nil {
				t.Fatalf("RenderIndex: %v", err)
			}
			compareGolden(t, "_testdata/"+f.name+"_index.md", buf.Bytes())
		})
	}
}

// TestGoldenSpec compares each fixture's first spec rendered via RenderSpec
// against the committed golden. Fixtures with no spec body (hook-only) are
// skipped.
func TestGoldenSpec(t *testing.T) {
	saved := projectRoot
	projectRoot = fixtureProjectRoot
	t.Cleanup(func() { projectRoot = saved })

	for _, f := range allFixtures() {
		f := f
		t.Run(f.name, func(t *testing.T) {
			suite := f.build()
			if len(suite.SpecResults) == 0 {
				t.Skip("fixture has no spec to render")
			}
			var buf bytes.Buffer
			if err := RenderSpec(&buf, suite, suite.SpecResults[0]); err != nil {
				t.Fatalf("RenderSpec: %v", err)
			}
			compareGolden(t, "_testdata/"+f.name+"_spec.md", buf.Bytes())
		})
	}
}

// compareGolden reads the path and asserts equality. With -update it writes
// the supplied bytes to disk instead.
func compareGolden(t *testing.T, path string, got []byte) {
	t.Helper()
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, got, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run `go test -update ./mdgen/...` to create)", path, err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("golden %s mismatch (re-run with -update to refresh after deliberate format change)\n%s",
			path, diff(string(want), string(got)))
	}
}

// diff returns a tiny line-by-line diff sufficient for reading test failures
// without pulling in go-cmp. Mismatched lines are flagged, surrounding
// matched lines kept for context.
func diff(want, got string) string {
	wl := strings.Split(want, "\n")
	gl := strings.Split(got, "\n")
	var b strings.Builder
	max := len(wl)
	if len(gl) > max {
		max = len(gl)
	}
	for i := 0; i < max; i++ {
		var wline, gline string
		if i < len(wl) {
			wline = wl[i]
		}
		if i < len(gl) {
			gline = gl[i]
		}
		if wline == gline {
			b.WriteString("  " + wline + "\n")
		} else {
			b.WriteString("- " + wline + "\n")
			b.WriteString("+ " + gline + "\n")
		}
	}
	return b.String()
}
