package logger

import (
	"fmt"
	"log"
	"os"
)

type LoggerConfig struct {
	Level  string
	Format string
	File   string
}

var colorFormat = "[\x1b[%dm%s\x1b[0m]"

func init() {
	config := LoggerConfig{Format: "stdout", Level: "debug"}
	NewLogger(&config)
}

func NewLogger(config *LoggerConfig) {
	if config.Level == "debug" {
		log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	} else {
		log.SetFlags(log.Ldate | log.Ltime)
	}

	if config.Format == "log" {
		f, err := os.OpenFile(config.File, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			Critical("error opening file: %v", err)
		}

		log.SetOutput(f)
		defer f.Close()
	}
}

func Debug(line string, args ...interface{}) {
	log.Printf(formatLogLine("DEBG", line, 36), args...)
}

func Info(line string, args ...interface{}) {
	log.Printf(formatLogLine("INFO", line, 32), args...)
}

func Warning(line string, args ...interface{}) {
	log.Printf(formatLogLine("WARN", line, 31), args...)
}

func Error(line string, args ...interface{}) {
	log.Printf(formatLogLine("ERRO", line, 31), args...)
}

func Critical(line string, args ...interface{}) {
	log.Fatalf(formatLogLine("CRIT", line, 31), args...)
}

func formatLogLine(level, line string, color int) string {
	return fmt.Sprintf(colorFormat, color, level) + " " + line
}
