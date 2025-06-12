package logger

import (
	"os"
	"path/filepath"
	"testing"
)

var old = os.Stdout
var fname = filepath.Join(os.TempDir(), "stdout")

func runLogTest(t *testing.T, logFunc func(), expected string) {
	temp := setStdout()
	defer func() {
		if err := temp.Close(); err != nil {
			t.Errorf("failed to close temp file: %v", err)
		}
		os.Stdout = old
	}()

	logFunc()
	got, _ := os.ReadFile(fname)
	if expected != string(got) {
		t.Errorf("Expected %s to be => %s", expected, got)
	}
}

func TestDebugShouldWriteLogInJsonFormat(t *testing.T) {
	runLogTest(t, func() { Debug("log debug message") }, "{\"logLevel\":\"debug\",\"message\":\"log debug message\"}\n")
}

func TestDebugfShouldWriteLogInJsonFormat(t *testing.T) {
	runLogTest(t, func() { Debugf("log %s debug message", "formatted") }, "{\"logLevel\":\"debug\",\"message\":\"log formatted debug message\"}\n")
}

func TestInfoShouldWriteLogInJsonFormat(t *testing.T) {
	runLogTest(t, func() { Info("log info message") }, "{\"logLevel\":\"info\",\"message\":\"log info message\"}\n")
}

func TestInfofShouldWriteLogInJsonFormat(t *testing.T) {
	runLogTest(t, func() { Infof("log %s info message", "formatted") }, "{\"logLevel\":\"info\",\"message\":\"log formatted info message\"}\n")
}

func TestWarnfShouldWriteLogInJsonFormat(t *testing.T) {
	runLogTest(t, func() { Warnf("log %s warning message", "formatted") }, "{\"logLevel\":\"warning\",\"message\":\"log formatted warning message\"}\n")
}

func setStdout() *os.File {
	temp, _ := os.Create(fname)
	os.Stdout = temp
	return temp
}
