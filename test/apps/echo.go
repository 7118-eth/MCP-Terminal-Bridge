package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("Echo Test Application")
	fmt.Println("Type 'quit' or 'exit' to stop")
	fmt.Println("Type 'clear' to clear the screen")
	fmt.Println("Type 'color' to test ANSI colors")
	fmt.Println("------------------------------")

	scanner := bufio.NewScanner(os.Stdin)
	prompt := "> "

	for {
		fmt.Print(prompt)
		
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		
		// Handle special commands
		switch strings.ToLower(strings.TrimSpace(input)) {
		case "quit", "exit":
			fmt.Println("Goodbye!")
			return
		case "clear":
			// Clear screen using ANSI escape sequence
			fmt.Print("\033[2J\033[H")
			fmt.Println("Screen cleared!")
		case "color":
			testColors()
		case "help":
			showHelp()
		default:
			// Echo the input with some formatting
			fmt.Printf("You typed: %s\n", input)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}

func testColors() {
	fmt.Println("\nColor Test:")
	
	// Basic colors
	fmt.Println("\033[31mRed text\033[0m")
	fmt.Println("\033[32mGreen text\033[0m")
	fmt.Println("\033[33mYellow text\033[0m")
	fmt.Println("\033[34mBlue text\033[0m")
	fmt.Println("\033[35mMagenta text\033[0m")
	fmt.Println("\033[36mCyan text\033[0m")
	fmt.Println("\033[37mWhite text\033[0m")
	
	// Bold
	fmt.Println("\033[1mBold text\033[0m")
	
	// Underline
	fmt.Println("\033[4mUnderlined text\033[0m")
	
	// Background colors
	fmt.Println("\033[41mRed background\033[0m")
	fmt.Println("\033[42mGreen background\033[0m")
	
	// Combined
	fmt.Println("\033[1;33;44mBold yellow on blue\033[0m")
	
	fmt.Println()
}

func showHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  quit/exit - Exit the program")
	fmt.Println("  clear     - Clear the screen")
	fmt.Println("  color     - Show color test")
	fmt.Println("  help      - Show this help")
	fmt.Println("  <text>    - Echo the text back")
	fmt.Println()
}