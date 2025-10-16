package logger

import (
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
	logFile *os.File
}

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error
	OutputPath string // path to log file
	MaxSize    int64  // max size in MB before rotation
}

var (
	// DefaultLogger is the global logger instance
	DefaultLogger *Logger
)

// Init initializes the global logger
func Init(config Config) error {
	if config.OutputPath == "" {
		config.OutputPath = "logs/app.log"
	}

	// Create logs directory if not exists
	logDir := filepath.Dir(config.OutputPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Open log file
	logFile, err := os.OpenFile(config.OutputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// Create multi-writer (file + stdout)
	multiWriter := io.MultiWriter(logFile, os.Stdout)

	// Parse log level
	var level slog.Level
	switch config.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create JSON handler for structured logging
	handler := slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Rename 'time' to 'timestamp' for Kibana
			if a.Key == slog.TimeKey {
				a.Key = "timestamp"
			}
			return a
		},
	})

	logger := slog.New(handler)

	DefaultLogger = &Logger{
		Logger:  logger,
		logFile: logFile,
	}

	return nil
}

// Close closes the log file
func Close() error {
	if DefaultLogger != nil && DefaultLogger.logFile != nil {
		return DefaultLogger.logFile.Close()
	}
	return nil
}

// GenerateCorrelationID generates a unique correlation ID for request tracking
func GenerateCorrelationID() string {
	return uuid.New().String()
}

// LogEntry represents a structured log entry for HTTP requests
type LogEntry struct {
	Timestamp      string                 `json:"timestamp"`
	Level          string                 `json:"level"`
	Method         string                 `json:"method,omitempty"`
	Path           string                 `json:"path,omitempty"`
	Status         int                    `json:"status,omitempty"`
	Latency        int64                  `json:"latency_ms,omitempty"`
	ClientIP       string                 `json:"client_ip,omitempty"`
	UserAgent      string                 `json:"user_agent,omitempty"`
	RequestBody    string                 `json:"request_body,omitempty"`
	ResponseBody   string                 `json:"response_body,omitempty"`
	UserID         *uint                  `json:"user_id,omitempty"`
	CorrelationID  string                 `json:"correlation_id,omitempty"`
	ErrorType      string                 `json:"error_type,omitempty"`
	ErrorMessage   string                 `json:"error_message,omitempty"`
	Message        string                 `json:"message,omitempty"`
	AdditionalData map[string]interface{} `json:"additional_data,omitempty"`
}

// LogHTTPRequest logs HTTP request/response with structured format
func LogHTTPRequest(entry LogEntry) {
	if DefaultLogger == nil {
		return
	}

	// Set timestamp if not provided
	if entry.Timestamp == "" {
		entry.Timestamp = time.Now().Format(time.RFC3339)
	}

	// Convert to JSON for structured logging
	data, err := json.Marshal(entry)
	if err != nil {
		DefaultLogger.Error("Failed to marshal log entry",
			slog.String("error", err.Error()),
		)
		return
	}

	// Write raw JSON to preserve structure
	var logData map[string]interface{}
	if err := json.Unmarshal(data, &logData); err != nil {
		DefaultLogger.Error("Failed to unmarshal log entry")
		return
	}

	// Log based on level
	switch entry.Level {
	case "error":
		DefaultLogger.Error(entry.Message,
			convertToSlogAttrs(logData)...,
		)
	case "warn":
		DefaultLogger.Warn(entry.Message,
			convertToSlogAttrs(logData)...,
		)
	case "debug":
		DefaultLogger.Debug(entry.Message,
			convertToSlogAttrs(logData)...,
		)
	default:
		DefaultLogger.Info(entry.Message,
			convertToSlogAttrs(logData)...,
		)
	}
}

// convertToSlogAttrs converts map to slog attributes
func convertToSlogAttrs(data map[string]interface{}) []any {
	attrs := make([]any, 0, len(data))
	for k, v := range data {
		// Skip fields that are already included in message
		if k == "timestamp" || k == "level" || k == "message" || k == "msg" {
			continue
		}

		switch val := v.(type) {
		case string:
			if val != "" {
				attrs = append(attrs, slog.String(k, val))
			}
		case int:
			if val != 0 {
				attrs = append(attrs, slog.Int(k, val))
			}
		case int64:
			if val != 0 {
				attrs = append(attrs, slog.Int64(k, val))
			}
		case float64:
			if val != 0 {
				attrs = append(attrs, slog.Float64(k, val))
			}
		case bool:
			attrs = append(attrs, slog.Bool(k, val))
		case nil:
			// Skip nil values
		default:
			// Handle other types as strings
			attrs = append(attrs, slog.Any(k, v))
		}
	}
	return attrs
}

// Info logs info level message
func Info(msg string, args ...any) {
	if DefaultLogger != nil {
		DefaultLogger.Info(msg, args...)
	}
}

// Error logs error level message
func Error(msg string, args ...any) {
	if DefaultLogger != nil {
		DefaultLogger.Error(msg, args...)
	}
}

// Warn logs warn level message
func Warn(msg string, args ...any) {
	if DefaultLogger != nil {
		DefaultLogger.Warn(msg, args...)
	}
}

// Debug logs debug level message
func Debug(msg string, args ...any) {
	if DefaultLogger != nil {
		DefaultLogger.Debug(msg, args...)
	}
}

// WithFields creates a logger with additional fields
func WithFields(fields map[string]interface{}) *slog.Logger {
	if DefaultLogger == nil {
		return nil
	}

	attrs := make([]any, 0, len(fields))
	for k, v := range fields {
		attrs = append(attrs, slog.Any(k, v))
	}

	return DefaultLogger.With(attrs...)
}
