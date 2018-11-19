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

package regenerate

import (
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/getgauge/html-report/env"
	"github.com/getgauge/html-report/gauge_messages"
	"github.com/getgauge/html-report/generator"
	"github.com/getgauge/html-report/theme"
	"github.com/golang/protobuf/proto"
)

// Report generates html report from saved result.
func Report(inputFile, reportsDir, themePath, pRoot string) {
	b, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Fatal(err.Error())
	}
	psr := &gauge_messages.ProtoSuiteResult{}
	err = proto.Unmarshal(b, psr)
	if err != nil {
		log.Fatalf("Unable to read last run data from %s. Error: %s", inputFile, err.Error())
	}
	res := generator.ToSuiteResult(pRoot, psr)

	env.CreateDirectory(reportsDir)
	if themePath == "" {
		workingDir, _ := env.GetCurrentExecutableDir()
		themePath = theme.GetDefaultThemePath(filepath.Dir(workingDir))
	}
	generator.GenerateReport(res, reportsDir, themePath, true)
}
