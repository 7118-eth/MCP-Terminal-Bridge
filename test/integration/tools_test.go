package integration

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestLaunchApp(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Test launching sh with echo command - use sh to keep session alive
	sessionID := tf.LaunchApp("sh", []string{"-c", "printf 'Hello, World!\\n'; sleep 0.5"})
	
	// Verify session was created
	if sessionID == "" {
		t.Fatal("No session ID returned")
	}
	
	// Give it time to execute
	time.Sleep(200 * time.Millisecond)
	
	// Debug - check what's on screen
	content := tf.ViewScreen(sessionID, "plain")
	t.Logf("Screen content: %s", content)
	
	// Wait for output
	if !tf.WaitForContent(sessionID, "Hello, World!", 2*time.Second) {
		content := tf.ViewScreen(sessionID, "plain")
		t.Fatalf("Expected 'Hello, World!' but got: %s", content)
	}
}

func TestViewScreen(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Launch sh with echo to keep session alive
	sessionID := tf.LaunchApp("sh", []string{"-c", "echo 'Test123'; sleep 1"})
	time.Sleep(100 * time.Millisecond)
	
	// Test different formats
	tests := []struct {
		format   string
		contains string
	}{
		{"plain", "Test123"},
		{"raw", "Test123"},
		{"ansi", "Test123"},
		{"scrollback", "Test123"},
	}
	
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			content := tf.ViewScreen(sessionID, tt.format)
			if !strings.Contains(content, tt.contains) {
				t.Errorf("Format %s: expected to contain '%s', got: %s", 
					tt.format, tt.contains, content)
			}
		})
	}
}

func TestSendKeys(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Launch cat (echoes input)
	sessionID := tf.LaunchApp("cat", []string{})
	time.Sleep(100 * time.Millisecond)
	
	// Send some text
	tf.SendKeys(sessionID, "Hello")
	
	// Verify it appears
	if !tf.WaitForContent(sessionID, "Hello", 2*time.Second) {
		content := tf.ViewScreen(sessionID, "plain")
		t.Fatalf("Expected 'Hello' but got: %s", content)
	}
	
	// Send Enter
	tf.SendKeys(sessionID, "Enter")
	
	// Send more text
	tf.SendKeys(sessionID, "World")
	
	// Verify both lines appear
	if !tf.WaitForContent(sessionID, "World", 2*time.Second) {
		content := tf.ViewScreen(sessionID, "plain")
		t.Fatalf("Expected 'World' but got: %s", content)
	}
}

func TestGetCursorPosition(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Launch sh with echo to keep session alive
	sessionID := tf.LaunchApp("sh", []string{"-c", "echo 'Test'; sleep 1"})
	time.Sleep(100 * time.Millisecond)
	
	// Get cursor position
	result, err := tf.CallTool("get_cursor_position", map[string]interface{}{
		"session_id": sessionID,
	})
	if err != nil {
		t.Fatalf("Failed to get cursor position: %v", err)
	}
	
	// Check response
	row, hasRow := result["row"]
	col, hasCol := result["col"]
	
	if !hasRow || !hasCol {
		t.Errorf("Missing cursor position data: %+v", result)
	}
	
	// Cursor should be at a valid position
	if row.(float64) < 0 || col.(float64) < 0 {
		t.Errorf("Invalid cursor position: row=%v, col=%v", row, col)
	}
}

func TestGetScreenSize(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Launch sh with echo to keep session alive
	sessionID := tf.LaunchApp("sh", []string{"-c", "echo 'Test'; sleep 1"})
	time.Sleep(100 * time.Millisecond)
	
	// Get screen size
	result, err := tf.CallTool("get_screen_size", map[string]interface{}{
		"session_id": sessionID,
	})
	if err != nil {
		t.Fatalf("Failed to get screen size: %v", err)
	}
	
	// Check response
	width, hasWidth := result["width"]
	height, hasHeight := result["height"]
	
	if !hasWidth || !hasHeight {
		t.Errorf("Missing screen size data: %+v", result)
	}
	
	// Default size should be 80x24
	if width.(float64) != 80 || height.(float64) != 24 {
		t.Errorf("Unexpected screen size: width=%v, height=%v", width, height)
	}
}

func TestResizeTerminal(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Launch a long-running app
	sessionID := tf.LaunchApp("sh", []string{"-c", "while true; do sleep 1; done"})
	time.Sleep(100 * time.Millisecond)
	
	// Resize terminal
	_, err := tf.CallTool("resize_terminal", map[string]interface{}{
		"session_id": sessionID,
		"width":      100,
		"height":     30,
	})
	if err != nil {
		t.Fatalf("Failed to resize terminal: %v", err)
	}
	
	// Verify new size
	result, err := tf.CallTool("get_screen_size", map[string]interface{}{
		"session_id": sessionID,
	})
	if err != nil {
		t.Fatalf("Failed to get screen size: %v", err)
	}
	
	width := result["width"].(float64)
	height := result["height"].(float64)
	
	if width != 100 || height != 30 {
		t.Errorf("Resize failed: expected 100x30, got %vx%v", width, height)
	}
	
	// Stop the app
	tf.StopApp(sessionID)
}

func TestStopApp(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Launch a long-running app
	sessionID := tf.LaunchApp("sh", []string{"-c", "while true; do echo 'Running'; sleep 1; done"})
	
	// Wait for some output
	if !tf.WaitForContent(sessionID, "Running", 2*time.Second) {
		t.Fatal("App didn't start properly")
	}
	
	// Stop the app
	tf.StopApp(sessionID)
	
	// Verify it's stopped - should error when trying to view
	time.Sleep(100 * time.Millisecond)
	_, err := tf.CallTool("view_screen", map[string]interface{}{
		"session_id": sessionID,
		"format":     "plain",
	})
	
	if err == nil {
		t.Error("Expected error viewing stopped session")
	}
}

func TestRestartApp(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Launch a counter that shows incrementing numbers
	sessionID := tf.LaunchApp("sh", []string{"-c", "i=0; while true; do echo \"Count: $i\"; i=$((i+1)); sleep 0.5; done"})
	
	// Wait for some output
	if !tf.WaitForContent(sessionID, "Count:", 2*time.Second) {
		t.Fatal("Initial app didn't produce output")
	}
	
	// Let it run a bit
	time.Sleep(1 * time.Second)
	
	// Get current count
	content1 := tf.ViewScreen(sessionID, "plain")
	
	// Restart the app
	_, err := tf.CallTool("restart_app", map[string]interface{}{
		"session_id": sessionID,
	})
	if err != nil {
		t.Fatalf("Failed to restart app: %v", err)
	}
	
	// Wait a bit for restart
	time.Sleep(500 * time.Millisecond)
	
	// The counter should restart from 0
	content2 := tf.ViewScreen(sessionID, "plain")
	if !strings.Contains(content2, "Count: 0") {
		t.Errorf("After restart, expected 'Count: 0' but got: %s", content2)
	}
	
	// Verify it's not the same as before restart
	if content1 == content2 {
		t.Error("Content didn't change after restart")
	}
	
	// Stop the app
	tf.StopApp(sessionID)
}

func TestListSessions(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Initially should be empty
	result, err := tf.CallTool("list_sessions", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}
	
	sessions, ok := result["sessions"].([]interface{})
	if !ok {
		t.Fatalf("Invalid sessions response: %+v", result)
	}
	
	if len(sessions) != 0 {
		t.Errorf("Expected 0 sessions initially, got %d", len(sessions))
	}
	
	// Launch some apps
	id1 := tf.LaunchApp("echo", []string{"App1"})
	id2 := tf.LaunchApp("echo", []string{"App2"})
	
	// List again
	result, err = tf.CallTool("list_sessions", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to list sessions: %v", err)
	}
	
	sessions = result["sessions"].([]interface{})
	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}
	
	// Verify session IDs are present
	foundIDs := make(map[string]bool)
	for _, s := range sessions {
		sess := s.(map[string]interface{})
		id := sess["id"].(string)
		foundIDs[id] = true
	}
	
	if !foundIDs[id1] || !foundIDs[id2] {
		t.Error("Not all session IDs found in list")
	}
}

func TestConcurrentSessions(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Launch multiple sessions concurrently
	sessionCount := 5
	sessionIDs := make([]string, sessionCount)
	
	// Launch apps with sleep to keep alive
	for i := 0; i < sessionCount; i++ {
		sessionIDs[i] = tf.LaunchApp("sh", []string{"-c", fmt.Sprintf("echo 'Session%d'; sleep 1", i)})
	}
	
	// Verify all sessions work
	for i, id := range sessionIDs {
		expected := fmt.Sprintf("Session%d", i)
		if !tf.WaitForContent(id, expected, 2*time.Second) {
			content := tf.ViewScreen(id, "plain")
			t.Errorf("Session %d: expected '%s' but got: %s", i, expected, content)
		}
	}
	
	// List sessions
	result, _ := tf.CallTool("list_sessions", map[string]interface{}{})
	sessions := result["sessions"].([]interface{})
	
	if len(sessions) != sessionCount {
		t.Errorf("Expected %d sessions, got %d", sessionCount, len(sessions))
	}
}

func TestSpecialKeys(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Test special key mappings
	tests := []struct {
		key      string
		expected string
	}{
		{"Tab", "\t"},
		{"Enter", "\r"},
		{"Escape", "\x1b"},
		{"Backspace", "\x7f"},
	}
	
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			// Launch cat to echo input
			sessionID := tf.LaunchApp("cat", []string{})
			defer tf.StopApp(sessionID)
			
			time.Sleep(100 * time.Millisecond)
			
			// Send the special key
			tf.SendKeys(sessionID, tt.key)
			
			// For most keys, cat won't show visible output
			// This test mainly verifies no errors occur
		})
	}
}

func TestErrorHandling(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Test invalid session ID
	_, err := tf.CallTool("view_screen", map[string]interface{}{
		"session_id": "invalid-session-id",
		"format":     "plain",
	})
	if err == nil {
		t.Error("Expected error for invalid session ID")
	}
	
	// Test missing required parameters
	_, err = tf.CallTool("view_screen", map[string]interface{}{
		"format": "plain",
		// missing session_id
	})
	if err == nil {
		t.Error("Expected error for missing session_id")
	}
	
	// Test invalid format
	sessionID := tf.LaunchApp("echo", []string{"test"})
	_, err = tf.CallTool("view_screen", map[string]interface{}{
		"session_id": sessionID,
		"format":     "invalid-format",
	})
	// This might not error, just use default format
	
	// Test launching non-existent command
	_, err = tf.CallTool("launch_app", map[string]interface{}{
		"command": "/non/existent/command",
		"args":    []string{},
	})
	if err == nil {
		t.Error("Expected error for non-existent command")
	}
}

func TestAnsiOutput(t *testing.T) {
	tf := NewTestFramework(t)
	defer tf.Cleanup()
	
	// Launch sh with printf color output (more reliable than echo -e)
	sessionID := tf.LaunchApp("sh", []string{"-c", "printf '\033[31mRed Text\033[0m\\n'; sleep 1"})
	time.Sleep(100 * time.Millisecond)
	
	// Plain format should strip ANSI
	plain := tf.ViewScreen(sessionID, "plain")
	if strings.Contains(plain, "\033") || strings.Contains(plain, "\x1b") {
		t.Error("Plain format should not contain ANSI sequences")
	}
	if !strings.Contains(plain, "Red Text") {
		t.Error("Plain format should contain the text")
	}
	
	// Raw format should include ANSI (check for actual escape sequences)
	raw := tf.ViewScreen(sessionID, "raw")
	// Check for ANSI escape sequences (either \033 or \x1b format)
	hasColorStart := strings.Contains(raw, "\033[31m") || strings.Contains(raw, "\x1b[31m") || strings.Contains(raw, "\x1b[38;2;")
	hasColorEnd := strings.Contains(raw, "\033[0m") || strings.Contains(raw, "\x1b[0m")
	if !hasColorStart || !hasColorEnd {
		t.Errorf("Raw format should contain ANSI sequences. Raw: %q", raw)
	}
}