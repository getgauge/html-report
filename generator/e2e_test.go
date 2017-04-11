// Copyright 2015 ThoughtWorks, Inc.

// This file is part of getgauge/html-report.

// getgauge/html-report is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// getgauge/html-report is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with getgauge/html-report.  If not, see <http://www.gnu.org/licenses/>.

package generator

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	helper "github.com/getgauge/html-report/test_helper"
)

var suiteRes3 = newProtoSuiteRes(true, 1, 1, 60, nil, nil, passSpecRes1, failSpecResWithStepFailure, skippedSpecRes)
var suiteResWithBeforeSuiteFailure = newProtoSuiteRes(true, 0, 0, 0, newProtoHookFailure(), nil)
var templateBasePath, _ = filepath.Abs(filepath.Join("..", "themes", "default"))

func TestEndToEndHTMLGenerationWhenBeforeSuiteFails(t *testing.T) {
	reportsDir = filepath.Join("_testdata", "e2e")
	r := ToSuiteResult("", suiteResWithBeforeSuiteFailure)
	err := GenerateReports(r, templateBasePath)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}
	gotContent, err := ioutil.ReadFile(filepath.Join(reportsDir, "index.html"))
	if err != nil {
		t.Errorf("Error reading generated HTML file: %s", err.Error())
	}
	wantContent, err := ioutil.ReadFile(filepath.Join("_testdata", "expectedE2E", "before_suite_fail.html"))
	if err != nil {
		t.Errorf("Error reading expected HTML file: %s", err.Error())
	}
	got := helper.RemoveNewline(string(gotContent))
	want := helper.RemoveNewline(string(wantContent))
	helper.AssertEqual(want, got, "index.html", t)
	cleanUp(t)
}

func TestEndToEndHTMLGeneration(t *testing.T) {
	expectedFiles := []string{"index.html", "passing_specification_1.html", "failing_specification_1.html", "skipped_specification.html", "js/search_index.js"}
	reportsDir = filepath.Join("_testdata", "e2e")

	r := ToSuiteResult("", suiteRes3)
	err := GenerateReports(r, templateBasePath)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}

	verifyExpectedFiles(t, "simpleSuiteRes", reportsDir, expectedFiles)
	cleanUp(t)
}

func TestEndToEndHTMLGenerationForThemeWithRelativePath(t *testing.T) {
	expectedFiles := []string{"index.html", "passing_specification_1.html", "failing_specification_1.html", "skipped_specification.html", "js/search_index.js"}
	reportsDir = filepath.Join("_testdata", "e2e")
	defaultThemePath := filepath.Join("..", "themes", "default")

	r := ToSuiteResult("", suiteRes3)
	err := GenerateReports(r, defaultThemePath)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}

	verifyExpectedFiles(t, "simpleSuiteRes", reportsDir, expectedFiles)
	cleanUp(t)
}

func TestEndToEndHTMLGenerationForCustomTheme(t *testing.T) {
	expectedFiles := []string{"index.html", "passing_specification_1.html", "failing_specification_1.html", "skipped_specification.html", "js/search_index.js"}
	reportsDir := filepath.Join("_testdata", "e2e")
	defaultThemePath := filepath.Join("_testdata", "dummyReportTheme")

	r := ToSuiteResult("", suiteRes3)
	err := GenerateReports(r, defaultThemePath)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}

	verifyExpectedFiles(t, "simpleSuiteRes", reportsDir, expectedFiles)
	cleanUp(t)
}

func TestEndToEndHTMLGenerationFromSavedResult(t *testing.T) {
	expectedFiles := []string{"index.html", "passing_specification_1.html", "failing_specification_1.html", "skipped_specification.html", "js/search_index.js"}
	reportDir := filepath.Join("_testdata", "e2e")
	inputFile := filepath.Join("_testdata", "last_run_result.json")

	RegenerateReport(inputFile, reportDir, templateBasePath)
	for _, expectedFile := range expectedFiles {
		gotContent, err := ioutil.ReadFile(filepath.Join(reportDir, expectedFile))
		if err != nil {
			t.Errorf("Error reading generated HTML file: %s", err.Error())
		}
		wantContent, err := ioutil.ReadFile(filepath.Join("_testdata", "expectedE2E", "simpleSuiteRes", expectedFile))
		if err != nil {
			t.Errorf("Error reading expected HTML file: %s", err.Error())
		}
		got := helper.RemoveNewline(string(gotContent))
		want := helper.RemoveNewline(string(wantContent))
		helper.AssertEqual(want, got, expectedFile, t)
	}
	cleanUp(t)
}

func TestEndToEndHTMLGenerationForNestedSpecs(t *testing.T) {
	var suiteRes4 = newProtoSuiteRes(false, 0, 0, 100, nil, nil, passSpecRes1, nestedSpecRes)
	expectedFiles := []string{
		"index.html",
		"passing_specification_1.html",
		filepath.Join("nested", "nested_specification.html"),
		filepath.Join("nested", "index.html"),
		"js/search_index.js",
	}
	reportsDir = filepath.Join("_testdata", "e2e")
	
	r := ToSuiteResult("", suiteRes4)
	err := GenerateReports(r, templateBasePath)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}
	verifyExpectedFiles(t, "nestedSuiteRes", reportsDir, expectedFiles)
	cleanUp(t)
}

func cleanUp(t *testing.T) {
	s, err := filepath.Glob(filepath.Join(reportsDir, "*"))
	if err != nil {
		t.Error(err)
	}
	for _, f := range s {
		if f != filepath.Join(reportsDir, ".gitkeep") {
			os.RemoveAll(f)
		}
	}
}

func verifyExpectedFiles(t *testing.T, suiteRes, reportDir string, expectedFiles []string) {
	for _, expectedFile := range expectedFiles {
		gotContent, err := ioutil.ReadFile(filepath.Join(reportDir, expectedFile))
		if err != nil {
			t.Errorf("Error reading generated HTML file: %s", err.Error())
		}
		wantContent, err := ioutil.ReadFile(filepath.Join("_testdata", "expectedE2E", suiteRes, expectedFile))
		if err != nil {
			t.Errorf("Error reading expected HTML file: %s", err.Error())
		}
		got := helper.RemoveNewline(string(gotContent))
		want := helper.RemoveNewline(string(wantContent))
		helper.AssertEqual(want, got, expectedFile, t)
	}
}
