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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"strings"

	"runtime"

	"github.com/getgauge/common"
	"github.com/getgauge/html-report/env"
	"github.com/getgauge/html-report/gauge_messages"
	"github.com/getgauge/html-report/generator"
	"github.com/getgauge/html-report/listener"
	"github.com/getgauge/html-report/logger"
	"github.com/getgauge/html-report/theme"
)

const (
	resultJsFile    = "result.js"
	htmlReport      = "html-report"
	setupAction     = "setup"
	executionAction = "execution"
	gaugeHost       = "127.0.0.1"
	gaugePortEnv    = "plugin_connection_port"
	pluginActionEnv = "html-report_action"
	timeFormat      = "2006-01-02 15.04.05"
	defaultTheme    = "default"
	resultFile      = "last_run_result.json"
)

type nameGenerator interface {
	randomName() string
}

type timeStampedNameGenerator struct {
}

func (T timeStampedNameGenerator) randomName() string {
	return time.Now().Format(timeFormat)
}

var pluginsDir string

func createExecutionReport() {
	pluginsDir, _ = os.Getwd()
	os.Chdir(env.GetProjectRoot())
	listener, err := listener.NewGaugeListener(gaugeHost, os.Getenv(gaugePortEnv))
	if err != nil {
		fmt.Println("Could not create the gauge listener")
		os.Exit(1)
	}
	listener.OnSuiteResult(createReport)
	listener.Start()
}

func createReport(suiteResult *gauge_messages.SuiteExecutionResult) {
	projectRoot, err := common.GetProjectRoot()
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	reportsDir := getReportsDirectory(getNameGen())
	res := generator.ToSuiteResult(projectRoot, suiteResult.GetSuiteResult())
	logger.Debug("Transformed SuiteResult to report structure")
	go createReportExecutableFile(reportsDir, pluginsDir)
	t := theme.GetThemePath(pluginsDir)
	generator.GenerateReport(res, reportsDir, t)
	logger.Debug("Done generating HTML report using theme from %s", t)
}

func getNameGen() nameGenerator {
	var nameGen nameGenerator
	if env.ShouldOverwriteReports() {
		nameGen = nil
	} else {
		nameGen = timeStampedNameGenerator{}
	}
	return nameGen
}

func getReportsDirectory(nameGen nameGenerator) string {
	reportsDir, err := filepath.Abs(os.Getenv(env.GaugeReportsDirEnvName))
	if reportsDir == "" || err != nil {
		reportsDir = env.DefaultReportsDir
	}
	env.CreateDirectory(reportsDir)
	var currentReportDir string
	if nameGen != nil {
		currentReportDir = filepath.Join(reportsDir, htmlReport, nameGen.randomName())
	} else {
		currentReportDir = filepath.Join(reportsDir, htmlReport)
	}
	env.CreateDirectory(currentReportDir)
	return currentReportDir
}

func createReportExecutableFile(reportsDir, pluginsDir string) {
	_, bName := env.GetCurrentExecutableDir()
	exPath := filepath.Join(pluginsDir, "bin", bName)
	exTarget := filepath.Join(reportsDir, bName)
	if fileExists(exTarget) {
		os.Remove(exTarget)
	}
	if runtime.GOOS == "windows" {
		createBatFileToExecuteHtmlReport(exPath, exTarget)
	} else {
		createSymlinkToHtmlReport(exPath, exTarget)
	}
}

func createBatFileToExecuteHtmlReport(exPath, exTarget string) {
	content := "@echo off \n" + exPath + " %*"
	o := []byte(content)
	exTarget = strings.TrimSuffix(exTarget, filepath.Ext(exTarget))
	outF := exTarget + ".bat"
	err := ioutil.WriteFile(outF, o, common.NewFilePermissions)
	if err != nil {
		log.Printf("[Warning] Failed to write to %s. Reason: %s\n", outF, err.Error())
		return
	}
	logger.Debug("Generated %s", outF)
}

func createSymlinkToHtmlReport(exPath, exTarget string) {
	if _, err := os.Lstat(exTarget); err == nil {
		os.Remove(exTarget)
	}
	err := os.Symlink(exPath, exTarget)
	if err != nil {
		log.Printf("[Warning] Unable to create symlink %s\n", exTarget)
	}
	logger.Debug("Generated symlink %s", exTarget)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}
