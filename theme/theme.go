/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package theme

import (
	"os"
	"path/filepath"

	"github.com/getgauge/common"
)

const (
	reportThemeProperty = "GAUGE_HTML_REPORT_THEME_PATH"
)

var templateBasePath string

func GetDefaultThemePath(pluginsDir string) string {
	if templateBasePath == "" {
		templateBasePath = filepath.Join(pluginsDir, "themes")
	}
	return filepath.Join(templateBasePath, "default")
}

func CopyReportTemplateFiles(themePath, reportDir string) error {
	r := filepath.Join(themePath, "assets")
	_, err := common.MirrorDir(r, reportDir)
	return err
}

func GetThemePath(pluginsDir string) string {
	t := os.Getenv(reportThemeProperty)
	if t == "" {
		t = GetDefaultThemePath(pluginsDir)
	}
	return t
}
