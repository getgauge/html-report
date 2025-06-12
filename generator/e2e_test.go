/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/html-report/env"
	helper "github.com/getgauge/html-report/test_helper"
)

var suiteRes3 = newProtoSuiteRes(true, 1, 1, 60, nil, nil, passSpecRes1, failSpecResWithStepFailure, skippedSpecRes)
var suiteResWithBeforeSuiteFailure = newProtoSuiteRes(true, 0, 0, 0, newProtoHookFailure(), newProtoHookFailure())
var templateBasePath, _ = filepath.Abs(filepath.Join("..", "themes", "default"))

func TestEndToEndHTMLGenerationWhenBeforeSuiteFails(t *testing.T) {
	reportDir := filepath.Join("_testdata", "e2e")
	r := ToSuiteResult("", suiteResWithBeforeSuiteFailure)
	err := GenerateReports(r, reportDir, templateBasePath, true)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}
	gotContent, err := os.ReadFile(filepath.Join(reportDir, "index.html"))
	if err != nil {
		t.Errorf("Error reading generated HTML file: %s", err.Error())
	}
	wantContent, err := os.ReadFile(filepath.Join("_testdata", "expectedE2E", "before_suite_fail.html"))
	if err != nil {
		t.Errorf("Error reading expected HTML file: %s", err.Error())
	}
	got := helper.RemoveNewline(string(gotContent))
	want := helper.RemoveNewline(string(wantContent))
	helper.AssertEqual(want, got, "index.html", t)
	cleanUp(t, reportDir)
}

func TestEndToEndHTMLGeneration(t *testing.T) {
	expectedFiles := []string{"index.html", "passing_specification_1.html", "failing_specification_1.html", "skipped_specification.html", "js/search_index.js"}
	reportDir := filepath.Join("_testdata", "e2e")

	r := ToSuiteResult("", suiteRes3)
	err := GenerateReports(r, reportDir, templateBasePath, true)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}

	verifyExpectedFiles(t, "simpleSuiteRes", reportDir, expectedFiles)
	cleanUp(t, reportDir)
}

func TestEndToEndMinifiedHTMLGeneration(t *testing.T) {
	expectedFiles := []string{"index.html", "passing_specification_1.html", "failing_specification_1.html", "skipped_specification.html", "js/search_index.js"}
	reportDir := filepath.Join("_testdata", "e2e")
	helper.SetEnvOrFail(t, "gauge_minify_reports", "true")
	r := ToSuiteResult("", suiteRes3)
	err := GenerateReports(r, reportDir, templateBasePath, true)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}

	verifyExpectedFiles(t, "minifiedSimpleSuiteRes", reportDir, expectedFiles)
	helper.UnsetEnvOrFail(t, "gauge_minify_reports")
	cleanUp(t, reportDir)
}

func TestEndToEndHTMLGenerationWithPreAndPostHookScreenshots(t *testing.T) {
	expectedFiles := []string{"index.html", "passing_specification_1.html", "failing_specification_1.html", "skipped_specification.html", "js/search_index.js"}
	reportDir := filepath.Join("_testdata", "e2e")
	suiteRes := newProtoSuiteRes(true, 1, 1, 60, nil, nil, passSpecRes1, failSpecResWithStepFailure, skippedSpecRes)
	suiteRes.PreHookScreenshotFiles = []string{"pre-hook-screenshot-1.png", "pre-hook-screenshot-2.png"}
	suiteRes.PostHookScreenshotFiles = []string{"post-hook-screenshot-1.png", "post-hook-screenshot-2.png"}
	r := ToSuiteResult("", suiteRes)
	err := GenerateReports(r, reportDir, templateBasePath, true)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}

	verifyExpectedFiles(t, "simpleSuiteResWithHookScreenshots", reportDir, expectedFiles)
	cleanUp(t, reportDir)
}

func TestEndToEndHTMLGenerationForThemeWithRelativePath(t *testing.T) {
	expectedFiles := []string{"index.html", "passing_specification_1.html", "failing_specification_1.html", "skipped_specification.html", "js/search_index.js"}
	reportDir := filepath.Join("_testdata", "e2e")
	defaultThemePath := filepath.Join("..", "themes", "default")

	r := ToSuiteResult("", suiteRes3)
	err := GenerateReports(r, reportDir, defaultThemePath, true)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}

	verifyExpectedFiles(t, "simpleSuiteRes", reportDir, expectedFiles)
	cleanUp(t, reportDir)
}

func TestEndToEndHTMLGenerationForCustomTheme(t *testing.T) {
	expectedFiles := []string{"index.html", "passing_specification_1.html", "failing_specification_1.html", "skipped_specification.html", "js/search_index.js"}
	reportDir := filepath.Join("_testdata", "e2e")
	defaultThemePath := filepath.Join("_testdata", "dummyReportTheme")

	r := ToSuiteResult("", suiteRes3)
	err := GenerateReports(r, reportDir, defaultThemePath, true)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}

	verifyExpectedFiles(t, "simpleSuiteRes", reportDir, expectedFiles)
	cleanUp(t, reportDir)
}

func TestEndToEndHTMLGenerationForNestedSpecs(t *testing.T) {
	helper.SetEnvOrFail(t, env.UseNestedSpecs, "true")
	var suiteRes4 = newProtoSuiteRes(false, 0, 0, 100, nil, nil, passSpecRes1, nestedSpecRes)
	expectedFiles := []string{
		"index.html",
		"passing_specification_1.html",
		filepath.Join("nested", "nested_specification.html"),
		filepath.Join("nested", "index.html"),
		"js/search_index.js",
	}
	reportDir := filepath.Join("_testdata", "e2e")

	r := ToSuiteResult("", suiteRes4)
	err := GenerateReports(r, reportDir, templateBasePath, true)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}
	verifyExpectedFiles(t, "nestedSuiteRes", reportDir, expectedFiles)
	cleanUp(t, reportDir)
}

func cleanUp(t *testing.T, reportDir string) {
	s, err := filepath.Glob(filepath.Join(reportDir, "*"))
	if err != nil {
		t.Error(err)
	}
	for _, f := range s {
		if f != filepath.Join(reportDir, ".gitkeep") {
			if err := os.RemoveAll(f); err != nil {
				t.Errorf("Failed to remove file %s: %v", f, err)
			}
		}
	}
}

func verifyExpectedFiles(t *testing.T, suiteRes, reportDir string, expectedFiles []string) {
	for _, expectedFile := range expectedFiles {
		gotContent, err := os.ReadFile(filepath.Join(reportDir, expectedFile))
		if err != nil {
			t.Errorf("Error reading generated HTML file: %s", err.Error())
		}
		wantContent, err := os.ReadFile(filepath.Join("_testdata", "expectedE2E", suiteRes, expectedFile))
		if err != nil {
			t.Errorf("Error reading expected HTML file: %s", err.Error())
		}
		got := helper.RemoveNewline(string(gotContent))
		want := helper.RemoveNewline(string(wantContent))
		helper.AssertEqual(want, got, expectedFile, t)
	}
}
