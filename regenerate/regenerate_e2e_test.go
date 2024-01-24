package regenerate

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	helper "github.com/getgauge/html-report/test_helper"
	"google.golang.org/protobuf/proto"
)

var templateBasePath, _ = filepath.Abs(filepath.Join("..", "themes", "default"))

// setup converts _testdata/last_run_result.json to _testdata/last_run_result (binary serialized)
// editing json is easy, helps in maintaining the setup data
// but the actual regenerate requires serialized proto data, hence this conversion.
func setup() {
	inputFile := filepath.Join("_testdata", "last_run_result.json")
	b, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	psr := &gauge_messages.ProtoSuiteResult{}
	err = json.Unmarshal(b, psr)
	if err != nil {
		log.Fatalf("Unable to read last run data from %s. Error: %s", inputFile, err.Error())
	}
	by, _ := proto.Marshal(psr)
	f := filepath.Join("_testdata", "last_run_result")
	err = os.WriteFile(f, by, 0644)
	if err != nil {
		log.Fatalf("Unable to write file %s. Error: %s", f, err.Error())
	}
}

func TestEndToEndHTMLGenerationFromSavedResult(t *testing.T) {
	setup()
	expectedFiles := []string{"index.html", "specs/example.html", "js/search_index.js"}
	reportDir := filepath.Join("_testdata", "e2e")
	inputFile := filepath.Join("_testdata", "last_run_result")

	Report(inputFile, reportDir, templateBasePath, "/tmp/foo/")
	for _, expectedFile := range expectedFiles {
		gotContent, err := os.ReadFile(filepath.Join(reportDir, expectedFile))
		if err != nil {
			t.Errorf("Error reading generated HTML file: %s", err.Error())
		}
		wantContent, err := os.ReadFile(filepath.Join("_testdata", "expectedE2E", "simpleSuiteRes", expectedFile))
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
