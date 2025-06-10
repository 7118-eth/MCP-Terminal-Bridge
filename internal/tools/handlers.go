package tools

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/bioharz/mcp-terminal-tester/internal/session"
	"github.com/bioharz/mcp-terminal-tester/internal/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

type Handlers struct {
	sessionManager *session.Manager
}

func NewHandlers(sm *session.Manager) *Handlers {
	return &Handlers{
		sessionManager: sm,
	}
}

func (h *Handlers) LaunchApp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	utils.LogToolCall("launch_app", "")
	
	args := request.GetArguments()
	command, ok := args["command"].(string)
	if !ok {
		err := fmt.Errorf("command parameter is required")
		slog.Error("Invalid tool call", 
			slog.String("tool", "launch_app"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// Extract args if provided
	var cmdArgs []string
	if argsParam, exists := args["args"]; exists {
		// Try []interface{} first
		if argsArray, ok := argsParam.([]interface{}); ok {
			for _, arg := range argsArray {
				if argStr, ok := arg.(string); ok {
					cmdArgs = append(cmdArgs, argStr)
				}
			}
		} else if argsArray, ok := argsParam.([]string); ok {
			// Also try []string directly
			cmdArgs = argsArray
		}
		slog.Debug("Extracted args", 
			slog.String("tool", "launch_app"),
			slog.Any("args", cmdArgs),
			slog.Any("raw_args", argsParam),
		)
	}

	// Extract env if provided
	env := make(map[string]string)
	if envParam, exists := args["env"]; exists {
		if envMap, ok := envParam.(map[string]interface{}); ok {
			for k, v := range envMap {
				if vStr, ok := v.(string); ok {
					env[k] = vStr
				}
			}
		}
	}

	// Create new session
	sess, err := h.sessionManager.CreateSession(command, cmdArgs, env)
	if err != nil {
		utils.LogError(err, "Failed to launch app",
			slog.String("tool", "launch_app"),
			slog.String("command", command),
		)
		return nil, fmt.Errorf("failed to launch app: %w", err)
	}

	slog.Info("App launched successfully",
		slog.String("tool", "launch_app"),
		slog.String("session_id", sess.ID),
		slog.String("command", command),
	)

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf(`{"session_id": "%s", "success": true}`, sess.ID),
			},
		},
	}, nil
}

func (h *Handlers) ViewScreen(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	sessionID, ok := args["session_id"].(string)
	if !ok {
		err := fmt.Errorf("session_id parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "view_screen"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	utils.LogToolCall("view_screen", sessionID)

	format := "plain"
	if formatParam, exists := args["format"]; exists {
		if f, ok := formatParam.(string); ok {
			format = f
		}
	}

	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	content, err := sess.GetScreen(format)
	if err != nil {
		return nil, err
	}

	row, col := sess.GetCursorPosition()

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf(`{"content": %q, "cursor": {"row": %d, "col": %d}}`, content, row, col),
			},
		},
	}, nil
}

func (h *Handlers) SendKeys(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	sessionID, ok := args["session_id"].(string)
	if !ok {
		err := fmt.Errorf("session_id parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "send_keys"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	keys, ok := args["keys"].(string)
	if !ok {
		err := fmt.Errorf("keys parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "send_keys"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	utils.LogToolCall("send_keys", sessionID, slog.Int("key_count", len(keys)))

	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	// Map special keys
	mappedKeys := MapKeys(keys)
	if mappedKeys != keys {
		slog.Debug("Keys mapped",
			slog.String("original", keys),
			slog.String("mapped", fmt.Sprintf("%q", mappedKeys)),
		)
	}

	if err := sess.SendKeys(mappedKeys); err != nil {
		utils.LogError(err, "Failed to send keys",
			slog.String("tool", "send_keys"),
			slog.String("session_id", sessionID),
		)
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: `{"success": true}`,
			},
		},
	}, nil
}

func (h *Handlers) GetCursorPosition(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	sessionID, ok := args["session_id"].(string)
	if !ok {
		err := fmt.Errorf("session_id parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "get_cursor_position"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	utils.LogToolCall("get_cursor_position", sessionID)

	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	row, col := sess.GetCursorPosition()

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf(`{"row": %d, "col": %d}`, row, col),
			},
		},
	}, nil
}

func (h *Handlers) GetScreenSize(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	sessionID, ok := args["session_id"].(string)
	if !ok {
		err := fmt.Errorf("session_id parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "get_screen_size"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	utils.LogToolCall("get_screen_size", sessionID)

	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	width, height := sess.GetScreenSize()

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf(`{"width": %d, "height": %d}`, width, height),
			},
		},
	}, nil
}

func (h *Handlers) RestartApp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	sessionID, ok := args["session_id"].(string)
	if !ok {
		err := fmt.Errorf("session_id parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "restart_app"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	utils.LogToolCall("restart_app", sessionID)

	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if err := sess.Restart(); err != nil {
		return nil, fmt.Errorf("failed to restart app: %w", err)
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: `{"success": true}`,
			},
		},
	}, nil
}

func (h *Handlers) StopApp(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	sessionID, ok := args["session_id"].(string)
	if !ok {
		err := fmt.Errorf("session_id parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "stop_app"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}
	
	utils.LogToolCall("stop_app", sessionID)

	if err := h.sessionManager.RemoveSession(sessionID); err != nil {
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: `{"success": true}`,
			},
		},
	}, nil
}

func (h *Handlers) ListSessions(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	utils.LogToolCall("list_sessions", "")
	
	sessions := h.sessionManager.ListSessions()
	
	slog.Debug("Sessions listed",
		slog.String("tool", "list_sessions"),
		slog.Int("count", len(sessions)),
	)

	// Convert sessions to JSON string
	var sessionStrings []string
	for _, s := range sessions {
		sessionStrings = append(sessionStrings, fmt.Sprintf(`{"id": %q, "command": %q, "state": %q, "created": %q}`, 
			s.ID, s.Command, s.State, s.Created.Format("2006-01-02T15:04:05Z")))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf(`{"sessions": [%s]}`, strings.Join(sessionStrings, ", ")),
			},
		},
	}, nil
}

func (h *Handlers) ResizeTerminal(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()
	
	// Debug logging
	slog.Debug("ResizeTerminal called", 
		slog.String("tool", "resize_terminal"),
		slog.Any("args", args),
	)
	
	sessionID, ok := args["session_id"].(string)
	if !ok {
		err := fmt.Errorf("session_id parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "resize_terminal"),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	// Try to get width as float64 or int
	var width float64
	if w, ok := args["width"].(float64); ok {
		width = w
	} else if w, ok := args["width"].(int); ok {
		width = float64(w)
	} else {
		err := fmt.Errorf("width parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "resize_terminal"),
			slog.String("error", err.Error()),
			slog.Any("width_type", fmt.Sprintf("%T", args["width"])),
		)
		return nil, err
	}

	// Try to get height as float64 or int
	var height float64
	if h, ok := args["height"].(float64); ok {
		height = h
	} else if h, ok := args["height"].(int); ok {
		height = float64(h)
	} else {
		err := fmt.Errorf("height parameter is required")
		slog.Error("Invalid tool call",
			slog.String("tool", "resize_terminal"),
			slog.String("error", err.Error()),
			slog.Any("height_type", fmt.Sprintf("%T", args["height"])),
		)
		return nil, err
	}

	utils.LogToolCall("resize_terminal", sessionID,
		slog.Int("width", int(width)),
		slog.Int("height", int(height)),
	)

	sess, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if err := sess.Resize(int(width), int(height)); err != nil {
		utils.LogError(err, "Failed to resize terminal",
			slog.String("tool", "resize_terminal"),
			slog.String("session_id", sessionID),
		)
		return nil, err
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf(`{"success": true, "width": %d, "height": %d}`, int(width), int(height)),
			},
		},
	}, nil
}