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
var templateBasePath,_ = filepath.Abs(filepath.Join("..","themes", "default"))

func TestEndToEndHTMLGenerationWhenBeforeSuiteFails(t *testing.T) {
	reportDir := filepath.Join("_testdata", "e2e")
	r := ToSuiteResult("", suiteResWithBeforeSuiteFailure)
	err := GenerateReports(r, reportDir, templateBasePath)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}
	gotContent, err := ioutil.ReadFile(filepath.Join(reportDir, "index.html"))
	if err != nil {
		t.Errorf("Error reading generated HTML file: %s", err.Error())
	}
	wantContent, err := ioutil.ReadFile(filepath.Join("_testdata", "expectedE2E", "before_suite_fail.html"))
	if err != nil {
		t.Errorf("Error reading expected HTML file: %s", err.Error())
	}
	got := helper.RemoveNewline(string(gotContent))
	want := helper.RemoveNewline(string(wantContent))
	os.Remove(filepath.Join(reportDir, "index.html"))
	helper.AssertEqual(want, got, "index.html", t)
}

func TestEndToEndHTMLGeneration(t *testing.T) {
	expectedFiles := []string{"index.html", "passing_specification_1.html", "failing_specification_1.html", "skipped_specification.html", "js/search_index.js"}
	reportDir := filepath.Join("_testdata", "e2e")

	r := ToSuiteResult("", suiteRes3)
	err := GenerateReports(r, reportDir, templateBasePath)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}

	verifyExpectedFiles(t, reportDir, expectedFiles)
}

func TestEndToEndHTMLGenerationForThemeWithRelativePath(t *testing.T) {
	expectedFiles := []string{"index.html", "passing_specification_1.html", "failing_specification_1.html", "skipped_specification.html", "js/search_index.js"}
	reportDir := filepath.Join("_testdata", "e2e")
	defaultThemePath := filepath.Join("..", "themes", "default")

	r := ToSuiteResult("", suiteRes3)
	err := GenerateReports(r, reportDir, defaultThemePath)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}
	
	verifyExpectedFiles(t, reportDir, expectedFiles)
}

func TestEndToEndHTMLGenerationForCustomTheme(t *testing.T) {
	expectedFiles := []string{"index.html", "passing_specification_1.html", "failing_specification_1.html", "skipped_specification.html", "js/search_index.js"}
	reportDir := filepath.Join("_testdata", "e2e")
	defaultThemePath := filepath.Join("_testdata", "dummyReportTheme")

	r := ToSuiteResult("", suiteRes3)
	err := GenerateReports(r, reportDir, defaultThemePath)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}

	verifyExpectedFiles(t, reportDir, expectedFiles)
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
		wantContent, err := ioutil.ReadFile(filepath.Join("_testdata", "expectedE2E", expectedFile))
		if err != nil {
			t.Errorf("Error reading expected HTML file: %s", err.Error())
		}
		got := helper.RemoveNewline(string(gotContent))
		want := helper.RemoveNewline(string(wantContent))
		os.Remove(filepath.Join(reportDir, expectedFile))
		helper.AssertEqual(want, got, expectedFile, t)
	}
	cleanUp(t)
}

func cleanUp(t *testing.T) {
	reportDir := filepath.Join("_testdata", "e2e")
	os.RemoveAll(filepath.Join(reportDir, "css"))
	os.RemoveAll(filepath.Join(reportDir, "images"))
	os.RemoveAll(filepath.Join(reportDir, "fonts"))
	s, err := filepath.Glob(filepath.Join(reportDir, "js", "*.js"))
	if err != nil {
		t.Error(err)
	}
	for _, f := range s {
		os.Remove(f)
	}
}

func verifyExpectedFiles(t *testing.T, reportDir string, expectedFiles []string) {
	for _, expectedFile := range expectedFiles {
		gotContent, err := ioutil.ReadFile(filepath.Join(reportDir, expectedFile))
		if err != nil {
			t.Errorf("Error reading generated HTML file: %s", err.Error())
		}
		wantContent, err := ioutil.ReadFile(filepath.Join("_testdata", "expectedE2E", expectedFile))
		if err != nil {
			t.Errorf("Error reading expected HTML file: %s", err.Error())
		}
		got := helper.RemoveNewline(string(gotContent))
		want := helper.RemoveNewline(string(wantContent))
		os.Remove(filepath.Join(reportDir, expectedFile))
		helper.AssertEqual(want, got, expectedFile, t)
	}
	
}