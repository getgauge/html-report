package logger

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var old = os.Stdout
var fname = filepath.Join(os.TempDir(), "stdout")

func TestDebugShoudWiteLogInJsonFormat(t *testing.T) {
	temp := setStdout()
	defer temp.Close()

	Debug("log debug message")
	got, _ := ioutil.ReadFile(fname)
	want := "{\"logLevel\":\"debug\",\"message\":\"log debug message\"}\n"

	if want != string(got) {
		t.Errorf("Expected %s to be => %s", want, got)
	}
	os.Stdout = old
}

func TestDebugfShoudWiteLogInJsonFormat(t *testing.T) {
	temp := setStdout()
	defer temp.Close()

	Debugf("log %s debug message", "formatted")
	got, _ := ioutil.ReadFile(fname)
	want := "{\"logLevel\":\"debug\",\"message\":\"log formatted debug message\"}\n"

	if want != string(got) {
		t.Errorf("Expected %s to be => %s", want, got)
	}
	os.Stdout = old
}

func TestInfoShoudWiteLogInJsonFormat(t *testing.T) {
	temp := setStdout()
	defer temp.Close()

	Info("log info message")
	got, _ := ioutil.ReadFile(fname)
	want := "{\"logLevel\":\"info\",\"message\":\"log info message\"}\n"

	if want != string(got) {
		t.Errorf("Expected %s to be => %s", want, got)
	}
	os.Stdout = old
}

func TestInfofShoudWiteLogInJsonFormat(t *testing.T) {
	temp := setStdout()
	defer temp.Close()

	Infof("log %s info message", "formatted")
	got, _ := ioutil.ReadFile(fname)
	want := "{\"logLevel\":\"info\",\"message\":\"log formatted info message\"}\n"

	if want != string(got) {
		t.Errorf("Expected %s to be => %s", want, got)
	}
	os.Stdout = old
}

func TestWarnfShoudWiteLogInJsonFormat(t *testing.T) {
	temp := setStdout()
	defer temp.Close()

	Warnf("log %s warning message", "formatted")
	got, _ := ioutil.ReadFile(fname)
	want := "{\"logLevel\":\"warning\",\"message\":\"log formatted warning message\"}\n"

	if want != string(got) {
		t.Errorf("Expected %s to be => %s", want, got)
	}
	os.Stdout = old
}

func setStdout() *os.File {
	temp, _ := os.Create(fname)
	os.Stdout = temp
	return temp
}
