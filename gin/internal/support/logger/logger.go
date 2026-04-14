package logger

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Init configures the standard logger to write to both stdout and a rotating log file.
func Init(logFilePath string) {
	buildWriter := func(path string) io.Writer {
		if path == "" {
			return nil
		}
		if dir := filepath.Dir(path); dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil
			}
		}
		return &lumberjack.Logger{
			Filename:   path,
			MaxSize:    50, // megabytes
			MaxBackups: 10,
			MaxAge:     28,   // days
			Compress:   true, // disabled by default
		}
	}

	primary := buildWriter(logFilePath)
	writers := []io.Writer{os.Stdout}
	if primary != nil {
		writers = append(writers, primary)
	}

	mirrorPath := strings.TrimSpace(os.Getenv("LOG_VIEWER_FILE"))
	if mirrorPath == "" {
		candidate := filepath.Join("..", "admin", "storage", "logs", filepath.Base(logFilePath))
		if _, err := os.Stat(filepath.Dir(candidate)); err == nil {
			mirrorPath = candidate
		}
	}
	if mirrorPath != "" && filepath.Clean(mirrorPath) != filepath.Clean(logFilePath) {
		if mirror := buildWriter(mirrorPath); mirror != nil {
			writers = append(writers, mirror)
		}
	}

	appEnv := strings.TrimSpace(os.Getenv("APP_ENV"))
	if appEnv == "" {
		appEnv = "local"
	}
	minLevel := parseLevel(strings.TrimSpace(os.Getenv("LOG_LEVEL")))
	multiWriter := io.MultiWriter(writers...)
	log.SetOutput(&laravelWriter{
		out:      multiWriter,
		env:      appEnv,
		minLevel: minLevel,
	})
	log.SetFlags(0)
}

func Debugf(format string, args ...any) { log.Printf("DEBUG "+format, args...) }
func Infof(format string, args ...any)  { log.Printf("INFO "+format, args...) }
func Warnf(format string, args ...any)  { log.Printf("WARN "+format, args...) }
func Errorf(format string, args ...any) { log.Printf("ERROR "+format, args...) }

type laravelWriter struct {
	out      io.Writer
	env      string
	minLevel logLevel
}

func (w *laravelWriter) Write(p []byte) (int, error) {
	raw := string(bytes.TrimSpace(p))
	if raw == "" {
		return len(p), nil
	}
	level := detectLevel(raw)
	if level < w.minLevel {
		return len(p), nil
	}

	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		normalized := stripLevelPrefix(trimmed)
		formatted := "[" + time.Now().Format("2006-01-02 15:04:05") + "] " + w.env + "." + level.String() + ": " + normalized + "\n"
		if _, err := w.out.Write([]byte(formatted)); err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

type logLevel int

const (
	levelDebug logLevel = iota
	levelInfo
	levelWarn
	levelError
)

func parseLevel(raw string) logLevel {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return levelDebug
	case "warn", "warning":
		return levelWarn
	case "error":
		return levelError
	default:
		return levelInfo
	}
}

func detectLevel(message string) logLevel {
	upper := strings.ToUpper(message)
	switch {
	case strings.Contains(upper, "FATAL"), strings.Contains(upper, "PANIC"), strings.Contains(upper, "ERROR"), strings.Contains(upper, ".ERROR]"), strings.Contains(upper, "[ERROR]"):
		return levelError
	case strings.Contains(upper, "WARN"), strings.Contains(upper, ".WARN]"), strings.Contains(upper, "[WARN]"), strings.Contains(upper, "WARNING"):
		return levelWarn
	case strings.HasPrefix(upper, "DEBUG "), strings.Contains(upper, ".DEBUG]"), strings.Contains(upper, "[DEBUG]"):
		return levelDebug
	default:
		return levelInfo
	}
}

func stripLevelPrefix(message string) string {
	switch {
	case strings.HasPrefix(strings.ToUpper(message), "DEBUG "):
		return strings.TrimSpace(message[6:])
	case strings.HasPrefix(strings.ToUpper(message), "INFO "):
		return strings.TrimSpace(message[5:])
	case strings.HasPrefix(strings.ToUpper(message), "WARN "):
		return strings.TrimSpace(message[5:])
	case strings.HasPrefix(strings.ToUpper(message), "ERROR "):
		return strings.TrimSpace(message[6:])
	default:
		return message
	}
}

func (l logLevel) String() string {
	switch l {
	case levelDebug:
		return "DEBUG"
	case levelWarn:
		return "WARNING"
	case levelError:
		return "ERROR"
	default:
		return "INFO"
	}
}
