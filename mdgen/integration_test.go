/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package mdgen

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/getgauge/html-report/env"
)

// withProjectRoot stamps the package-level projectRoot for the duration of a
// test. Mirrors what ToSuiteResult does in production.
func withProjectRoot(t *testing.T, root string) {
	t.Helper()
	saved := projectRoot
	projectRoot = root
	t.Cleanup(func() { projectRoot = saved })
}

// withScreenshotFiles seeds the package-level screenshotFiles slice and
// restores it on teardown — mirrors the side effect of toStep / toHookFailure.
func withScreenshotFiles(t *testing.T, files []string) {
	t.Helper()
	saved := screenshotFiles
	screenshotFiles = files
	t.Cleanup(func() { screenshotFiles = saved })
}

func TestGenerateReports_writesIndexAndPerSpec(t *testing.T) {
	withProjectRoot(t, "/proj")

	suite := &SuiteResult{
		ProjectName:         "demo",
		Timestamp:           "Jan 2, 2026 at 3:04pm",
		Environment:         "default",
		ExecutionTime:       1500,
		PassedSpecsCount:    1,
		FailedSpecsCount:    1,
		PassedScenarioCount: 1,
		FailedScenarioCount: 1,
		ExecutionStatus:     fail,
		SpecResults: []*spec{
			{
				SpecHeading:         "Login flow",
				FileName:            "/proj/specs/login.spec",
				ExecutionTime:       1000,
				ExecutionStatus:     pass,
				PassedScenarioCount: 1,
				Scenarios: []*scenario{
					{Heading: "ok", ExecutionStatus: pass, Items: []item{{Kind: stepKind, Step: &step{
						Fragments: []*fragment{textFrag("login")}, Result: &result{Status: pass},
					}}}},
				},
			},
			{
				SpecHeading:         "Checkout",
				FileName:            "/proj/specs/sub/checkout.spec",
				ExecutionTime:       500,
				ExecutionStatus:     fail,
				FailedScenarioCount: 1,
				Scenarios: []*scenario{
					{Heading: "bad", ExecutionStatus: fail, Items: []item{{Kind: stepKind, Step: &step{
						Fragments: []*fragment{textFrag("checkout")},
						Result:    &result{Status: fail, ErrorMessage: "boom"},
					}}}},
				},
			},
		},
	}

	tmp := t.TempDir()
	if err := GenerateReports(suite, tmp); err != nil {
		t.Fatalf("GenerateReports: %v", err)
	}

	// index.md exists at root.
	indexBytes, err := os.ReadFile(filepath.Join(tmp, "index.md"))
	if err != nil {
		t.Fatalf("read index: %v", err)
	}
	indexText := string(indexBytes)

	// Per-spec files exist at the paths the index links to.
	for _, want := range []string{"specs/login.md", "specs/sub/checkout.md"} {
		if _, err := os.Stat(filepath.Join(tmp, want)); err != nil {
			t.Errorf("expected spec file %s missing: %v", want, err)
		}
		if !strings.Contains(indexText, want) {
			t.Errorf("index.md missing link target %q", want)
		}
	}

	// Each link in the index should resolve to a real file on disk — this
	// catches mismatches between indexLinkHref and the writer's path logic.
	for _, target := range mdLinkTargets(indexText) {
		if !strings.HasSuffix(target, ".md") {
			continue
		}
		full := filepath.Join(tmp, target)
		if _, err := os.Stat(full); err != nil {
			t.Errorf("index link %q points to missing file %s: %v", target, full, err)
		}
	}
}

// mdLinkRe extracts the href portion of a Markdown inline link.
var mdLinkRe = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

func mdLinkTargets(s string) []string {
	matches := mdLinkRe.FindAllStringSubmatch(s, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, m[1])
	}
	return out
}

func TestGenerateReports_beforeSuiteHookFailure_skipsSpecPages(t *testing.T) {
	withProjectRoot(t, "/proj")

	// Even with spec results present, a Before Suite failure means they
	// never executed; we should write only index.md.
	suite := &SuiteResult{
		ProjectName:            "demo",
		BeforeSuiteHookFailure: &hookFailure{HookName: "Before Suite", ErrMsg: "down"},
		SpecResults: []*spec{
			{SpecHeading: "x", FileName: "/proj/specs/x.spec"},
		},
	}

	tmp := t.TempDir()
	if err := GenerateReports(suite, tmp); err != nil {
		t.Fatalf("GenerateReports: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "index.md")); err != nil {
		t.Errorf("index.md missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, "specs", "x.md")); !os.IsNotExist(err) {
		t.Errorf("spec page should NOT be written when before-suite hook failed; err=%v", err)
	}
}

func TestGenerateReports_nestedSpecs_writesPerDirIndex(t *testing.T) {
	withProjectRoot(t, "/proj")
	t.Setenv(env.UseNestedSpecs, "true")

	suite := &SuiteResult{
		ProjectName: "demo",
		SpecResults: []*spec{
			{SpecHeading: "auth", FileName: "/proj/specs/auth/login.spec", ExecutionStatus: pass},
			{SpecHeading: "billing", FileName: "/proj/specs/billing/checkout.spec", ExecutionStatus: pass},
		},
	}

	tmp := t.TempDir()
	if err := GenerateReports(suite, tmp); err != nil {
		t.Fatalf("GenerateReports: %v", err)
	}

	for _, want := range []string{
		filepath.Join("specs", "index.md"),
		filepath.Join("specs", "auth", "index.md"),
		filepath.Join("specs", "billing", "index.md"),
	} {
		if _, err := os.Stat(filepath.Join(tmp, want)); err != nil {
			t.Errorf("expected nested index %s missing: %v", want, err)
		}
	}
}

func TestGenerateReports_overwritesExistingDir(t *testing.T) {
	withProjectRoot(t, "/proj")

	tmp := t.TempDir()
	// Pre-populate with a stale file the regenerate should overwrite.
	stale := filepath.Join(tmp, "index.md")
	if err := os.WriteFile(stale, []byte("STALE CONTENT"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	suite := &SuiteResult{ProjectName: "demo", SpecResults: []*spec{}}
	if err := GenerateReports(suite, tmp); err != nil {
		t.Fatalf("GenerateReports: %v", err)
	}

	got, err := os.ReadFile(stale)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if strings.Contains(string(got), "STALE CONTENT") {
		t.Errorf("expected index.md to be overwritten, got %q", got)
	}
}

func TestGenerateReports_copiesScreenshots(t *testing.T) {
	withProjectRoot(t, "/proj")

	// Lay out a synthetic screenshots dir; transform.go's globals capture the
	// filenames during proto translation, so we seed the same slice here.
	srcDir := t.TempDir()
	shotPath := filepath.Join(srcDir, "scn-1.png")
	if err := os.WriteFile(shotPath, []byte("PNG-BYTES"), 0o644); err != nil {
		t.Fatalf("seed shot: %v", err)
	}
	t.Setenv(env.ScreenshotsDirName, srcDir)
	withScreenshotFiles(t, []string{"scn-1.png"})

	suite := &SuiteResult{ProjectName: "demo"}
	tmp := t.TempDir()
	if err := GenerateReports(suite, tmp); err != nil {
		t.Fatalf("GenerateReports: %v", err)
	}

	dst := filepath.Join(tmp, "images", "scn-1.png")
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("expected screenshot copied to %s, got err=%v", dst, err)
	}
}
