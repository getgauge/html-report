/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
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
	helper.SetEnvOrFail(t, env.GaugeReportsDirEnvName, userSetReportsDir)
	expectedReportsDir := filepath.Join(userSetReportsDir, htmlReport)
	defer func(path string) {
		if err := os.RemoveAll(path); err != nil {
			t.Errorf("Failed to remove directory %s: %v", path, err)
		}
	}(userSetReportsDir)

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
	helper.SetEnvOrFail(t, env.GaugeReportsDirEnvName, userSetReportsDir)
	helper.SetEnvOrFail(t, env.OverwriteReportsEnvProperty, "true")
	nameGen := &testNameGenerator{}
	expectedReportsDir := filepath.Join(userSetReportsDir, htmlReport, nameGen.randomName())
	defer func(path string) {
		if err := os.RemoveAll(path); err != nil {
			t.Errorf("Failed to remove directory %s: %v", path, err)
		}
	}(userSetReportsDir)

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
	helper.SetEnvOrFail(t, env.OverwriteReportsEnvProperty, "true")
	nameGen := getNameGen()
	if nameGen != nil {
		t.Errorf("Expected nameGen == nil, got %s", nameGen)
	}

	helper.SetEnvOrFail(t, env.OverwriteReportsEnvProperty, "false")
	nameGen = getNameGen()
	switch nameGen.(type) {
	case timeStampedNameGenerator:
	default:
		t.Errorf("Expected nameGen to be type timeStampedNameGenerator, got %s", reflect.TypeOf(nameGen))
	}
}

func TestCreateReportExecutableFileShouldCreateExecFile(t *testing.T) {
	isSaveExecutionResultDisabled = func() bool { return false }
	exPath := filepath.Join(os.TempDir(), "html-report")
	exTargetFileName := "html-report-target"
	if runtime.GOOS == "windows" {
		exTargetFileName = "html-report-target.bat"
	}
	exTarget := filepath.Join(os.TempDir(), exTargetFileName)
	_, err := os.Create(exPath)
	if err != nil {
		t.Errorf("could not create %s. %s", exPath, err.Error())
	}
	defer helper.RemoveOrFail(t, exPath)
	defer helper.RemoveOrFail(t, exTarget)

	createReportExecutableFile(exPath, exTarget)

	if !fileExists(exTarget) {
		t.Errorf("Could not create a symlink of src: %s to  dst: %s", exPath, exTarget)
	}
}
func TestCreateReportExecutableFileShouldNotCreateExecFile(t *testing.T) {
	isSaveExecutionResultDisabled = func() bool { return true }
	exPath := filepath.Join(os.TempDir(), "html-report")
	exTarget := filepath.Join(os.TempDir(), "html-report-target")
	_, err := os.Create(exPath)
	if err != nil {
		t.Errorf("could not create %s. %s", exPath, err.Error())
	}

	defer helper.RemoveOrFail(t, exPath)
	defer helper.RemoveOrFail(t, exTarget)
	defer helper.UnsetEnvOrFail(t, env.SaveExecutionResult)
	createReportExecutableFile(exPath, exTarget)
	if fileExists(exTarget) {
		t.Errorf("Expected not to create a symlink of src: %s to  dst: %s", exPath, exTarget)
	}
}
