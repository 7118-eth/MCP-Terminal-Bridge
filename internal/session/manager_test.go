package session

import (
	"fmt"
	"testing"
	"time"
	
	"github.com/bioharz/mcp-terminal-tester/internal/utils"
)

func TestManager_CreateSession(t *testing.T) {
	// Initialize logger for tests
	utils.InitLogger()
	
	manager := NewManager()
	
	// Test creating a session
	sess, err := manager.CreateSession("echo", []string{"test"}, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	
	if sess == nil {
		t.Fatal("Session is nil")
	}
	
	if sess.Command != "echo" {
		t.Errorf("Expected command 'echo', got '%s'", sess.Command)
	}
	
	if len(sess.Args) != 1 || sess.Args[0] != "test" {
		t.Errorf("Expected args ['test'], got %v", sess.Args)
	}
	
	// Verify session is in manager
	retrieved, err := manager.GetSession(sess.ID)
	if err != nil {
		t.Errorf("Failed to retrieve session: %v", err)
	}
	
	if retrieved.ID != sess.ID {
		t.Error("Retrieved session has different ID")
	}
	
	// Clean up
	manager.RemoveSession(sess.ID)
}

func TestManager_MaxSessions(t *testing.T) {
	utils.InitLogger()
	manager := NewManager()
	manager.maxSessions = 3 // Set low limit for testing
	
	// Create sessions up to limit
	var sessions []*Session
	for i := 0; i < 3; i++ {
		sess, err := manager.CreateSession("echo", []string{}, nil)
		if err != nil {
			t.Fatalf("Failed to create session %d: %v", i, err)
		}
		sessions = append(sessions, sess)
	}
	
	// Try to create one more - should fail
	_, err := manager.CreateSession("echo", []string{}, nil)
	if err == nil {
		t.Error("Expected error when exceeding max sessions")
	}
	
	// Remove one session
	manager.RemoveSession(sessions[0].ID)
	
	// Now we should be able to create another
	_, err = manager.CreateSession("echo", []string{}, nil)
	if err != nil {
		t.Errorf("Should be able to create session after removing one: %v", err)
	}
	
	// Clean up
	for _, sess := range sessions[1:] {
		manager.RemoveSession(sess.ID)
	}
}

func TestManager_GetSession_NotFound(t *testing.T) {
	utils.InitLogger()
	manager := NewManager()
	
	_, err := manager.GetSession("non-existent-id")
	if err == nil {
		t.Error("Expected error for non-existent session")
	}
}

func TestManager_RemoveSession(t *testing.T) {
	utils.InitLogger()
	manager := NewManager()
	
	// Create a session
	sess, err := manager.CreateSession("echo", []string{}, nil)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}
	
	// Remove it
	err = manager.RemoveSession(sess.ID)
	if err != nil {
		t.Errorf("Failed to remove session: %v", err)
	}
	
	// Verify it's gone
	_, err = manager.GetSession(sess.ID)
	if err == nil {
		t.Error("Session should not exist after removal")
	}
	
	// Try to remove again - should error
	err = manager.RemoveSession(sess.ID)
	if err == nil {
		t.Error("Expected error when removing non-existent session")
	}
}

func TestManager_ListSessions(t *testing.T) {
	utils.InitLogger()
	manager := NewManager()
	
	// Initially empty
	sessions := manager.ListSessions()
	if len(sessions) != 0 {
		t.Error("Expected no sessions initially")
	}
	
	// Create some sessions
	sess1, _ := manager.CreateSession("echo", []string{"1"}, nil)
	sess2, _ := manager.CreateSession("echo", []string{"2"}, nil)
	
	// List should have 2
	sessions = manager.ListSessions()
	if len(sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(sessions))
	}
	
	// Verify session info
	foundIDs := make(map[string]bool)
	for _, info := range sessions {
		foundIDs[info.ID] = true
		if info.Command != "echo" {
			t.Errorf("Expected command 'echo', got '%s'", info.Command)
		}
	}
	
	if !foundIDs[sess1.ID] || !foundIDs[sess2.ID] {
		t.Error("Not all sessions found in list")
	}
	
	// Clean up
	manager.RemoveSession(sess1.ID)
	manager.RemoveSession(sess2.ID)
}

func TestManager_CleanupIdleSessions(t *testing.T) {
	utils.InitLogger()
	manager := NewManager()
	manager.sessionTimeout = 100 * time.Millisecond // Short timeout for testing
	
	// Create sessions
	sess1, _ := manager.CreateSession("echo", []string{}, nil)
	sess2, _ := manager.CreateSession("echo", []string{}, nil)
	
	// Make sess1 idle by setting LastActive in the past
	sess1.mu.Lock()
	sess1.LastActive = time.Now().Add(-200 * time.Millisecond)
	sess1.mu.Unlock()
	
	// Run cleanup
	manager.CleanupIdleSessions()
	
	// sess1 should be gone, sess2 should remain
	_, err := manager.GetSession(sess1.ID)
	if err == nil {
		t.Error("Idle session should have been cleaned up")
	}
	
	_, err = manager.GetSession(sess2.ID)
	if err != nil {
		t.Error("Active session should not have been cleaned up")
	}
	
	// Clean up
	manager.RemoveSession(sess2.ID)
}

func TestManager_ConcurrentAccess(t *testing.T) {
	manager := NewManager()
	done := make(chan bool)
	
	// Concurrent creates
	for i := 0; i < 5; i++ {
		go func(id int) {
			_, err := manager.CreateSession("echo", []string{fmt.Sprintf("%d", id)}, nil)
			if err != nil {
				t.Errorf("Concurrent create %d failed: %v", id, err)
			}
			done <- true
		}(i)
	}
	
	// Wait for creates
	for i := 0; i < 5; i++ {
		<-done
	}
	
	// Verify all sessions created
	sessions := manager.ListSessions()
	if len(sessions) != 5 {
		t.Errorf("Expected 5 sessions, got %d", len(sessions))
	}
	
	// Concurrent reads
	for _, sess := range sessions {
		go func(id string) {
			_, err := manager.GetSession(id)
			if err != nil {
				t.Errorf("Concurrent get failed: %v", err)
			}
			done <- true
		}(sess.ID)
	}
	
	// Wait for reads
	for i := 0; i < 5; i++ {
		<-done
	}
	
	// Clean up
	for _, sess := range sessions {
		manager.RemoveSession(sess.ID)
	}
}