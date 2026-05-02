/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package regenerate

import (
	"os"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/html-report/env"
	"github.com/getgauge/html-report/logger"
	"github.com/getgauge/html-report/mdgen"
	"google.golang.org/protobuf/proto"
)

// Report regenerates a Markdown report from a previously persisted
// last_run_result.bin (proto-serialized SuiteResult).
func Report(inputFile, reportsDir, pRoot string) {
	b, err := os.ReadFile(inputFile)
	if err != nil {
		logger.Fatal(err.Error())
	}
	psr := &gauge_messages.ProtoSuiteResult{}
	if err := proto.Unmarshal(b, psr); err != nil {
		logger.Fatalf("Unable to read last run data from %s. Error: %s", inputFile, err.Error())
	}
	res := mdgen.ToSuiteResult(pRoot, psr)
	env.CreateDirectory(reportsDir)
	if err := mdgen.GenerateReports(res, reportsDir); err != nil {
		logger.Fatalf("Failed to regenerate report: %s", err.Error())
	}
}
