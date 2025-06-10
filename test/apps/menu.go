package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var menuItems = []string{
	"Show System Info",
	"Test Cursor Movement",
	"Test Colors and Attributes",
	"Test Box Drawing",
	"Test Input Echo",
	"Clear Screen",
	"Exit",
}

var selectedIndex = 0

func main() {
	// Enable raw mode to capture arrow keys
	oldState, err := makeRaw()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting raw mode: %v\n", err)
		return
	}
	defer restore(oldState)

	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h") // Show cursor on exit

	clearScreen()
	
	for {
		drawMenu()
		
		// Read single character
		var buf [3]byte
		n, _ := os.Stdin.Read(buf[:])
		
		if n == 1 {
			switch buf[0] {
			case 'q', 'Q', 27: // q, Q, or ESC
				clearScreen()
				return
			case 13, 10: // Enter
				handleSelection()
			case 'k', 'K': // Up
				moveUp()
			case 'j', 'J': // Down
				moveDown()
			}
		} else if n == 3 {
			// Check for arrow keys (ESC [ A/B/C/D)
			if buf[0] == 27 && buf[1] == '[' {
				switch buf[2] {
				case 'A': // Up arrow
					moveUp()
				case 'B': // Down arrow
					moveDown()
				}
			}
		}
	}
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func drawMenu() {
	// Move cursor to top
	fmt.Print("\033[H")
	
	// Draw header
	fmt.Println("\033[1;36m╔════════════════════════════════════════╗\033[0m")
	fmt.Println("\033[1;36m║      Terminal Test Menu System         ║\033[0m")
	fmt.Println("\033[1;36m╠════════════════════════════════════════╣\033[0m")
	fmt.Println("\033[1;36m║  Use ↑/↓ or j/k to navigate           ║\033[0m")
	fmt.Println("\033[1;36m║  Press Enter to select                ║\033[0m")
	fmt.Println("\033[1;36m║  Press q or ESC to quit               ║\033[0m")
	fmt.Println("\033[1;36m╚════════════════════════════════════════╝\033[0m")
	fmt.Println()

	// Draw menu items
	for i, item := range menuItems {
		if i == selectedIndex {
			// Highlighted item
			fmt.Printf("\033[7m  ▶ %s  \033[0m\n", item)
		} else {
			fmt.Printf("    %s\n", item)
		}
	}
}

func moveUp() {
	if selectedIndex > 0 {
		selectedIndex--
	} else {
		selectedIndex = len(menuItems) - 1
	}
}

func moveDown() {
	if selectedIndex < len(menuItems)-1 {
		selectedIndex++
	} else {
		selectedIndex = 0
	}
}

func handleSelection() {
	clearScreen()
	
	switch selectedIndex {
	case 0: // System Info
		showSystemInfo()
	case 1: // Cursor Movement
		testCursorMovement()
	case 2: // Colors
		testColors()
	case 3: // Box Drawing
		testBoxDrawing()
	case 4: // Input Echo
		testInputEcho()
	case 5: // Clear Screen
		clearScreen()
		fmt.Println("Screen cleared!")
	case 6: // Exit
		os.Exit(0)
	}
	
	fmt.Println("\nPress any key to continue...")
	var buf [1]byte
	os.Stdin.Read(buf[:])
}

func showSystemInfo() {
	fmt.Println("\033[1;33mSystem Information:\033[0m")
	fmt.Printf("OS: %s\n", runtime.GOOS)
	fmt.Printf("Architecture: %s\n", runtime.GOARCH)
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("Terminal Size: Run 'stty size' to check\n")
}

func testCursorMovement() {
	fmt.Println("\033[1;33mCursor Movement Test:\033[0m")
	
	// Save cursor position
	fmt.Print("\033[s")
	
	// Move cursor around
	fmt.Print("\033[5;10HCursor at (5,10)")
	fmt.Print("\033[10;20HCursor at (10,20)")
	fmt.Print("\033[15;30HCursor at (15,30)")
	
	// Restore cursor position
	fmt.Print("\033[u")
	fmt.Println("\nCursor restored to original position")
}

func testColors() {
	fmt.Println("\033[1;33mColor and Attribute Test:\033[0m\n")
	
	// Foreground colors
	fmt.Println("Foreground colors:")
	for i := 30; i <= 37; i++ {
		fmt.Printf("\033[%dmColor %d ", i, i)
	}
	fmt.Println("\033[0m")
	
	// Background colors
	fmt.Println("\nBackground colors:")
	for i := 40; i <= 47; i++ {
		fmt.Printf("\033[%dmColor %d ", i, i)
	}
	fmt.Println("\033[0m")
	
	// Attributes
	fmt.Println("\nAttributes:")
	fmt.Println("\033[1mBold\033[0m")
	fmt.Println("\033[3mItalic\033[0m")
	fmt.Println("\033[4mUnderline\033[0m")
	fmt.Println("\033[5mBlink\033[0m")
	fmt.Println("\033[7mReverse\033[0m")
	
	// 256 colors
	fmt.Println("\n256 Color palette:")
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			color := i*16 + j
			fmt.Printf("\033[48;5;%dm %3d ", color, color)
		}
		fmt.Println("\033[0m")
	}
}

func testBoxDrawing() {
	fmt.Println("\033[1;33mBox Drawing Test:\033[0m\n")
	
	// Single line box
	fmt.Println("┌─────────────────┐")
	fmt.Println("│ Single Line Box │")
	fmt.Println("└─────────────────┘")
	
	// Double line box
	fmt.Println("\n╔═════════════════╗")
	fmt.Println("║ Double Line Box ║")
	fmt.Println("╚═════════════════╝")
	
	// Mixed box
	fmt.Println("\n╔═══════┬═════════╗")
	fmt.Println("║ Cell 1 │ Cell 2  ║")
	fmt.Println("╟───────┼─────────╢")
	fmt.Println("║ Cell 3 │ Cell 4  ║")
	fmt.Println("╚═══════╧═════════╝")
}

func testInputEcho() {
	fmt.Println("\033[1;33mInput Echo Test:\033[0m")
	fmt.Println("Type something and press Enter (type 'done' to finish):")
	
	// Temporarily restore cooked mode for line input
	restore(nil)
	defer makeRaw()
	
	for {
		fmt.Print("> ")
		var input string
		fmt.Scanln(&input)
		
		if input == "done" {
			break
		}
		
		fmt.Printf("You typed: \033[1;32m%s\033[0m\n", input)
	}
}

// Terminal mode functions (platform-specific implementations would go here)
func makeRaw() (interface{}, error) {
	// This is a simplified version
	// In a real implementation, you'd use termios on Unix or Windows Console API
	exec.Command("stty", "-echo", "raw").Run()
	return nil, nil
}

func restore(oldState interface{}) error {
	// Restore terminal state
	exec.Command("stty", "echo", "-raw").Run()
	return nil
}