package logging

import (
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// LoggingMiddleware provides structured logging for HTTP requests
type LoggingMiddleware struct {
	logger *zap.Logger
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// LogRequest logs HTTP request details with structured fields
func (m *LoggingMiddleware) LogRequest(r *http.Request, additionalFields ...zap.Field) {
	fields := []zap.Field{
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.String("request_id", generateRequestID()),
	}
	
	// Add query parameters if present
	if r.URL.RawQuery != "" {
		fields = append(fields, zap.String("query", r.URL.RawQuery))
	}
	
	// Add additional fields
	fields = append(fields, additionalFields...)
	
	m.logger.Info("request started", fields...)
}

// LogResponse logs HTTP response details with structured fields
func (m *LoggingMiddleware) LogResponse(r *http.Request, statusCode int, duration time.Duration, additionalFields ...zap.Field) {
	fields := []zap.Field{
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Int("status", statusCode),
		zap.Duration("duration", duration),
		zap.String("remote_addr", r.RemoteAddr),
	}
	
	// Add additional fields
	fields = append(fields, additionalFields...)
	
	// Determine log level based on status code
	if statusCode >= 500 {
		m.logger.Error("request completed", fields...)
	} else if statusCode >= 400 {
		m.logger.Warn("request completed", fields...)
	} else {
		m.logger.Info("request completed", fields...)
	}
}

// LogAPIValidation logs API validation events
func (m *LoggingMiddleware) LogAPIValidation(r *http.Request, valid bool, reason string, additionalFields ...zap.Field) {
	fields := []zap.Field{
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Bool("valid", valid),
		zap.String("reason", reason),
		zap.String("remote_addr", r.RemoteAddr),
	}
	
	fields = append(fields, additionalFields...)
	
	if valid {
		m.logger.Info("api validation successful", fields...)
	} else {
		m.logger.Warn("api validation failed", fields...)
	}
}

// LogAPIKeyValidation logs API key validation events
func (m *LoggingMiddleware) LogAPIKeyValidation(r *http.Request, keyName string, valid bool, subscription string, additionalFields ...zap.Field) {
	fields := []zap.Field{
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("api_key_name", keyName),
		zap.Bool("valid", valid),
		zap.String("subscription", subscription),
		zap.String("remote_addr", r.RemoteAddr),
	}
	
	fields = append(fields, additionalFields...)
	
	if valid {
		m.logger.Info("api key validation successful", fields...)
	} else {
		m.logger.Warn("api key validation failed", fields...)
	}
}

// LogError logs error events with structured fields
func (m *LoggingMiddleware) LogError(r *http.Request, err error, additionalFields ...zap.Field) {
	fields := []zap.Field{
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.Error(err),
		zap.String("remote_addr", r.RemoteAddr),
	}
	
	fields = append(fields, additionalFields...)
	
	m.logger.Error("request error", fields...)
}

// WrapHandler wraps an http.Handler with logging middleware
func (m *LoggingMiddleware) WrapHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Wrap response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Log request start
		m.LogRequest(r)
		
		// Call next handler
		next.ServeHTTP(rw, r)
		
		// Log request completion
		duration := time.Since(start)
		m.LogResponse(r, rw.statusCode, duration)
	})
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}