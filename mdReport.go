/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/html-report/env"
	"github.com/getgauge/html-report/logger"
	"github.com/getgauge/html-report/mdgen"
)

const (
	markdownReport  = "markdown-report"
	setupAction     = "setup"
	executionAction = "execution"
	pluginActionEnv = "markdown-report_action"
	timeFormat      = "2006-01-02_15.04.05"
)

type nameGenerator interface {
	randomName() string
}

type timeStampedNameGenerator struct{}

func (T timeStampedNameGenerator) randomName() string {
	return time.Now().Format(timeFormat)
}

var pluginsDir string

// createReport is invoked by the gRPC handler when a suite finishes. It
// transforms the proto result and writes a Markdown report tree.
func createReport(suiteResult *gauge_messages.SuiteExecutionResult) {
	projectRoot, err := common.GetProjectRoot()
	if err != nil {
		logger.Debugf("Failed to generate report. %s", err.Error())
		return
	}
	reportsDir := getReportsDirectory(getNameGen())
	res := mdgen.ToSuiteResult(projectRoot, suiteResult.GetSuiteResult())
	logger.Debug("Transformed SuiteResult to report structure")
	go createReportExecutableFile(getExecutableAndTargetPath(reportsDir, pluginsDir))
	if err := mdgen.GenerateReports(res, reportsDir); err != nil {
		logger.Fatalf("Failed to generate reports: %s\n", err.Error())
	}
	logger.Debugf("Done generating Markdown report at %s", reportsDir)
}

func getNameGen() nameGenerator {
	if env.ShouldOverwriteReports() {
		return nil
	}
	return timeStampedNameGenerator{}
}

func getReportsDirectory(nameGen nameGenerator) string {
	reportsDir, err := filepath.Abs(os.Getenv(env.GaugeReportsDirEnvName))
	if reportsDir == "" || err != nil {
		reportsDir = env.DefaultReportsDir
	}
	env.CreateDirectory(reportsDir)
	var currentReportDir string
	if nameGen != nil {
		currentReportDir = filepath.Join(reportsDir, markdownReport, nameGen.randomName())
	} else {
		currentReportDir = filepath.Join(reportsDir, markdownReport)
	}
	env.CreateDirectory(currentReportDir)
	return currentReportDir
}

func getExecutableAndTargetPath(reportsDir string, pluginsDir string) (exPath string, exTarget string) {
	_, bName := env.GetCurrentExecutableDir()
	exPath = filepath.Join(pluginsDir, "bin", bName)
	exTarget = filepath.Join(reportsDir, bName)
	return
}

func createReportExecutableFile(exPath, exTarget string) {
	if isSaveExecutionResultDisabled() {
		return
	}
	if fileExists(exTarget) {
		if err := os.Remove(exTarget); err != nil {
			logger.Debugf("[Warning] Unable to remove existing file %s. Reason: %s\n", exTarget, err.Error())
			return
		}
	}
	if runtime.GOOS == "windows" {
		createBatFileToExecuteReport(exPath, exTarget)
	} else {
		createSymlinkToReport(exPath, exTarget)
	}
}

func createBatFileToExecuteReport(exPath, exTarget string) {
	content := "@echo off \n" + exPath + " %*"
	o := []byte(content)
	exTarget = strings.TrimSuffix(exTarget, filepath.Ext(exTarget))
	outF := exTarget + ".bat"
	err := os.WriteFile(outF, o, common.NewFilePermissions)
	if err != nil {
		logger.Debugf("[Warning] Failed to write to %s. Reason: %s\n", outF, err.Error())
		return
	}
	logger.Debugf("Generated %s", outF)
}

func createSymlinkToReport(exPath, exTarget string) {
	if _, err := os.Lstat(exTarget); err == nil {
		if err := os.Remove(exTarget); err != nil {
			logger.Debugf("[Warning] Unable to remove existing symlink %s\n", exTarget)
			return
		}
	}
	if err := os.Symlink(exPath, exTarget); err != nil {
		logger.Debugf("[Warning] Unable to create symlink %s\n", exTarget)
	}
	logger.Debugf("Generated symlink %s", exTarget)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

var isSaveExecutionResultDisabled = func() bool {
	return os.Getenv(env.SaveExecutionResult) == "false"
}
