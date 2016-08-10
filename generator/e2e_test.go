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
)

var suiteRes3 = newProtoSuiteRes(true, 1, 1, 60, nil, nil, passSpecRes1, failSpecResWithStepFailure, skippedSpecRes)
var suiteResWithBeforeSuiteFailure = newProtoSuiteRes(true, 0, 0, 0, newProtoHookFailure(), nil)

func TestEndToEndHTMLGenerationWhenBeforeSuiteFails(t *testing.T) {
	reportDir := filepath.Join("_testdata", "e2e")
	ProjectRoot = ""

	err := GenerateReports(suiteResWithBeforeSuiteFailure, reportDir)

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
	got := removeNewline(string(gotContent))
	want := removeNewline(string(wantContent))
	os.Remove(filepath.Join(reportDir, "index.html"))
	if got != want {
		t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
	}
}

func TestEndToEndHTMLGeneration(t *testing.T) {
	expectedFiles := []string{"index.html", "passing_specification_1.html", "failing_specification_1.html", "skipped_specification.html", "search_index.json"}
	reportDir := filepath.Join("_testdata", "e2e")
	ProjectRoot = ""

	err := GenerateReports(suiteRes3, reportDir)

	if err != nil {
		t.Errorf("Expected error to be nil. Got: %s", err.Error())
	}
	for _, expectedFile := range expectedFiles {
		gotContent, err := ioutil.ReadFile(filepath.Join(reportDir, expectedFile))
		if err != nil {
			t.Errorf("Error reading generated HTML file: %s", err.Error())
		}
		wantContent, err := ioutil.ReadFile(filepath.Join("_testdata", "expectedE2E", expectedFile))
		if err != nil {
			t.Errorf("Error reading expected HTML file: %s", err.Error())
		}
		got := removeNewline(string(gotContent))
		want := removeNewline(string(wantContent))
		os.Remove(filepath.Join(reportDir, expectedFile))
		if got != want {
			t.Errorf("want:\n%q\ngot:\n%q\n", want, got)
		}
	}
}
