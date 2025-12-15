package main

import (
	"io"
	"log"
	"os"
)

var (
	verboseEnabled bool
	verboseLogger  = log.New(io.Discard, "beet [verbose] ", log.LstdFlags)
)

func configureVerboseLogging(enabled bool) {
	verboseEnabled = enabled
	if enabled {
		verboseLogger.SetOutput(os.Stderr)
	} else {
		verboseLogger.SetOutput(io.Discard)
	}
}

func logVerbose(format string, args ...interface{}) {
	if !verboseEnabled {
		return
	}
	verboseLogger.Printf(format, args...)
}
