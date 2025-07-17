package logging

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LokiHook implements zapcore.WriteSyncer to send logs to Loki
type LokiHook struct {
	lokiURL string
	labels  map[string]string
	client  *http.Client
}

// LokiLog represents a single log entry for Loki
type LokiLog struct {
	Streams []LokiStream `json:"streams"`
}

// LokiStream represents a stream of logs with labels
type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// NewLokiHook creates a new Loki hook for zap logger
func NewLokiHook(lokiURL string, labels map[string]string) *LokiHook {
	if labels == nil {
		labels = make(map[string]string)
	}
	
	// Add default labels
	if _, exists := labels["job"]; !exists {
		labels["job"] = "veil-application"
	}
	
	return &LokiHook{
		lokiURL: lokiURL + "/loki/api/v1/push",
		labels:  labels,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Write implements zapcore.WriteSyncer
func (h *LokiHook) Write(p []byte) (n int, err error) {
	// Parse the JSON log entry to extract timestamp and message
	var logEntry map[string]interface{}
	if err := json.Unmarshal(p, &logEntry); err != nil {
		// If we can't parse JSON, send as plain text
		return h.sendToLoki(string(p), time.Now())
	}

	// Extract timestamp
	var timestamp time.Time
	if ts, ok := logEntry["ts"].(string); ok {
		if parsedTime, err := time.Parse(time.RFC3339, ts); err == nil {
			timestamp = parsedTime
		} else {
			timestamp = time.Now()
		}
	} else {
		timestamp = time.Now()
	}

	// Convert back to JSON string for Loki
	logJSON, err := json.Marshal(logEntry)
	if err != nil {
		return h.sendToLoki(string(p), timestamp)
	}

	return h.sendToLoki(string(logJSON), timestamp)
}

// Sync implements zapcore.WriteSyncer
func (h *LokiHook) Sync() error {
	return nil
}

// sendToLoki sends a log entry to Loki
func (h *LokiHook) sendToLoki(message string, timestamp time.Time) (int, error) {
	// Create labels with dynamic values from log entry
	labels := make(map[string]string)
	for k, v := range h.labels {
		labels[k] = v
	}

	// Try to extract additional labels from JSON if possible
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(message), &logEntry); err == nil {
		// Add level as label if present
		if level, ok := logEntry["level"].(string); ok {
			labels["level"] = level
		}
		
		// Add method as label if present
		if method, ok := logEntry["method"].(string); ok {
			labels["method"] = method
		}
		
		// Add status as label if present  
		if status, ok := logEntry["status"]; ok {
			labels["status"] = fmt.Sprintf("%v", status)
		}
		
		// Add path as label if present
		if path, ok := logEntry["path"].(string); ok {
			labels["path"] = path
		}
	}

	lokiLog := LokiLog{
		Streams: []LokiStream{
			{
				Stream: labels,
				Values: [][]string{
					{
						fmt.Sprintf("%d", timestamp.UnixNano()),
						message,
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(lokiLog)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", h.lokiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		// Read error response
		scanner := bufio.NewScanner(resp.Body)
		var errorMsg string
		for scanner.Scan() {
			errorMsg += scanner.Text()
		}
		return 0, fmt.Errorf("loki returned status %d: %s", resp.StatusCode, errorMsg)
	}

	return len(message), nil
}

// CreateLokiLogger creates a zap logger configured to send logs to Loki
func CreateLokiLogger(lokiURL string, labels map[string]string) (*zap.Logger, error) {
	// Create Loki hook
	lokiHook := NewLokiHook(lokiURL, labels)

	// Create encoder config for JSON output
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "ts"
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderConfig.LevelKey = "level"
	encoderConfig.MessageKey = "msg"
	encoderConfig.CallerKey = "caller"
	encoderConfig.StacktraceKey = "stacktrace"

	// Create core that writes to both console and Loki
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	core := zapcore.NewTee(
		// Console output for local development
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.InfoLevel),
		// Loki output for centralized logging
		zapcore.NewCore(jsonEncoder, lokiHook, zapcore.DebugLevel),
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return logger, nil
}