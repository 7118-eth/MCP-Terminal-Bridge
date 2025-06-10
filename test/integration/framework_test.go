package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bioharz/mcp-terminal-tester/internal/session"
	"github.com/bioharz/mcp-terminal-tester/internal/tools"
	"github.com/bioharz/mcp-terminal-tester/internal/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestFramework provides a test harness for integration testing
type TestFramework struct {
	manager  *session.Manager
	handlers *tools.Handlers
	t        *testing.T
}

// NewTestFramework creates a new test framework
func NewTestFramework(t *testing.T) *TestFramework {
	utils.InitLogger()
	manager := session.NewManager()
	handlers := tools.NewHandlers(manager)
	
	return &TestFramework{
		manager:  manager,
		handlers: handlers,
		t:        t,
	}
}

// CallTool simulates calling an MCP tool
func (tf *TestFramework) CallTool(toolName string, args map[string]interface{}) (map[string]interface{}, error) {
	ctx := context.Background()
	
	// Create proper CallToolRequest
	request := mcp.CallToolRequest{
		Request: mcp.Request{
			// Jsonrpc and Id would normally be set by the protocol layer
		},
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: args,
		},
	}
	
	// Call the appropriate handler
	var result *mcp.CallToolResult
	var err error
	switch toolName {
	case "launch_app":
		result, err = tf.handlers.LaunchApp(ctx, request)
	case "view_screen":
		result, err = tf.handlers.ViewScreen(ctx, request)
	case "send_keys":
		result, err = tf.handlers.SendKeys(ctx, request)
	case "get_cursor_position":
		result, err = tf.handlers.GetCursorPosition(ctx, request)
	case "get_screen_size":
		result, err = tf.handlers.GetScreenSize(ctx, request)
	case "resize_terminal":
		result, err = tf.handlers.ResizeTerminal(ctx, request)
	case "restart_app":
		result, err = tf.handlers.RestartApp(ctx, request)
	case "stop_app":
		result, err = tf.handlers.StopApp(ctx, request)
	case "list_sessions":
		result, err = tf.handlers.ListSessions(ctx, request)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
	
	if err != nil {
		return nil, err
	}
	
	// Extract response from result
	if len(result.Content) == 0 {
		return nil, fmt.Errorf("empty response")
	}
	
	// Parse the JSON response
	textContent, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		return nil, fmt.Errorf("unexpected content type")
	}
	
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		// Some tools return plain text, not JSON
		response = map[string]interface{}{
			"content": textContent.Text,
		}
	}
	
	return response, nil
}

// LaunchApp is a helper to launch an app and return session ID
func (tf *TestFramework) LaunchApp(command string, args []string) string {
	result, err := tf.CallTool("launch_app", map[string]interface{}{
		"command": command,
		"args":    args,
	})
	if err != nil {
		tf.t.Fatalf("Failed to launch app: %v", err)
	}
	
	sessionID, ok := result["session_id"].(string)
	if !ok {
		tf.t.Fatalf("No session_id in response: %+v", result)
	}
	
	return sessionID
}

// ViewScreen is a helper to view screen content
func (tf *TestFramework) ViewScreen(sessionID string, format string) string {
	result, err := tf.CallTool("view_screen", map[string]interface{}{
		"session_id": sessionID,
		"format":     format,
	})
	if err != nil {
		tf.t.Fatalf("Failed to view screen: %v", err)
	}
	
	content, ok := result["content"].(string)
	if !ok {
		tf.t.Fatalf("No content in response: %+v", result)
	}
	
	return content
}

// SendKeys is a helper to send keys
func (tf *TestFramework) SendKeys(sessionID string, keys string) {
	_, err := tf.CallTool("send_keys", map[string]interface{}{
		"session_id": sessionID,
		"keys":       keys,
	})
	if err != nil {
		tf.t.Fatalf("Failed to send keys: %v", err)
	}
}

// WaitForContent waits for specific content to appear on screen
func (tf *TestFramework) WaitForContent(sessionID string, expected string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		content := tf.ViewScreen(sessionID, "plain")
		if strings.Contains(content, expected) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	return false
}

// StopApp is a helper to stop an app
func (tf *TestFramework) StopApp(sessionID string) {
	_, err := tf.CallTool("stop_app", map[string]interface{}{
		"session_id": sessionID,
	})
	if err != nil {
		tf.t.Fatalf("Failed to stop app: %v", err)
	}
}

// Cleanup cleans up all sessions
func (tf *TestFramework) Cleanup() {
	sessions := tf.manager.ListSessions()
	for _, sess := range sessions {
		tf.manager.RemoveSession(sess.ID)
	}
}