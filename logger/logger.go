package logger

import (
	"log"
	"os"
)

type LoggerConfig struct {
	Level  string
	Format string
	File   string
}

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
	log.Printf(formatLogLine("DEBG", line), args...)
}

func Info(line string, args ...interface{}) {
	log.Printf(formatLogLine("INFO", line), args...)
}

func Warning(line string, args ...interface{}) {
	log.Printf(formatLogLine("WARN", line), args...)
}

func Error(line string, args ...interface{}) {
	log.Printf(formatLogLine("ERRO", line), args...)
}

func Critical(line string, args ...interface{}) {
	log.Fatalf(formatLogLine("CRIT", line), args...)
}

func formatLogLine(level string, line string) string {
	return level + ": " + line
}
