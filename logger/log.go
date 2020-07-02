package logger

import (
	"encoding/json"
	"fmt"
	"os"
)

type logMessage struct {
	Level   string `json:"logLevel"`
	Message string `json:"message"`
}

func (message *logMessage) toJSON() (string, error) {
	json, err := json.Marshal(message)
	if err != nil {
		return "", err
	}
	return string(json), nil
}

//Init initialize logger
func Init() {
}

//Debug logs debug message
func Debug(msg string) {
	log("debug", msg)
}

//Debugf logs debug message
func Debugf(format string, args ...interface{}) {
	Debug(fmt.Sprintf(format, args...))
}

//Info logs debug message
func Info(msg string) {
	log("info", msg)
}

//Infof logs info message
func Infof(format string, args ...interface{}) {
	Info(fmt.Sprintf(format, args...))
}

//Fatal logs CRITICAL messages and exits
func Fatal(msg string) {
	log("fatal", msg)
	os.Exit(1)
}

//Fatalf logs CRITICAL messages and exits
func Fatalf(format string, args ...interface{}) {
	Fatal(fmt.Sprintf(format, args...))
}

//Warnf logs warning message
func Warnf(format string, args ...interface{}) {
	log("warning", fmt.Sprintf(format, args...))
}

func log(logLevel string, msg string) {
	message := logMessage{Level: logLevel, Message: msg}
	m, _ := message.toJSON()
	fmt.Fprintln(os.Stdout, m)
}
