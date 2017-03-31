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
package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	htmldiff "github.com/documize/html-diff"
)

var re = regexp.MustCompile("[\\s]*[\n\t][\\s]*")

func TestEndToEndHTMLGenerationFromSavedResult(t *testing.T) {
	expectedFiles := []string{"index.html", "passing_specification_1.html", "failing_specification_1.html", "skipped_specification.html", "js/search_index.js"}
	reportDir := filepath.Join("generator", "_testdata", "e2e")
	inputFile := filepath.Join("generator", "_testdata", "last_run_result.json")

	regenerateReport(inputFile, reportDir, "default")
	for _, expectedFile := range expectedFiles {
		gotContent, err := ioutil.ReadFile(filepath.Join(reportDir, expectedFile))
		if err != nil {
			t.Errorf("Error reading generated HTML file: %s", err.Error())
		}
		wantContent, err := ioutil.ReadFile(filepath.Join("generator", "_testdata", "expectedE2E", expectedFile))
		if err != nil {
			t.Errorf("Error reading expected HTML file: %s", err.Error())
		}
		got := removeNewline(string(gotContent))
		want := removeNewline(string(wantContent))
		os.Remove(filepath.Join(reportDir, expectedFile))
		assertEqual(want, got, expectedFile, t)
	}
	cleanUp(t)
}

func cleanUp(t *testing.T) {
	reportDir := filepath.Join("generator", "_testdata", "e2e")
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

func assertEqual(expected, actual, testName string, t *testing.T) {
	if expected != actual {
		diffHTML := compare(expected, actual)
		tmpFile, err := ioutil.TempFile("", "")
		if err != nil {
			t.Errorf("Unable to dump to tmp file. Raw content:\n%s\n", diffHTML)
		}
		fileName := fmt.Sprintf("%s.html", tmpFile.Name())
		ioutil.WriteFile(fileName, []byte(diffHTML), 0644)
		tmpFile.Close()
		t.Errorf("%s -  View Diff Output : %s\n", testName, fileName)
	}
}
func compare(a, b string) string {
	var cfg = &htmldiff.Config{
		InsertedSpan: []htmldiff.Attribute{{Key: "style", Val: "background-color: palegreen;"}},
		DeletedSpan:  []htmldiff.Attribute{{Key: "style", Val: "background-color: lightpink;"}},
		ReplacedSpan: []htmldiff.Attribute{{Key: "style", Val: "background-color: lightskyblue;"}},
		CleanTags:    []string{""},
	}

	res, _ := cfg.HTMLdiff([]string{html.EscapeString(a), html.EscapeString(b)})
	return res[0]
}

func removeNewline(s string) string {
	return re.ReplaceAllLiteralString(s, "")
}
