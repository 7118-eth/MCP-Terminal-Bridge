package main

import (
	"fmt"
	"os"
	"strings"
)

// Simple vim-like editor for testing terminal interactions
// This is a simplified version to test MCP Terminal capabilities

type Mode int

const (
	NormalMode Mode = iota
	InsertMode
	CommandMode
)

type Editor struct {
	lines       []string
	cursorX     int
	cursorY     int
	mode        Mode
	filename    string
	modified    bool
	message     string
	screenRows  int
	screenCols  int
	topLine     int // Top line displayed on screen
}

func NewEditor() *Editor {
	return &Editor{
		lines:      []string{""},
		cursorX:    0,
		cursorY:    0,
		mode:       NormalMode,
		screenRows: 24,
		screenCols: 80,
		topLine:    0,
	}
}

func (e *Editor) loadFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		// Create new file
		e.lines = []string{""}
		e.filename = filename
		e.message = fmt.Sprintf("\"%s\" [New File]", filename)
		return nil
	}
	
	e.filename = filename
	e.lines = strings.Split(string(content), "\n")
	if len(e.lines) == 0 {
		e.lines = []string{""}
	}
	// Remove last empty line if file ends with newline
	if len(e.lines) > 1 && e.lines[len(e.lines)-1] == "" {
		e.lines = e.lines[:len(e.lines)-1]
	}
	e.message = fmt.Sprintf("\"%s\" %d lines", filename, len(e.lines))
	return nil
}

func (e *Editor) saveFile() error {
	if e.filename == "" {
		e.message = "No filename"
		return fmt.Errorf("no filename")
	}
	
	content := strings.Join(e.lines, "\n")
	err := os.WriteFile(e.filename, []byte(content), 0644)
	if err != nil {
		e.message = fmt.Sprintf("Error: %s", err.Error())
		return err
	}
	
	e.modified = false
	e.message = fmt.Sprintf("\"%s\" written", e.filename)
	return nil
}

func (e *Editor) clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func (e *Editor) moveCursor(x, y int) {
	fmt.Printf("\033[%d;%dH", y+1, x+1)
}

func (e *Editor) draw() {
	e.clearScreen()
	
	// Draw file content
	for i := 0; i < e.screenRows-2; i++ {
		lineNum := e.topLine + i
		if lineNum < len(e.lines) {
			line := e.lines[lineNum]
			if len(line) > e.screenCols {
				line = line[:e.screenCols]
			}
			fmt.Print(line)
		} else {
			fmt.Print("~")
		}
		fmt.Print("\033[K") // Clear to end of line
		if i < e.screenRows-3 {
			fmt.Print("\n")
		}
	}
	
	// Status line
	fmt.Print("\n\033[7m") // Reverse video
	
	var modeStr string
	switch e.mode {
	case NormalMode:
		modeStr = ""
	case InsertMode:
		modeStr = "-- INSERT --"
	case CommandMode:
		modeStr = ":"
	}
	
	filename := e.filename
	if filename == "" {
		filename = "[No Name]"
	}
	if e.modified {
		filename += " [+]"
	}
	
	status := fmt.Sprintf(" %s", filename)
	if modeStr != "" {
		status = fmt.Sprintf(" %s %s", modeStr, filename)
	}
	
	// Pad status line
	for len(status) < e.screenCols {
		status += " "
	}
	if len(status) > e.screenCols {
		status = status[:e.screenCols]
	}
	
	fmt.Print(status)
	fmt.Print("\033[0m") // Reset attributes
	
	// Message line
	fmt.Print("\n")
	if e.message != "" {
		msg := e.message
		if len(msg) > e.screenCols {
			msg = msg[:e.screenCols]
		}
		fmt.Print(msg)
	}
	fmt.Print("\033[K") // Clear to end of line
	
	// Position cursor
	displayY := e.cursorY - e.topLine + 1
	displayX := e.cursorX + 1
	if e.mode == CommandMode {
		displayY = e.screenRows
		displayX = 2 // After the ':'
	}
	e.moveCursor(displayX-1, displayY-1)
}

func (e *Editor) adjustScroll() {
	// Scroll down if cursor is below screen
	if e.cursorY >= e.topLine+e.screenRows-2 {
		e.topLine = e.cursorY - e.screenRows + 3
	}
	// Scroll up if cursor is above screen
	if e.cursorY < e.topLine {
		e.topLine = e.cursorY
	}
	if e.topLine < 0 {
		e.topLine = 0
	}
}

func (e *Editor) ensureCursorValid() {
	if e.cursorY >= len(e.lines) {
		e.cursorY = len(e.lines) - 1
	}
	if e.cursorY < 0 {
		e.cursorY = 0
	}
	
	if e.cursorY < len(e.lines) {
		lineLen := len(e.lines[e.cursorY])
		if e.cursorX > lineLen {
			e.cursorX = lineLen
		}
	}
	if e.cursorX < 0 {
		e.cursorX = 0
	}
}

func (e *Editor) insertChar(ch rune) {
	line := e.lines[e.cursorY]
	newLine := line[:e.cursorX] + string(ch) + line[e.cursorX:]
	e.lines[e.cursorY] = newLine
	e.cursorX++
	e.modified = true
}

func (e *Editor) insertNewline() {
	line := e.lines[e.cursorY]
	newLine := line[:e.cursorX]
	remainingLine := line[e.cursorX:]
	
	e.lines[e.cursorY] = newLine
	e.lines = append(e.lines[:e.cursorY+1], append([]string{remainingLine}, e.lines[e.cursorY+1:]...)...)
	
	e.cursorY++
	e.cursorX = 0
	e.modified = true
}

func (e *Editor) backspace() {
	if e.cursorX > 0 {
		line := e.lines[e.cursorY]
		newLine := line[:e.cursorX-1] + line[e.cursorX:]
		e.lines[e.cursorY] = newLine
		e.cursorX--
		e.modified = true
	} else if e.cursorY > 0 {
		// Join with previous line
		prevLine := e.lines[e.cursorY-1]
		currentLine := e.lines[e.cursorY]
		e.lines[e.cursorY-1] = prevLine + currentLine
		e.lines = append(e.lines[:e.cursorY], e.lines[e.cursorY+1:]...)
		e.cursorY--
		e.cursorX = len(prevLine)
		e.modified = true
	}
}

func (e *Editor) processNormalMode(ch byte) {
	e.message = "" // Clear message
	
	switch ch {
	case 'i':
		e.mode = InsertMode
	case 'a':
		e.mode = InsertMode
		e.cursorX++
		e.ensureCursorValid()
	case 'o':
		e.mode = InsertMode
		e.cursorX = len(e.lines[e.cursorY])
		e.insertNewline()
		e.cursorY--
		e.cursorX = 0
		e.insertNewline()
	case 'O':
		e.mode = InsertMode
		e.cursorX = 0
		e.insertNewline()
		e.cursorY--
	case ':':
		e.mode = CommandMode
	case 'h':
		e.cursorX--
		e.ensureCursorValid()
	case 'j':
		e.cursorY++
		e.ensureCursorValid()
	case 'k':
		e.cursorY--
		e.ensureCursorValid()
	case 'l':
		e.cursorX++
		e.ensureCursorValid()
	case '0':
		e.cursorX = 0
	case '$':
		if e.cursorY < len(e.lines) {
			e.cursorX = len(e.lines[e.cursorY])
		}
	case 'g':
		// Simple gg implementation (go to top)
		e.cursorY = 0
		e.cursorX = 0
	case 'G':
		// Go to bottom
		e.cursorY = len(e.lines) - 1
		e.cursorX = 0
		e.ensureCursorValid()
	case 'x':
		// Delete character
		if e.cursorY < len(e.lines) {
			line := e.lines[e.cursorY]
			if e.cursorX < len(line) {
				newLine := line[:e.cursorX] + line[e.cursorX+1:]
				e.lines[e.cursorY] = newLine
				e.modified = true
			}
		}
	case 'd':
		// Simple dd implementation (delete line)
		if len(e.lines) > 1 {
			e.lines = append(e.lines[:e.cursorY], e.lines[e.cursorY+1:]...)
			e.modified = true
			e.ensureCursorValid()
		} else {
			e.lines[0] = ""
			e.cursorX = 0
			e.modified = true
		}
	}
	
	e.adjustScroll()
}

func (e *Editor) processInsertMode(ch byte) {
	switch ch {
	case 27: // Escape
		e.mode = NormalMode
		if e.cursorX > 0 {
			e.cursorX--
		}
		e.ensureCursorValid()
	case 13, 10: // Enter
		e.insertNewline()
	case 8, 127: // Backspace/Delete
		e.backspace()
	default:
		if ch >= 32 && ch < 127 { // Printable characters
			e.insertChar(rune(ch))
		}
	}
	
	e.adjustScroll()
}

func (e *Editor) processCommandMode(ch byte) {
	switch ch {
	case 27: // Escape
		e.mode = NormalMode
	case 13, 10: // Enter
		// For simplicity, just handle :q and :w
		if ch == 'q' {
			os.Exit(0)
		} else if ch == 'w' {
			e.saveFile()
		}
		e.mode = NormalMode
	case 'q':
		if !e.modified {
			os.Exit(0)
		} else {
			e.message = "No write since last change (use :q! to override)"
			e.mode = NormalMode
		}
	case 'w':
		e.saveFile()
		e.mode = NormalMode
	}
}

func main() {
	editor := NewEditor()
	
	// Load file if specified
	if len(os.Args) > 1 {
		filename := os.Args[1]
		editor.loadFile(filename)
	}
	
	// Enable raw mode for terminal
	// This is a simplified version - in a real implementation you'd use termios
	fmt.Print("\033[?25h") // Show cursor
	defer fmt.Print("\033[?25h\033[0m\033[2J\033[H") // Cleanup on exit
	
	editor.draw()
	
	// Simple main loop
	var buf [1]byte
	for {
		n, _ := os.Stdin.Read(buf[:])
		if n > 0 {
			ch := buf[0]
			
			switch editor.mode {
			case NormalMode:
				editor.processNormalMode(ch)
			case InsertMode:
				editor.processInsertMode(ch)
			case CommandMode:
				editor.processCommandMode(ch)
			}
			
			editor.draw()
		}
	}
}