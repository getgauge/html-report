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

package theme

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	helper "github.com/getgauge/html-report/test_helper"
)

func init() {
	templateBasePath = filepath.Join("..", "themes")
}

func TestCopyingReportTemplates(t *testing.T) {
	dirToCopy := filepath.Join(os.TempDir(), randomName())
	defer os.RemoveAll(dirToCopy)

	err := CopyReportTemplateFiles(GetThemePath(""), dirToCopy)
	if err != nil {
		t.Errorf("Expected error == nil, got: %s \n", err.Error())
	}
	verifyReportTemplateFilesAreCopied(dirToCopy, t)
}

func verifyReportTemplateFilesAreCopied(dest string, t *testing.T) {
	reportDir := filepath.Join(GetThemePath(""), "assets")
	filepath.Walk(reportDir, func(path string, info os.FileInfo, err error) error {
		path = strings.Replace(path, reportDir, "", 1)
		destFilePath := filepath.Join(dest, path)
		if !helper.FileExists(destFilePath) {
			t.Errorf("File %s not copied.", destFilePath)
		}
		return nil
	})
}

func randomName() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
