/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
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
	err := filepath.Walk(reportDir, func(path string, info os.FileInfo, err error) error {
		path = strings.Replace(path, reportDir, "", 1)
		destFilePath := filepath.Join(dest, path)
		if !helper.FileExists(destFilePath) {
			t.Errorf("File %s not copied.", destFilePath)
		}
		return nil
	})
	if err != nil {
		t.Errorf("unable to walk %s. %s", reportDir, err.Error())
	}
}

func randomName() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
