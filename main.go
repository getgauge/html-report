/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/
package main

import (
	"net"
	"os"

	"github.com/getgauge/common"
	"github.com/getgauge/html-report/env"
	"github.com/getgauge/html-report/gauge_messages"
	"github.com/getgauge/html-report/logger"
	"github.com/getgauge/html-report/regenerate"
	flag "github.com/getgauge/mflag"
	"google.golang.org/grpc"
)

var inputFile = flag.String([]string{"-input", "i"}, "", "Source file to generate report from. This should be generated in <PROJECTROOT>/.gauge folder.")
var outDir = flag.String([]string{"-output", "o"}, "", "Output location for generating report. Will create directory if it doesn't exist.")
var themePath = flag.String([]string{"-theme", "t"}, "", "Theme to use for generating html report. 'default' theme will be used if not specified.")

func main() {
	flag.Parse()
	if *inputFile != "" {
		if *outDir == "" {
			flag.PrintDefaults()
			os.Exit(1)
		}
		projectRoot, err := common.GetProjectRoot()
		if err != nil {
			logger.Fatalf("%s", err.Error())
		}
		if !common.FileExists(*inputFile) {
			logger.Fatalf("Input file does not exist: %s", *inputFile)
		}
		regenerate.Report(*inputFile, *outDir, *themePath, projectRoot)
		return
	}

	action := os.Getenv(pluginActionEnv)
	if action == setupAction {
		env.AddDefaultPropertiesToProject()
	} else if action == executionAction {
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
