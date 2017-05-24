package regenerate

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	helper "github.com/getgauge/html-report/test_helper"
)

var templateBasePath, _ = filepath.Abs(filepath.Join("..", "themes", "default"))

func TestEndToEndHTMLGenerationFromSavedResult(t *testing.T) {
	expectedFiles := []string{"index.html", "example.html", "js/search_index.js"}
	reportDir := filepath.Join("_testdata", "e2e")
	inputFile := filepath.Join("_testdata", "last_run_result")

	RegenerateReport(inputFile, reportDir, templateBasePath, "")
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
	cleanUp(t, reportDir)
}

func cleanUp(t *testing.T, reportDir string) {
	s, err := filepath.Glob(filepath.Join(reportDir, "*"))
	if err != nil {
		t.Error(err)
	}
	for _, f := range s {
		if f != filepath.Join(reportDir, ".gitkeep") {
			os.RemoveAll(f)
		}
	}
}
