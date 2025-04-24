package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// sanitizeOutput removes excessive newlines and trims spaces.
func sanitizeOutput(output string) string {
	output = strings.TrimSpace(output)
	output = strings.ReplaceAll(output, "\r\n", "\n")
	output = strings.ReplaceAll(output, "\r", "\n")
	output = strings.ReplaceAll(output, "\n\n", "\n")
	return output
}

// logPrintf writes formatted logs to a file and optionally to the console.
func logPrintf(logFile *os.File, cliDebug bool, format string, args ...interface{}) {
	logLine := fmt.Sprintf(format, args...)
	logFile.WriteString(logLine + "\n")
	if cliDebug {
		fmt.Printf(format, args...)
	}
}

// logPrintln writes logs with a newline to a file and optionally to the console.
func logPrintln(logFile *os.File, cliDebug bool, args ...interface{}) {
	logLine := fmt.Sprintln(args...)
	logFile.WriteString(logLine)
	if cliDebug {
		fmt.Println(args...)
	}
}

// logErrorf writes error logs to a file and optionally to the console.
func logErrorf(logFile *os.File, cliDebug bool, format string, args ...interface{}) {
	logLine := fmt.Sprintf(format, args...)
	logFile.WriteString(logLine + "\n")
	if cliDebug {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

// generateLogFilePath creates a sortable log file path with a timestamp.
func generateLogFilePath() (string, error) {
	now := time.Now().UTC()
	logFileName := now.Format("20060102T150405.000000") + ".log"
	logDir := ".kubecopilot/logs"
	logFilePath := filepath.Join(logDir, logFileName)

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create log directory: %w", err)
	}
	return logFilePath, nil
}
