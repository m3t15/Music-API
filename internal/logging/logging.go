package logging

import (
	"fmt"
	"strings"
)

type logData struct {
	level     string // level for file (info, warn, error, fatal)
	message   string // log message
	timestamp string // datetime of the log
}

// Super rudimentary logging (in this func I want to add further normalizaiton for some of the variables/add more data idk yet)
func Logger(level string, message string, timestamp string) logData {
	// Normalize data:
	message = strings.TrimSpace(message) // might need to revisit this in the future if error messages end up actually being multi-line
	level = strings.ToTitle(level)
	logInfo := logData{level: level, message: message, timestamp: timestamp}
	logText(logInfo)
	logDB(logInfo)
	return logInfo
}

// TO DO: Finish log to text file functionality (right now just output to terminal)
func logText(logInfo logData) {
	fmt.Printf("Level: %s | Message: %s | Time: %s\n", logInfo.level, logInfo.message, logInfo.timestamp)
}

// TO DO: Finish functionality for logging stuff to the action database
func logDB(logInfo logData) {

}

// TO DO: This is staged for future use in logging any post or patches that trigger a change/upload
func LogPostPatch() {

}
