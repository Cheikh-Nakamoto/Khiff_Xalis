package utils

import (
	"fmt"
	"os"
)

type Log struct {
	// You can add fields here if needed, e.g., log level, output destination, etc.
	message        string
	error          string
	file_detection string
	level          string
	output         string
	created_at     string
}

func NewLog() *Log {
	return &Log{}
}

func (l *Log) Info(msg string) {
	l.message = fmt.Sprintf("[INFO] %s", msg)
	l.level = "INFO"
	fmt.Printf("[INFO] %s", msg)
}
func (l *Log) Debug(msg string) {
	l.message = fmt.Sprintf("[DEBUG] %s", msg)
	l.level = "DEBUG"
	fmt.Printf("[DEBUG] %s", msg)
}
func (l *Log) Warning(msg string) {
	l.message = fmt.Sprintf("[WARNING] %s", msg)
	l.level = "WARNING"
	fmt.Printf("[WARNING] %s", msg)
}

func (l *Log) Error(msg string) {
	l.message = fmt.Sprintf("[ERROR] %s", msg)
	l.level = "ERROR"
	fmt.Printf("[ERROR] %s", msg)
}

func (l *Log) Fatal(msg string) {
	l.message = fmt.Sprintf("[FATAL] %s", msg)
	l.level = "FATAL"
	fmt.Printf("[FATAL] %s", msg)
	os.Exit(1) // Terminate the program immediately
	// You can also call os.Exit(1) here if you want to terminate the program on fatal errors.
}

func LogToFile(filename string, message string, level string) error {
	// Open the file in append mode, create it if it doesn't exist
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	colorLevel := SwitchLogLevelColor(level)
	_, err = file.WriteString(fmt.Sprintf("[%s] %s", colorLevel, message))
	if err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}

	return nil
}

func SwitchLogLevelColor(level string) string {
	switch level {
	case "INFO":
		return fmt.Sprintf("\033[34m%s\033[0m", level) // Blue
	case "DEBUG":
		return fmt.Sprintf("\033[36m%s\033[0m", level) // Cyan
	case "WARNING":
		return fmt.Sprintf("\033[33m%s\033[0m", level) // Yellow
	case "ERROR":
		return fmt.Sprintf("\033[31m%s\033[0m", level) // Red
	case "FATAL":
		return fmt.Sprintf("\033[35m%s\033[0m", level) // Magenta
	default:
		return fmt.Sprintf("%s", level) // Default color
	}
}
