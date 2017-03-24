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
	"strings"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

var now = time.Now()

type testNameGenerator struct {
}

func init() {
	templateBasePath = "themes"
}

func (T testNameGenerator) randomName() string {
	return now.Format(timeFormat)
}

func (s *MySuite) TestCopyingReportTemplates(c *C) {
	dirToCopy := filepath.Join(os.TempDir(), randomName())
	defer os.RemoveAll(dirToCopy)

	err := copyReportTemplateFiles(dirToCopy)
	c.Assert(err, IsNil)
	verifyReportTemplateFilesAreCopied(dirToCopy, c)
}

func (s *MySuite) TestGetReportsDirectory(c *C) {
	userSetReportsDir := filepath.Join(os.TempDir(), randomName())
	os.Setenv(gaugeReportsDirEnvName, userSetReportsDir)
	expectedReportsDir := filepath.Join(userSetReportsDir, htmlReport)
	defer os.RemoveAll(userSetReportsDir)

	reportsDir := getReportsDirectory(nil)

	c.Assert(reportsDir, Equals, expectedReportsDir)
	if !fileExists(expectedReportsDir) {
		c.Errorf("Expected %s report directory doesn't exist", expectedReportsDir)
	}
}

func (s *MySuite) TestGetReportsDirectoryWithOverrideFlag(c *C) {
	userSetReportsDir := filepath.Join(os.TempDir(), randomName())
	os.Setenv(gaugeReportsDirEnvName, userSetReportsDir)
	os.Setenv(overwriteReportsEnvProperty, "true")
	nameGen := &testNameGenerator{}
	expectedReportsDir := filepath.Join(userSetReportsDir, htmlReport, nameGen.randomName())
	defer os.RemoveAll(userSetReportsDir)

	reportsDir := getReportsDirectory(nameGen)

	c.Assert(reportsDir, Equals, expectedReportsDir)
	if !fileExists(expectedReportsDir) {
		c.Errorf("Expected %s report directory doesn't exist", expectedReportsDir)
	}
}

func randomName() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func verifyReportTemplateFilesAreCopied(dest string, c *C) {
	filepath.Walk(reportTemplateDir, func(path string, info os.FileInfo, err error) error {
		path = strings.Replace(path, reportTemplateDir, "", 1)
		destFilePath := filepath.Join(dest, path)
		if !fileExists(destFilePath) {
			c.Errorf("File %s not copied.", destFilePath)
		}
		return nil
	})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

func (s *MySuite) TestCreatingReportShouldOverwriteReportsBasedOnEnv(c *C) {
	os.Setenv(overwriteReportsEnvProperty, "true")
	nameGen := getNameGen()
	c.Assert(nameGen, Equals, nil)

	os.Setenv(overwriteReportsEnvProperty, "false")
	nameGen = getNameGen()
	c.Assert(nameGen, Equals, timeStampedNameGenerator{})
}
