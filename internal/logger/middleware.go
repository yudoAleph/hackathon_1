package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ResponseWriter wraps gin.ResponseWriter to capture response body
type ResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body
func (w *ResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggingMiddleware logs all HTTP requests and responses
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate correlation ID
		correlationID := GenerateCorrelationID()
		c.Set("correlation_id", correlationID)

		// Start timer
		startTime := time.Now()

		// Capture request body
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				// Restore the body for downstream handlers
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Sanitize sensitive data (passwords)
				if strings.Contains(c.Request.URL.Path, "/auth/") {
					requestBody = sanitizeRequestBody(requestBody)
				}
			}
		}

		// Wrap response writer to capture response body
		responseWriter := &ResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = responseWriter

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime).Milliseconds()

		// Get user ID from context (set by auth middleware)
		var userID *uint
		if uid, exists := c.Get("userID"); exists {
			if id, ok := uid.(uint); ok {
				userID = &id
			}
		}

		// Capture response body
		responseBody := responseWriter.body.String()

		// Limit response body size for logging (max 1000 chars)
		if len(responseBody) > 1000 {
			responseBody = responseBody[:1000] + "... (truncated)"
		}

		// Determine log level based on status code
		level := "info"
		var errorType, errorMessage string

		status := c.Writer.Status()
		if status >= 500 {
			level = "error"
			errorType = "server_error"
			errorMessage = extractErrorMessage(responseBody)
		} else if status >= 400 {
			level = "warn"
			errorType = "client_error"
			errorMessage = extractErrorMessage(responseBody)
		}

		// Check if there were any errors
		if len(c.Errors) > 0 {
			level = "error"
			errorType = "handler_error"
			errorMessage = c.Errors.String()
		}

		// Create log entry
		entry := LogEntry{
			Timestamp:     time.Now().Format(time.RFC3339),
			Level:         level,
			Method:        c.Request.Method,
			Path:          c.Request.URL.Path,
			Status:        status,
			Latency:       latency,
			ClientIP:      c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
			RequestBody:   limitString(requestBody, 500),
			ResponseBody:  limitString(responseBody, 500),
			UserID:        userID,
			CorrelationID: correlationID,
			ErrorType:     errorType,
			ErrorMessage:  errorMessage,
			Message:       generateLogMessage(c.Request.Method, c.Request.URL.Path, status, latency),
		}

		// Log the request
		LogHTTPRequest(entry)
	}
}

// sanitizeRequestBody removes sensitive data from request body
func sanitizeRequestBody(body string) string {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return "[unable to parse]"
	}

	// Remove password field
	if _, exists := data["password"]; exists {
		data["password"] = "***REDACTED***"
	}

	sanitized, err := json.Marshal(data)
	if err != nil {
		return "[unable to sanitize]"
	}

	return string(sanitized)
}

// extractErrorMessage extracts error message from response body
func extractErrorMessage(body string) string {
	if body == "" {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return body
	}

	// Try to extract message field
	if msg, ok := data["message"].(string); ok {
		return msg
	}

	// Try to extract error field
	if err, ok := data["error"].(string); ok {
		return err
	}

	return ""
}

// generateLogMessage generates a human-readable log message
func generateLogMessage(method, path string, status int, latency int64) string {
	return strings.TrimSpace(strings.Join([]string{
		method,
		path,
		"-",
		statusText(status),
		"(",
		formatLatency(latency),
		")",
	}, " "))
}

// statusText returns human-readable status text
func statusText(status int) string {
	switch {
	case status >= 200 && status < 300:
		return "✓"
	case status >= 300 && status < 400:
		return "→"
	case status >= 400 && status < 500:
		return "⚠"
	case status >= 500:
		return "✗"
	default:
		return "?"
	}
}

// formatLatency formats latency in human-readable format
func formatLatency(latency int64) string {
	if latency < 1000 {
		return formatInt(latency) + "ms"
	}
	return formatFloat(float64(latency)/1000.0, 2) + "s"
}

// limitString limits string length
func limitString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// formatInt formats integer
func formatInt(n int64) string {
	return string(rune(n + '0'))
}

// formatFloat formats float with precision
func formatFloat(f float64, precision int) string {
	return formatFloatString(f, precision)
}

// formatFloatString converts float to string with precision
func formatFloatString(f float64, precision int) string {
	// Simple implementation for formatting
	s := ""
	if f < 0 {
		s = "-"
		f = -f
	}

	// Integer part
	intPart := int64(f)
	s += formatIntString(intPart)

	// Decimal part
	if precision > 0 {
		s += "."
		f -= float64(intPart)
		for i := 0; i < precision; i++ {
			f *= 10
			digit := int64(f) % 10
			s += formatIntString(digit)
		}
	}

	return s
}

// formatIntString converts int64 to string
func formatIntString(n int64) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}
