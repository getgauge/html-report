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
	"strings"
	"time"

	"encoding/json"

	"github.com/getgauge/common"
	"github.com/getgauge/html-report/gauge_messages"
	"github.com/getgauge/html-report/generator"
	"github.com/getgauge/html-report/listener"
)

const (
	defaultReportsDir           = "reports"
	gaugeReportsDirEnvName      = "gauge_reports_dir" // directory where reports are generated by plugins
	overwriteReportsEnvProperty = "overwrite_reports"
	resultJsFile                = "result.js"
	htmlReport                  = "html-report"
	setupAction                 = "setup"
	executionAction             = "execution"
	gaugeHost                   = "localhost"
	gaugePortEnv                = "plugin_connection_port"
	pluginActionEnv             = "html-report_action"
	timeFormat                  = "2006-01-02 15.04.05"
	defaultTheme                = "default"
	reportThemeProperty         = "GAUGE_HTML_REPORT_THEME_PATH"
	resultFile                  = "last_run_result.json"
)

var projectRoot string
var pluginDir string

type nameGenerator interface {
	randomName() string
}

type timeStampedNameGenerator struct {
}

func (T timeStampedNameGenerator) randomName() string {
	return time.Now().Format(timeFormat)
}

func findPluginAndProjectRoot() {
	projectRoot = os.Getenv(common.GaugeProjectRootEnv)
	if projectRoot == "" {
		fmt.Printf("Environment variable '%s' is not set. \n", common.GaugeProjectRootEnv)
		os.Exit(1)
	}

	var err error
	pluginDir, err = os.Getwd()
	if err != nil {
		fmt.Printf("Error finding current working directory: %s \n", err)
		os.Exit(1)
	}
}

func createExecutionReport() {
	os.Chdir(projectRoot)
	listener, err := listener.NewGaugeListener(gaugeHost, os.Getenv(gaugePortEnv))
	if err != nil {
		fmt.Println("Could not create the gauge listener")
		os.Exit(1)
	}
	listener.OnSuiteResult(createReport)
	listener.Start()
}

func addDefaultPropertiesToProject() {
	defaultPropertiesFile := getDefaultPropertiesFile()

	reportsDirProperty := &(common.Property{
		Comment:      "The path to the gauge reports directory. Should be either relative to the project directory or an absolute path",
		Name:         gaugeReportsDirEnvName,
		DefaultValue: defaultReportsDir})

	overwriteReportProperty := &(common.Property{
		Comment:      "Set as false if gauge reports should not be overwritten on each execution. A new time-stamped directory will be created on each execution.",
		Name:         overwriteReportsEnvProperty,
		DefaultValue: "true"})

	if !common.FileExists(defaultPropertiesFile) {
		fmt.Printf("Failed to setup html report plugin in project. Default properties file does not exist at %s. \n", defaultPropertiesFile)
		return
	}
	if err := common.AppendProperties(defaultPropertiesFile, reportsDirProperty, overwriteReportProperty); err != nil {
		fmt.Printf("Failed to setup html report plugin in project: %s \n", err)
		return
	}
	fmt.Println("Succesfully added configurations for html-report to env/default/default.properties")
}

func getDefaultPropertiesFile() string {
	return filepath.Join(projectRoot, "env", "default", "default.properties")
}

func createReport(suiteResult *gauge_messages.SuiteExecutionResult) {
	projectRoot, err := common.GetProjectRoot()
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	reportsDir := getReportsDirectory(getNameGen())
	generator.ProjectRoot = projectRoot
	res := generator.ToSuiteResult(suiteResult.GetSuiteResult())
	go saveLastExecutionResult(res, reportsDir)
	generator.GenerateReport(res, reportsDir, getThemePath())
}

func getNameGen() nameGenerator {
	var nameGen nameGenerator
	if shouldOverwriteReports() {
		nameGen = nil
	} else {
		nameGen = timeStampedNameGenerator{}
	}
	return nameGen
}

func getReportsDirectory(nameGen nameGenerator) string {
	reportsDir, err := filepath.Abs(os.Getenv(gaugeReportsDirEnvName))
	if reportsDir == "" || err != nil {
		reportsDir = defaultReportsDir
	}
	generator.CreateDirectory(reportsDir)
	var currentReportDir string
	if nameGen != nil {
		currentReportDir = filepath.Join(reportsDir, htmlReport, nameGen.randomName())
	} else {
		currentReportDir = filepath.Join(reportsDir, htmlReport)
	}
	generator.CreateDirectory(currentReportDir)
	return currentReportDir
}

func getThemePath() string {
	t := os.Getenv(reportThemeProperty)
	if t == "" {
		t = generator.GetDefaultThemePath()
	}
	return t
}

func shouldOverwriteReports() bool {
	envValue := os.Getenv(overwriteReportsEnvProperty)
	if strings.ToLower(envValue) == "true" {
		return true
	}
	return false
}

func saveLastExecutionResult(r *generator.SuiteResult, reportsDir string) {
	o, err := json.Marshal(r)
	if err != nil {
		log.Printf("[Warning] Error saving Last Execution Run: %s\n", err.Error())
		return
	}
	outF := filepath.Join(reportsDir, resultFile)
	err = ioutil.WriteFile(outF, o, common.NewFilePermissions)
	if err != nil {
		log.Printf("[Warning] Failed to write to %s. Reason: %s\n", outF, err.Error())
		return
	}
	fmt.Printf("Result from current execution has been saved to %s\n", outF)
	dir, bName := generator.GetCurrentExecutableDir()
	exPath := filepath.Join(dir, bName)
	exTarget := filepath.Join(reportsDir, bName)
	if fileExists(exTarget) {
		os.Remove(exTarget)
	}
	err = os.Symlink(exPath, exTarget)
	if err != nil {
		log.Printf("[Warning] Unable to create symlink %s\n", exTarget)
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}
