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
