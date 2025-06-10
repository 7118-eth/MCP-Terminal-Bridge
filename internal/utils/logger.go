package utils

import (
	"log/slog"
	"os"
	"strings"
)

var Logger *slog.Logger

func InitLogger() {
	// Get log level from environment
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "info"
	}

	var level slog.Level
	switch strings.ToLower(levelStr) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler with options
	opts := &slog.HandlerOptions{
		Level: level,
		AddSource: level == slog.LevelDebug, // Add source info in debug mode
	}

	// Use JSON handler for structured output
	handler := slog.NewJSONHandler(os.Stderr, opts)
	Logger = slog.New(handler)

	// Set as default
	slog.SetDefault(Logger)

	Logger.Info("Logger initialized", 
		slog.String("level", levelStr),
		slog.Bool("source", opts.AddSource),
	)
}

// Helper functions for common logging patterns

func LogError(err error, msg string, args ...any) {
	if err != nil {
		args = append(args, slog.String("error", err.Error()))
		Logger.Error(msg, args...)
	}
}

func LogSessionEvent(sessionID string, event string, args ...any) {
	args = append([]any{
		slog.String("session_id", sessionID),
		slog.String("event", event),
	}, args...)
	Logger.Info("session event", args...)
}

func LogToolCall(tool string, sessionID string, args ...any) {
	baseArgs := []any{
		slog.String("tool", tool),
	}
	if sessionID != "" {
		baseArgs = append(baseArgs, slog.String("session_id", sessionID))
	}
	args = append(baseArgs, args...)
	Logger.Debug("tool call", args...)
}