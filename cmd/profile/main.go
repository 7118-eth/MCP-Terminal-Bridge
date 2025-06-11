package main

import (
	"log"
	"os"
	"runtime/pprof"

	"github.com/bioharz/mcp-terminal-tester/internal/session"
	"github.com/bioharz/mcp-terminal-tester/internal/terminal"
	"github.com/bioharz/mcp-terminal-tester/internal/utils"
)

// Simple profiling program to identify hot paths
func main() {
	// Initialize logger
	utils.InitLogger()
	
	// Create CPU profile
	cpuFile, err := os.Create("cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer cpuFile.Close()

	if err := pprof.StartCPUProfile(cpuFile); err != nil {
		log.Fatal(err)
	}
	defer pprof.StopCPUProfile()

	// Test scenarios that represent common usage patterns
	testScenario()

	// Create memory profile
	memFile, err := os.Create("mem.prof")
	if err != nil {
		log.Fatal(err)
	}
	defer memFile.Close()

	if err := pprof.WriteHeapProfile(memFile); err != nil {
		log.Fatal(err)
	}

	log.Println("Profiling completed. Run:")
	log.Println("go tool pprof cpu.prof")
	log.Println("go tool pprof mem.prof")
}

func testScenario() {
	// Test 1: Session creation and management
	manager := session.NewManager()

	// Create multiple sessions
	sessions := make([]*session.Session, 10)
	for i := 0; i < 10; i++ {
		sess, err := manager.CreateSession("echo", []string{"test"}, nil)
		if err != nil {
			log.Printf("Failed to create session %d: %v", i, err)
			continue
		}
		sessions[i] = sess
	}

	// Test 2: Screen buffer operations (hot path)
	buffer := terminal.NewScreenBuffer(80, 24)
	
	// Simulate heavy buffer usage
	for i := 0; i < 1000; i++ {
		// Write data to buffer (common operation)
		data := []byte("This is test data with ANSI sequences \033[31mRed\033[0m\n")
		buffer.Write(data)
		
		// Render buffer (very common operation)
		buffer.Render("plain")
		buffer.Render("raw")
		
		// Move cursor around
		buffer.MoveCursor(i%80, i%24)
		
		// Clear operations
		if i%100 == 0 {
			buffer.Clear()
		}
	}

	// Test 3: ANSI parsing (hot path)
	parser := terminal.NewANSIParser(buffer)
	
	// Test various ANSI sequences
	sequences := []string{
		"\033[31mRed text\033[0m",
		"\033[1;32mBold green\033[0m",
		"\033[38;5;196mBright red\033[0m",
		"\033[2J\033[H",
		"\033[10;20H",
		"\033[K",
		"\033[s\033[u",
	}
	
	for i := 0; i < 1000; i++ {
		for _, seq := range sequences {
			parser.Parse([]byte(seq))
		}
	}

	// Test 4: Session operations
	for i := 0; i < 100; i++ {
		// List sessions (common operation)
		manager.ListSessions()
		
		// Get sessions (very common)
		for _, sess := range sessions {
			if sess != nil {
				manager.GetSession(sess.ID)
			}
		}
	}

	// Cleanup
	for _, sess := range sessions {
		if sess != nil {
			manager.RemoveSession(sess.ID)
		}
	}
}