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
	. "gopkg.in/check.v1"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

var now = time.Now()

type testNameGenerator struct {
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

func (s *MySuite) TestCreatingReport(c *C) {
	reportDir := filepath.Join(os.TempDir(), randomName())
	defer os.RemoveAll(reportDir)

	finalReportDir, err := createHtmlReport(reportDir, make([]byte, 0), nil)
	c.Assert(err, IsNil)

	expectedFinalReportDir := filepath.Join(reportDir, htmlReport)
	c.Assert(finalReportDir, Equals, expectedFinalReportDir)
	verifyReportTemplateFilesAreCopied(expectedFinalReportDir, c)
}

func (s *MySuite) TestCreatingReportWithNoOverWrite(c *C) {
	reportDir := filepath.Join(os.TempDir(), randomName())
	defer os.RemoveAll(reportDir)

	nameGen := testNameGenerator{}
	finalReportDir, err := createHtmlReport(reportDir, make([]byte, 0), nameGen)
	c.Assert(err, IsNil)

	expectedFinalReportDir := filepath.Join(reportDir, htmlReport, nameGen.randomName())
	c.Assert(finalReportDir, Equals, expectedFinalReportDir)
	verifyReportTemplateFilesAreCopied(expectedFinalReportDir, c)
}

func randomName() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func verifyReportTemplateFilesAreCopied(dest string, c *C) {
	filepath.Walk("report-template", func(path string, info os.FileInfo, err error) error {
		path = strings.Replace(path, "report-template", "", 1)
		c.Assert(fileExists(filepath.Join(dest, path)), Equals, true)
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
