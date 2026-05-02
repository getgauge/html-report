/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package regenerate

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"google.golang.org/protobuf/proto"
)

// setup converts _testdata/last_run_result.json to _testdata/last_run_result
// (binary serialized proto). Editing JSON keeps the test fixture human-friendly
// while the regenerate flow consumes proto.
func setup() {
	inputFile := filepath.Join("_testdata", "last_run_result.json")
	b, err := os.ReadFile(inputFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	psr := &gauge_messages.ProtoSuiteResult{}
	if err := json.Unmarshal(b, psr); err != nil {
		log.Fatalf("Unable to read last run data from %s. Error: %s", inputFile, err.Error())
	}
	by, _ := proto.Marshal(psr)
	f := filepath.Join("_testdata", "last_run_result")
	if err := os.WriteFile(f, by, 0644); err != nil {
		log.Fatalf("Unable to write file %s. Error: %s", f, err.Error())
	}
}

// TestEndToEndMarkdownGenerationFromSavedResult exercises the regenerate
// pipeline against a real last_run_result. It checks that the produced
// directory contains the expected Markdown skeleton; full output shape is
// pinned by golden tests in mdgen.
func TestEndToEndMarkdownGenerationFromSavedResult(t *testing.T) {
	setup()
	reportDir := filepath.Join("_testdata", "e2e")
	inputFile := filepath.Join("_testdata", "last_run_result")
	t.Cleanup(func() { cleanUp(t, reportDir) })

	Report(inputFile, reportDir, "/tmp/foo/")

	indexBytes, err := os.ReadFile(filepath.Join(reportDir, "index.md"))
	if err != nil {
		t.Fatalf("expected index.md, got error: %v", err)
	}
	got := string(indexBytes)
	for _, marker := range []string{"# Gauge Report", "## Summary", "## Specs"} {
		if !strings.Contains(got, marker) {
			t.Errorf("regenerated index.md missing expected marker %q", marker)
		}
	}

	// At least one spec page should have been written under specs/.
	specsDir := filepath.Join(reportDir, "specs")
	entries, err := os.ReadDir(specsDir)
	if err != nil {
		t.Fatalf("expected specs/ directory, got error: %v", err)
	}
	if !hasMarkdownChild(entries) {
		t.Errorf("expected at least one .md file under %s", specsDir)
	}
}

func hasMarkdownChild(entries []os.DirEntry) bool {
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			return true
		}
	}
	return false
}

func cleanUp(t *testing.T, reportDir string) {
	s, err := filepath.Glob(filepath.Join(reportDir, "*"))
	if err != nil {
		t.Error(err)
	}
	for _, f := range s {
		if f != filepath.Join(reportDir, ".gitkeep") {
			if err := os.RemoveAll(f); err != nil {
				return
			}
		}
	}
}
