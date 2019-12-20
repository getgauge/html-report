// Copyright 2019 ThoughtWorks, Inc.

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
	"context"
	"os"

	"github.com/getgauge/html-report/gauge_messages"
	"google.golang.org/grpc"
)

type handler struct {
	server *grpc.Server
}

func (h *handler) NotifyExecutionStarting(c context.Context, m *gauge_messages.ExecutionStartingRequest) (*gauge_messages.Empty, error) {
	return &gauge_messages.Empty{}, nil
}
func (h *handler) NotifySpecExecutionStarting(c context.Context, m *gauge_messages.SpecExecutionStartingRequest) (*gauge_messages.Empty, error) {
	return &gauge_messages.Empty{}, nil
}
func (h *handler) NotifyScenarioExecutionStarting(c context.Context, m *gauge_messages.ScenarioExecutionStartingRequest) (*gauge_messages.Empty, error) {
	return &gauge_messages.Empty{}, nil
}
func (h *handler) NotifyStepExecutionStarting(c context.Context, m *gauge_messages.StepExecutionStartingRequest) (*gauge_messages.Empty, error) {
	return &gauge_messages.Empty{}, nil
}
func (h *handler) NotifyStepExecutionEnding(c context.Context, m *gauge_messages.StepExecutionEndingRequest) (*gauge_messages.Empty, error) {
	return &gauge_messages.Empty{}, nil
}
func (h *handler) NotifyScenarioExecutionEnding(c context.Context, m *gauge_messages.ScenarioExecutionEndingRequest) (*gauge_messages.Empty, error) {
	return &gauge_messages.Empty{}, nil
}
func (h *handler) NotifySpecExecutionEnding(c context.Context, m *gauge_messages.SpecExecutionEndingRequest) (*gauge_messages.Empty, error) {
	return &gauge_messages.Empty{}, nil
}
func (h *handler) NotifyExecutionEnding(c context.Context, m *gauge_messages.ExecutionEndingRequest) (*gauge_messages.Empty, error) {
	return &gauge_messages.Empty{}, nil
}

func (h *handler) NotifySuiteResult(c context.Context, m *gauge_messages.SuiteExecutionResult) (*gauge_messages.Empty, error) {
	createReport(m, true)
	return &gauge_messages.Empty{}, nil
}

func (h *handler) Kill(c context.Context, m *gauge_messages.KillProcessRequest) (*gauge_messages.Empty, error) {
	defer h.stopServer()
	return &gauge_messages.Empty{}, nil
}

func (h *handler) stopServer() {
	h.server.Stop()
	os.Exit(0)
}
