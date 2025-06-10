package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

func main() {
	fmt.Println("Terminal Progress Bar and Animation Test")
	fmt.Println("======================================")
	fmt.Println()

	// Test 1: Simple progress bar
	fmt.Println("1. Simple Progress Bar:")
	simpleProgressBar()
	fmt.Println()

	// Test 2: Colored progress bar
	fmt.Println("2. Colored Progress Bar:")
	coloredProgressBar()
	fmt.Println()

	// Test 3: Spinner animation
	fmt.Println("3. Spinner Animation:")
	spinnerAnimation()
	fmt.Println()

	// Test 4: Multi-line progress
	fmt.Println("4. Multi-line Progress:")
	multiLineProgress()
	fmt.Println()

	// Test 5: Wave animation
	fmt.Println("5. Wave Animation:")
	waveAnimation()
	fmt.Println()

	fmt.Println("All tests completed!")
}

func simpleProgressBar() {
	width := 50
	for i := 0; i <= 100; i++ {
		filled := int(float64(width) * float64(i) / 100.0)
		bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
		fmt.Printf("\r[%s] %3d%%", bar, i)
		time.Sleep(20 * time.Millisecond)
	}
	fmt.Println()
}

func coloredProgressBar() {
	width := 50
	for i := 0; i <= 100; i++ {
		filled := int(float64(width) * float64(i) / 100.0)
		
		// Color based on percentage
		color := "\033[32m" // Green
		if i < 33 {
			color = "\033[31m" // Red
		} else if i < 66 {
			color = "\033[33m" // Yellow
		}
		
		bar := color + strings.Repeat("█", filled) + "\033[0m" + strings.Repeat("░", width-filled)
		fmt.Printf("\r[%s] %3d%%", bar, i)
		time.Sleep(20 * time.Millisecond)
	}
	fmt.Println()
}

func spinnerAnimation() {
	spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	
	for i := 0; i < 50; i++ {
		spinner := spinners[i%len(spinners)]
		fmt.Printf("\r\033[36m%s\033[0m Loading... %d%%", spinner, i*2)
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Printf("\r\033[32m✓\033[0m Complete!     \n")
}

func multiLineProgress() {
	tasks := []string{
		"Downloading files",
		"Processing data",
		"Optimizing results",
		"Generating report",
	}
	
	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")
	
	// Save cursor position
	fmt.Print("\033[s")
	
	// Print all tasks
	for i, task := range tasks {
		fmt.Printf("%d. %s: [ ] 0%%\n", i+1, task)
	}
	
	// Update each task
	for i := range tasks {
		for j := 0; j <= 100; j += 5 {
			// Move cursor to specific line
			fmt.Printf("\033[u\033[%dB", i)
			
			// Update progress
			width := 20
			filled := int(float64(width) * float64(j) / 100.0)
			bar := strings.Repeat("=", filled) + strings.Repeat("-", width-filled)
			
			if j == 100 {
				fmt.Printf("\033[2K\r%d. %s: \033[32m[%s] %d%% ✓\033[0m", i+1, tasks[i], bar, j)
			} else {
				fmt.Printf("\033[2K\r%d. %s: [%s] %d%%", i+1, tasks[i], bar, j)
			}
			
			time.Sleep(50 * time.Millisecond)
		}
	}
	
	// Move cursor to end
	fmt.Printf("\033[u\033[%dB\n", len(tasks))
}

func waveAnimation() {
	width := 60
	height := 10
	frames := 50
	
	// Hide cursor
	fmt.Print("\033[?25l")
	defer fmt.Print("\033[?25h")
	
	// Save cursor position
	fmt.Print("\033[s")
	
	// Create space for animation
	for i := 0; i < height; i++ {
		fmt.Println()
	}
	
	for frame := 0; frame < frames; frame++ {
		// Move cursor back to start
		fmt.Print("\033[u")
		
		// Draw wave
		for y := 0; y < height; y++ {
			line := ""
			for x := 0; x < width; x++ {
				// Calculate wave height at this position
				waveHeight := int(5 + 4*math.Sin(float64(x)*0.2+float64(frame)*0.2))
				
				if y == height-waveHeight-1 {
					// Wave crest
					line += "\033[36m~\033[0m"
				} else if y > height-waveHeight-1 {
					// Water
					line += "\033[34m≈\033[0m"
				} else {
					// Air
					line += " "
				}
			}
			fmt.Printf("\033[2K%s\n", line)
		}
		
		time.Sleep(100 * time.Millisecond)
	}
}