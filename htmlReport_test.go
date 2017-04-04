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
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"io/ioutil"

	"github.com/getgauge/html-report/generator"
)

var now = time.Now()

type testNameGenerator struct {
}

func init() {
	generator.TemplateBasePath = "themes"
}

func (T testNameGenerator) randomName() string {
	return now.Format(timeFormat)
}

func TestCopyingReportTemplates(t *testing.T) {
	dirToCopy := filepath.Join(os.TempDir(), randomName())
	defer os.RemoveAll(dirToCopy)

	err := generator.CopyReportTemplateFiles(getThemePath(), dirToCopy)
	if err != nil {
		t.Errorf("Expected error == nil, got: %s \n", err.Error())
	}
	verifyReportTemplateFilesAreCopied(dirToCopy, t)
}

func TestGetReportsDirectory(t *testing.T) {
	userSetReportsDir := filepath.Join(os.TempDir(), randomName())
	os.Setenv(gaugeReportsDirEnvName, userSetReportsDir)
	expectedReportsDir := filepath.Join(userSetReportsDir, htmlReport)
	defer os.RemoveAll(userSetReportsDir)

	reportsDir := getReportsDirectory(nil)

	if reportsDir != expectedReportsDir {
		t.Errorf("Expected reportsDir == %s, got: %s\n", expectedReportsDir, reportsDir)
	}
	if !fileExists(expectedReportsDir) {
		t.Errorf("Expected %s report directory doesn't exist", expectedReportsDir)
	}
}

func TestGetReportsDirectoryWithOverrideFlag(t *testing.T) {
	userSetReportsDir := filepath.Join(os.TempDir(), randomName())
	os.Setenv(gaugeReportsDirEnvName, userSetReportsDir)
	os.Setenv(overwriteReportsEnvProperty, "true")
	nameGen := &testNameGenerator{}
	expectedReportsDir := filepath.Join(userSetReportsDir, htmlReport, nameGen.randomName())
	defer os.RemoveAll(userSetReportsDir)

	reportsDir := getReportsDirectory(nameGen)

	if reportsDir != expectedReportsDir {
		t.Errorf("Expected reportsDir == %s, got: %s\n", expectedReportsDir, reportsDir)
	}
	if !fileExists(expectedReportsDir) {
		t.Errorf("Expected %s report directory doesn't exist", expectedReportsDir)
	}
}

func randomName() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func verifyReportTemplateFilesAreCopied(dest string, t *testing.T) {
	reportDir := filepath.Join(getThemePath(), "assets")
	filepath.Walk(reportDir, func(path string, info os.FileInfo, err error) error {
		path = strings.Replace(path, reportDir, "", 1)
		destFilePath := filepath.Join(dest, path)
		if !fileExists(destFilePath) {
			t.Errorf("File %s not copied.", destFilePath)
		}
		return nil
	})
}

func TestCreatingReportShouldOverwriteReportsBasedOnEnv(t *testing.T) {
	os.Setenv(overwriteReportsEnvProperty, "true")
	nameGen := getNameGen()
	if nameGen != nil {
		t.Errorf("Expected nameGen == nil, got %s", nameGen)
	}

	os.Setenv(overwriteReportsEnvProperty, "false")
	nameGen = getNameGen()
	switch nameGen.(type) {
	case timeStampedNameGenerator:
	default:
		t.Errorf("Expected nameGen to be type timeStampedNameGenerator, got %s", reflect.TypeOf(nameGen))
	}
}

func TestSaveLastExecutionResult(t *testing.T) {
	reportsDir := filepath.Join(os.TempDir(), randomName())
	os.MkdirAll(reportsDir, 0755)
	defer os.RemoveAll(reportsDir)
	res := &generator.SuiteResult{ProjectName: "foo"}

	saveLastExecutionResult(res, reportsDir)

	outF := filepath.Join(reportsDir, resultFile)

	o, err := ioutil.ReadFile(outF)
	if err != nil {
		t.Errorf("Error reading %s: %s", outF, err.Error())

	}
	got := string(o)
	want := "\"projectName\":\"foo\""
	if !strings.Contains(got, want) {
		t.Errorf("Want %s to be in %s", want, got)
	}
}
