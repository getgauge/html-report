package logger

import (
	"log"
	"os"
	"strings"
)

var debug bool

func Init() {
	var env = os.Getenv("GAUGE_LOG_LEVEL")
	if strings.ToLower(env) == "debug" {
		debug = true
	}
	log.SetPrefix("[html-report]")
}

func Debug(format string, args ...string) {
	if debug {
		log.Printf(format, args)
	}
}
