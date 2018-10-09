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
	"testing"
	"time"

	"github.com/getgauge/html-report/env"
	helper "github.com/getgauge/html-report/test_helper"
)

var now = time.Now()

type testNameGenerator struct {
}

func (T testNameGenerator) randomName() string {
	return now.Format(timeFormat)
}

func TestGetReportsDirectory(t *testing.T) {
	userSetReportsDir := filepath.Join(os.TempDir(), randomName())
	os.Setenv(env.GaugeReportsDirEnvName, userSetReportsDir)
	expectedReportsDir := filepath.Join(userSetReportsDir, htmlReport)
	defer os.RemoveAll(userSetReportsDir)

	reportsDir := getReportsDirectory(nil)

	if reportsDir != expectedReportsDir {
		t.Errorf("Expected reportsDir == %s, got: %s\n", expectedReportsDir, reportsDir)
	}
	if !helper.FileExists(expectedReportsDir) {
		t.Errorf("Expected %s report directory doesn't exist", expectedReportsDir)
	}
}

func TestGetReportsDirectoryWithOverrideFlag(t *testing.T) {
	userSetReportsDir := filepath.Join(os.TempDir(), randomName())
	os.Setenv(env.GaugeReportsDirEnvName, userSetReportsDir)
	os.Setenv(env.OverwriteReportsEnvProperty, "true")
	nameGen := &testNameGenerator{}
	expectedReportsDir := filepath.Join(userSetReportsDir, htmlReport, nameGen.randomName())
	defer os.RemoveAll(userSetReportsDir)

	reportsDir := getReportsDirectory(nameGen)

	if reportsDir != expectedReportsDir {
		t.Errorf("Expected reportsDir == %s, got: %s\n", expectedReportsDir, reportsDir)
	}
	if !helper.FileExists(expectedReportsDir) {
		t.Errorf("Expected %s report directory doesn't exist", expectedReportsDir)
	}
}

func randomName() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func TestCreatingReportShouldOverwriteReportsBasedOnEnv(t *testing.T) {
	os.Setenv(env.OverwriteReportsEnvProperty, "true")
	nameGen := getNameGen()
	if nameGen != nil {
		t.Errorf("Expected nameGen == nil, got %s", nameGen)
	}

	os.Setenv(env.OverwriteReportsEnvProperty, "false")
	nameGen = getNameGen()
	switch nameGen.(type) {
	case timeStampedNameGenerator:
	default:
		t.Errorf("Expected nameGen to be type timeStampedNameGenerator, got %s", reflect.TypeOf(nameGen))
	}
}

func TestCreateSymlinkToHTMLReportShouldCreateSymlink(t *testing.T) {
	exPath := filepath.Join(os.TempDir(), "html-report")
	exTarget := filepath.Join(os.TempDir(), "html-report-target")
	os.Create(exPath)
	defer os.Remove(exPath)
	defer os.Remove(exTarget)
	createSymlinkToHTMLReport(exPath, exTarget)
	if !fileExists(exTarget) {
		t.Errorf("Could not create a symlink of src: %s to  dst: %s", exPath, exTarget)
	}
}
func TestCreateSymlinkToHTMLReportShouldNotCreateSymlink(t *testing.T) {
	os.Setenv(env.SaveExecutionResult, "false")
	exPath := filepath.Join(os.TempDir(), "html-report")
	exTarget := filepath.Join(os.TempDir(), "html-report-target")
	os.Create(exPath)
	defer os.Remove(exPath)
	defer os.Remove(exTarget)
	defer os.Unsetenv(env.SaveExecutionResult)
	createSymlinkToHTMLReport(exPath, exTarget)
	if fileExists(exTarget) {
		t.Errorf("Expected not to create a symlink of src: %s to  dst: %s", exPath, exTarget)
	}
}

func TestCreateBatFileToExecuteHTMLReportShouldCreateBatFile(t *testing.T) {
	exPath := filepath.Join(os.TempDir(), "html-report")
	exTarget := filepath.Join(os.TempDir(), "html-report-target.bat")
	os.Create(exPath)
	defer os.Remove(exPath)
	defer os.Remove(exTarget)
	createBatFileToExecuteHTMLReport(exPath, exTarget)
	if !fileExists(exTarget) {
		t.Errorf("Could not create file: %s", exTarget)
	}
}
func TestCreateBatFileToExecuteHTMLReportShouldNotCreateBatFile(t *testing.T) {
	os.Setenv(env.SaveExecutionResult, "false")
	exPath := filepath.Join(os.TempDir(), "html-report")
	exTarget := filepath.Join(os.TempDir(), "html-report-target.bat")
	os.Create(exPath)
	defer os.Remove(exPath)
	defer os.Remove(exTarget)
	defer os.Unsetenv(env.SaveExecutionResult)
	createBatFileToExecuteHTMLReport(exPath, exTarget)
	if fileExists(exTarget) {
		t.Errorf("Expected not to create  file : %s", exTarget)
	}
}
