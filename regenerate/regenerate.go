/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package regenerate

import (
	"github.com/getgauge/html-report/logger"
	"io/ioutil"
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
		logger.Fatal(err.Error())
	}
	psr := &gauge_messages.ProtoSuiteResult{}
	err = proto.Unmarshal(b, psr)
	if err != nil {
		logger.Fatalf("Unable to read last run data from %s. Error: %s", inputFile, err.Error())
	}
	res := generator.ToSuiteResult(pRoot, psr)

	env.CreateDirectory(reportsDir)
	if themePath == "" {
		workingDir, _ := env.GetCurrentExecutableDir()
		themePath = theme.GetDefaultThemePath(filepath.Dir(workingDir))
	}
	generator.GenerateReport(res, reportsDir, themePath, true)
}
