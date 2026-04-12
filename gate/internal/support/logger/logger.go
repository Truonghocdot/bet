package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Init configures the standard logger to write to both stdout and a rotating log file.
func Init(logFilePath string) {
	// Ensure the directory exists
	if dir := filepath.Dir(logFilePath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Printf("failed to create log directory: %v", err)
			return
		}
	}

	fileLogger := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    50, // megabytes
		MaxBackups: 10,
		MaxAge:     28,   // days
		Compress:   true, // disabled by default
	}

	multiWriter := io.MultiWriter(os.Stdout, fileLogger)
	log.SetOutput(multiWriter)
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
}
