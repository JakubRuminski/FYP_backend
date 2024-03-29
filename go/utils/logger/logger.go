package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	BOLD_RED    = "\033[1;31m"       // error
	BOLD_ORANGE = "\033[1;38;5;208m" // warn
	BOLD_GREEN  = "\033[1;32m"       // info
	LIGHT_BLUE  = "\033[36m"         // debug


	RESET = "\033[0m"
)

type Logger struct {
	Environment string
	ClientID    string
}


func (l *Logger) InitRequestLogFile(clientID string) *os.File {
    filePath := fmt.Sprintf("/logs/%s.txt", clientID)

    file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
    if err != nil {
        l.ERROR("Error opening or creating log file: %s", err)
        return nil
    }

    // Create a multi-writer that writes to both the file and standard output
    multiWriter := io.MultiWriter(file, os.Stdout)

    // Set the output to the multi-writer
    log.SetOutput(multiWriter)

    return file
}


func (l *Logger) SetEnvironment(environment string) {
	l.Environment = environment
}


func (l *Logger) INFO(message string, args ...interface{}) {
    _, file, line, _ := runtime.Caller(1)
    padding := calculatePadding(file, line)
    logMessage := fmt.Sprintf("%s[%s:%d] %s [INFO] %s %s\n", BOLD_GREEN, file, line, padding, message, RESET)
    log.Printf(logMessage, args...)
}

func (l *Logger) DEBUG(message string, args ...interface{}) {
    if l.Environment != "PRODUCTION" {
        _, file, line, _ := runtime.Caller(1)
        padding := calculatePadding(file, line)
        logMessage := fmt.Sprintf("%s[%s:%d] %s [DEBUG] %s %s\n", LIGHT_BLUE, file, line, padding, message, RESET)
        log.Printf(logMessage, args...)
    }
}

func (l *Logger) DEBUG_WARN(message string, args ...interface{}) {
	if l.Environment != "PRODUCTION" {
		_, file, line, _ := runtime.Caller(1)
		padding := calculatePadding(file, line)
		logMessage := fmt.Sprintf("%s[%s:%d] %s [DEBUG_WARN] %s %s\n", BOLD_ORANGE, file, line, padding, message, RESET)
		log.Printf(logMessage, args...)
	}
}

func (l *Logger) WARN(message string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	padding := calculatePadding(file, line)
	logMessage := fmt.Sprintf("%s[%s:%d] %s [WARN] %s %s\n", BOLD_ORANGE, file, line, padding, message, RESET)
	log.Printf(logMessage, args...)
}

func (l *Logger) ERROR(message string, args ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	padding := calculatePadding(file, line)
	logMessage := fmt.Sprintf("%s[%s:%d] %s [ERROR] %s %s\n", BOLD_RED, file, line, padding, message, RESET)
	log.Printf(logMessage, args...)
}

func calculatePadding(file string, line int) string {
	paddingLength :=  75 - len(fmt.Sprintf("[%s:%d]", file, line))
	if paddingLength < 0 {
		paddingLength = 0
	}
	return strings.Repeat("_", paddingLength)
}

func ( Logger ) STARTTIME( ) time.Time {
	return time.Now()
}
func ( *Logger ) ENDTIME( startTime time.Time, formatString string, v ...interface{} ) {
	elapsed := time.Since(startTime).Seconds()
	elapsedTimeString := fmt.Sprintf("Time elapsed: %f", elapsed)

	if elapsed <= 0.5 {
		return
	}

	if elapsed > 10.0 {
		formatString = fmt.Sprintf("%s[DEBUGWARNING] %s COMPLETED This took more than 10 seconds. %s%s\n", BOLD_ORANGE, formatString, elapsedTimeString, RESET)

	} else if elapsed > 0.5 {
		formatString = fmt.Sprintf("%s[DEBUGWARNING] %s COMPLETED This took more than 1/2 a second. %s%s\n", BOLD_ORANGE, formatString, elapsedTimeString, RESET)
	
	}
    log.Printf( formatString, v... )
}