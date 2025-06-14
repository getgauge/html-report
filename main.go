/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/html-report/env"
	"github.com/getgauge/html-report/logger"
	"github.com/getgauge/html-report/regenerate"
	"google.golang.org/grpc"
)

const usage = `Usage of using_flag:
  -i, --input Source file to generate report from. This should be generated in <PROJECTROOT>/.gauge folder.
  -o, --output Output location for generating report. Will create directory if it doesn't exist.
  -t, --theme Theme to use for generating html report. 'default' theme will be used if not specified.
  -h, --help prints help information 
`

func main() {
	var inputFile string
	flag.StringVar(&inputFile, "input", "", "Source file to generate report from. This should be generated in <PROJECTROOT>/.gauge folder.")
	flag.StringVar(&inputFile, "i", "", "Source file to generate report from. This should be generated in <PROJECTROOT>/.gauge folder.")
	var outDir string
	flag.StringVar(&outDir, "output", "", "Output location for generating report. Will create directory if it doesn't exist.")
	flag.StringVar(&outDir, "o", "", "Output location for generating report. Will create directory if it doesn't exist.")
	var themePath string
	flag.StringVar(&themePath, "theme", "", "Theme to use for generating html report. 'default' theme will be used if not specified.")
	flag.StringVar(&themePath, "t", "", "Theme to use for generating html report. 'default' theme will be used if not specified.")

	flag.Usage = func() { fmt.Print(usage) }
	flag.Parse()
	if inputFile != "" {
		if outDir == "" {
			flag.PrintDefaults()
			os.Exit(1)
		}
		projectRoot, err := common.GetProjectRoot()
		if err != nil {
			logger.Fatalf("%s", err.Error())
		}
		if !common.FileExists(inputFile) {
			logger.Fatalf("Input file does not exist: %s", inputFile)
		}
		regenerate.Report(inputFile, outDir, themePath, projectRoot)
		return
	}

	switch action := os.Getenv(pluginActionEnv); action {
	case setupAction:
		env.AddDefaultPropertiesToProject()
	case executionAction:
		pluginsDir, _ = os.Getwd()
		err := os.Chdir(env.GetProjectRoot())
		if err != nil {
			logger.Fatalf("failed to chdir to %s. %s", pluginsDir, err.Error())
		}

		address, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
		if err != nil {
			logger.Fatalf("failed to start server.")
		}
		l, err := net.ListenTCP("tcp", address)
		if err != nil {
			logger.Fatalf("failed to start server.")
		}
		mSize := env.GetMaxMessageSize()
		logger.Debugf("Setting MaxRecvMsgSize = %d MB", mSize)
		server := grpc.NewServer(grpc.MaxRecvMsgSize(mSize * 1024 * 1024))
		h := &handler{server: server}
		gauge_messages.RegisterReporterServer(server, h)
		logger.Infof("Listening on port:%d", l.Addr().(*net.TCPAddr).Port)
		err = server.Serve(l)
		if err != nil {
			logger.Fatalf("failed to start server. %s", err.Error())
		}
	}
}
