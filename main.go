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
	"os"

	flag "github.com/getgauge/mflag"
)

var inputFile = flag.String([]string{"-input", "i"}, "", "Source json file to generate report from")
var outDir = flag.String([]string{"-out", "o"}, "", "Output location for generating report. Will create directory if it doesn't exist.")
var theme = flag.String([]string{"-theme", "t"}, "default", "Theme to use for generating html report. 'default' theme will be used if not specified.")

func main() {
	flag.Parse()
	if *inputFile != "" {
		if *outDir == "" {
			flag.PrintDefaults()
			os.Exit(1)
		}
		regenerateReport(*inputFile, *outDir, *theme)
		return
	}

	findPluginAndProjectRoot()
	action := os.Getenv(pluginActionEnv)
	if action == setupAction {
		addDefaultPropertiesToProject()
	} else if action == executionAction {
		createExecutionReport()
	}
}
