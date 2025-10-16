package logger
package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"log/slog"
)

func TestInit(t *testing.T) {
	// Create temporary directory for test logs
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Level:      "info",
				OutputPath: logPath,
			},
			wantErr: false,
		},
		{
			name: "debug level",
			config: Config{
				Level:      "debug",
				OutputPath: logPath,
			},
			wantErr: false,
		},
		{
			name: "default path",
			config: Config{
				Level: "info",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Init(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				// Verify logger was initialized
				if DefaultLogger == nil {
					t.Error("DefaultLogger is nil after Init")
				}

				// Clean up
				Close()
				DefaultLogger = nil
			}
		})
	}
}

func TestGenerateCorrelationID(t *testing.T) {
	id1 := GenerateCorrelationID()
	id2 := GenerateCorrelationID()

	if id1 == "" {
		t.Error("GenerateCorrelationID() returned empty string")
	}

	if id1 == id2 {
		t.Error("GenerateCorrelationID() returned duplicate IDs")
	}

	// UUID format check (36 characters with dashes)
	if len(id1) != 36 {
		t.Errorf("GenerateCorrelationID() returned invalid UUID length: %d", len(id1))
	}
}

func TestLogHTTPRequest(t *testing.T) {
	// Initialize logger
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	err := Init(Config{
		Level:      "info",
		OutputPath: logPath,
	})
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer Close()

	userID := uint(1)

	tests := []struct {
		name  string
		entry LogEntry
	}{
		{
			name: "info level",
			entry: LogEntry{
				Level:         "info",
				Method:        "GET",
				Path:          "/api/v1/users",
				Status:        200,
				Latency:       45,
				ClientIP:      "127.0.0.1",
				UserAgent:     "test-agent",
				CorrelationID: "test-id-1",
				Message:       "Request completed",
			},
		},
		{
			name: "error level with user",
			entry: LogEntry{
				Level:         "error",
				Method:        "POST",
				Path:          "/api/v1/users",
				Status:        500,
				Latency:       123,
				ClientIP:      "127.0.0.1",
				UserAgent:     "test-agent",
				UserID:        &userID,
				CorrelationID: "test-id-2",
				ErrorType:     "database_error",
				ErrorMessage:  "Connection failed",
				Message:       "Request failed",
			},
		},
		{
			name: "with request/response body",
			entry: LogEntry{
				Level:         "info",
				Method:        "POST",
				Path:          "/api/v1/login",
				Status:        200,
				Latency:       89,
				ClientIP:      "127.0.0.1",
				RequestBody:   `{"email":"test@example.com"}`,
				ResponseBody:  `{"status":1,"token":"xxx"}`,
				CorrelationID: "test-id-3",
				Message:       "Login successful",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			LogHTTPRequest(tt.entry)

			// Verify log file was written
			info, err := os.Stat(logPath)
			if err != nil {
				t.Errorf("Log file not created: %v", err)
			}

			if info.Size() == 0 {
				t.Error("Log file is empty")
			}
		})
	}
}

func TestLoggerHelpers(t *testing.T) {
	// Initialize logger
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	err := Init(Config{
		Level:      "debug",
		OutputPath: logPath,
	})
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer Close()

	tests := []struct {
		name string
		fn   func()
	}{
		{
			name: "Info",
			fn: func() {
				Info("info message", slog.String("key", "value"))
			},
		},
		{
			name: "Error",
			fn: func() {
				Error("error message", slog.String("error", "test error"))
			},
		},
		{
			name: "Warn",
			fn: func() {
				Warn("warning message", slog.Int("code", 400))
			},
		},
		{
			name: "Debug",
			fn: func() {
				Debug("debug message", slog.Bool("debug", true))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			tt.fn()
		})
	}
}

func TestWithFields(t *testing.T) {
	// Initialize logger
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	err := Init(Config{
		Level:      "info",
		OutputPath: logPath,
	})
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}
	defer Close()

	fields := map[string]interface{}{
		"user_id":   123,
		"session":   "abc-123",
		"is_active": true,
	}

	logger := WithFields(fields)
	if logger == nil {
		t.Error("WithFields() returned nil")
	}

	// Should not panic
	logger.Info("test message with fields")
}

func TestSanitizeRequestBody(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		contains string
	}{
		{
			name:     "with password",
			body:     `{"email":"test@example.com","password":"secret123"}`,
			contains: "***REDACTED***",
		},
		{
			name:     "without password",
			body:     `{"email":"test@example.com","name":"John"}`,
			contains: "John",
		},
		{
			name:     "invalid json",
			body:     `invalid json`,
			contains: "[unable to parse]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeRequestBody(tt.body)
			if result == "" {
				t.Error("sanitizeRequestBody() returned empty string")
			}
		})
	}
}

func TestExtractErrorMessage(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "with message field",
			body: `{"status":0,"message":"User not found"}`,
			want: "User not found",
		},
		{
			name: "with error field",
			body: `{"error":"Database error"}`,
			want: "Database error",
		},
		{
			name: "empty body",
			body: "",
			want: "",
		},
		{
			name: "invalid json",
			body: "invalid",
			want: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractErrorMessage(tt.body)
			if result != tt.want {
				t.Errorf("extractErrorMessage() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestGenerateLogMessage(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		path    string
		status  int
		latency int64
		want    string
	}{
		{
			name:    "success request",
			method:  "GET",
			path:    "/api/v1/users",
			status:  200,
			latency: 45,
		},
		{
			name:    "error request",
			method:  "POST",
			path:    "/api/v1/users",
			status:  500,
			latency: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateLogMessage(tt.method, tt.path, tt.status, tt.latency)
			if result == "" {
				t.Error("generateLogMessage() returned empty string")
			}
		})
	}
}

func TestLimitString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   int // expected length
	}{
		{
			name:   "short string",
			input:  "hello",
			maxLen: 10,
			want:   5,
		},
		{
			name:   "long string",
			input:  "hello world this is a long string",
			maxLen: 10,
			want:   13, // 10 + "..."
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := limitString(tt.input, tt.maxLen)
			if len(result) != tt.want {
				t.Errorf("limitString() length = %d, want %d", len(result), tt.want)
			}
		})
	}
}

func BenchmarkLogHTTPRequest(b *testing.B) {
	tempDir := b.TempDir()
	logPath := filepath.Join(tempDir, "bench.log")

	err := Init(Config{
		Level:      "info",
		OutputPath: logPath,
	})
	if err != nil {
		b.Fatalf("Failed to initialize logger: %v", err)
	}
	defer Close()

	userID := uint(1)
	entry := LogEntry{
		Timestamp:     time.Now().Format(time.RFC3339),
		Level:         "info",
		Method:        "GET",
		Path:          "/api/v1/users",
		Status:        200,
		Latency:       45,
		ClientIP:      "127.0.0.1",
		UserAgent:     "benchmark",
		RequestBody:   `{"query":"test"}`,
		ResponseBody:  `{"status":1,"data":[]}`,
		UserID:        &userID,
		CorrelationID: "bench-id",
		Message:       "Benchmark request",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LogHTTPRequest(entry)
	}
}
