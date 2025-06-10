package terminal

import (
	"bytes"
	"strconv"
	"strings"
)

// cursorState holds saved cursor position and attributes
type cursorState struct {
	x, y         int
	fg, bg       Color
	attrs        Attributes
}

type ANSIParser struct {
	buffer       *ScreenBuffer
	state        parserState
	escapeBuffer bytes.Buffer
	currentFG    Color
	currentBG    Color
	currentAttrs Attributes
	savedCursor  *cursorState // Per-parser cursor save state
}

type parserState int

const (
	stateNormal parserState = iota
	stateEscape
	stateCSI
	stateOSC
	stateDCS     // Device Control String
	stateCharset // Character set selection
)

func NewANSIParser(buffer *ScreenBuffer) *ANSIParser {
	return &ANSIParser{
		buffer:    buffer,
		state:     stateNormal,
		currentFG: Color{Default: true},
		currentBG: Color{Default: true},
	}
}

func (p *ANSIParser) Parse(data []byte) {
	for _, b := range data {
		switch p.state {
		case stateNormal:
			p.handleNormal(b)
		case stateEscape:
			p.handleEscape(b)
		case stateCSI:
			p.handleCSI(b)
		case stateOSC:
			p.handleOSC(b)
		case stateDCS:
			p.handleDCS(b)
		case stateCharset:
			p.handleCharset(b)
		}
	}
}

func (p *ANSIParser) handleNormal(b byte) {
	switch b {
	case 0x1B: // ESC
		p.state = stateEscape
		p.escapeBuffer.Reset()
	case '\r': // Carriage return
		p.buffer.MoveCursor(0, p.buffer.cursorY)
	case '\n': // Line feed
		p.buffer.cursorY++
		if p.buffer.cursorY >= p.buffer.height {
			p.buffer.ScrollUp()
			p.buffer.cursorY = p.buffer.height - 1
		}
	case '\t': // Tab
		// Move to next tab stop (every 8 columns)
		newX := ((p.buffer.cursorX / 8) + 1) * 8
		if newX >= p.buffer.width {
			newX = p.buffer.width - 1
		}
		p.buffer.MoveCursor(newX, p.buffer.cursorY)
	case '\b': // Backspace
		if p.buffer.cursorX > 0 {
			p.buffer.MoveCursor(p.buffer.cursorX-1, p.buffer.cursorY)
		}
	default:
		if b >= 0x20 && b < 0x7F { // Printable ASCII
			p.buffer.SetCell(p.buffer.cursorX, p.buffer.cursorY, rune(b), p.currentFG, p.currentBG, p.currentAttrs)
			p.buffer.cursorX++
			if p.buffer.cursorX >= p.buffer.width {
				p.buffer.cursorX = 0
				p.buffer.cursorY++
				if p.buffer.cursorY >= p.buffer.height {
					p.buffer.ScrollUp()
					p.buffer.cursorY = p.buffer.height - 1
				}
			}
		}
	}
}

func (p *ANSIParser) handleEscape(b byte) {
	switch b {
	case '[':
		p.state = stateCSI
		p.escapeBuffer.Reset()
	case ']':
		p.state = stateOSC
		p.escapeBuffer.Reset()
	case 'P':
		p.state = stateDCS
		p.escapeBuffer.Reset()
	case '(', ')', '*', '+': // Character set selection
		p.state = stateCharset
		p.escapeBuffer.WriteByte(b)
	case 'c': // RIS - Reset to Initial State
		p.buffer.Clear()
		p.currentFG = Color{Default: true}
		p.currentBG = Color{Default: true}
		p.currentAttrs = Attributes{}
		p.state = stateNormal
	case 'D': // IND - Index (move down one line)
		p.buffer.cursorY++
		if p.buffer.cursorY >= p.buffer.height {
			p.buffer.ScrollUp()
			p.buffer.cursorY = p.buffer.height - 1
		}
		p.state = stateNormal
	case 'M': // RI - Reverse Index (move up one line)
		if p.buffer.cursorY > 0 {
			p.buffer.cursorY--
		} else {
			p.buffer.ScrollDown()
		}
		p.state = stateNormal
	case 'E': // NEL - Next Line
		p.buffer.cursorX = 0
		p.buffer.cursorY++
		if p.buffer.cursorY >= p.buffer.height {
			p.buffer.ScrollUp()
			p.buffer.cursorY = p.buffer.height - 1
		}
		p.state = stateNormal
	case '7': // DECSC - Save Cursor
		p.saveCursor()
		p.state = stateNormal
	case '8': // DECRC - Restore Cursor
		p.restoreCursor()
		p.state = stateNormal
	case 'H': // HTS - Horizontal Tab Set
		// Set tab stop at current position
		p.state = stateNormal
	default:
		// Unknown escape sequence
		p.state = stateNormal
	}
}

func (p *ANSIParser) handleCSI(b byte) {
	// Check if this is a parameter byte or final byte
	if b >= 0x20 && b <= 0x3F {
		// Parameter byte
		p.escapeBuffer.WriteByte(b)
		return
	}

	// Final byte - execute the command
	params := p.parseCSIParams(p.escapeBuffer.String())
	
	switch b {
	case 'A': // Cursor up
		n := 1
		if len(params) > 0 && params[0] > 0 {
			n = params[0]
		}
		p.buffer.MoveCursor(p.buffer.cursorX, p.buffer.cursorY-n)
	case 'B': // Cursor down
		n := 1
		if len(params) > 0 && params[0] > 0 {
			n = params[0]
		}
		p.buffer.MoveCursor(p.buffer.cursorX, p.buffer.cursorY+n)
	case 'C': // Cursor forward
		n := 1
		if len(params) > 0 && params[0] > 0 {
			n = params[0]
		}
		p.buffer.MoveCursor(p.buffer.cursorX+n, p.buffer.cursorY)
	case 'D': // Cursor backward
		n := 1
		if len(params) > 0 && params[0] > 0 {
			n = params[0]
		}
		p.buffer.MoveCursor(p.buffer.cursorX-n, p.buffer.cursorY)
	case 'H', 'f': // Cursor position
		row, col := 1, 1
		if len(params) > 0 {
			row = params[0]
		}
		if len(params) > 1 {
			col = params[1]
		}
		// Convert from 1-based to 0-based
		p.buffer.MoveCursor(col-1, row-1)
	case 'J': // Erase display
		mode := 0
		if len(params) > 0 {
			mode = params[0]
		}
		switch mode {
		case 0: // Clear from cursor to end
			// Clear current line from cursor
			for x := p.buffer.cursorX; x < p.buffer.width; x++ {
				p.buffer.SetCell(x, p.buffer.cursorY, ' ', p.currentFG, p.currentBG, Attributes{})
			}
			// Clear lines below
			for y := p.buffer.cursorY + 1; y < p.buffer.height; y++ {
				p.buffer.ClearLine(y)
			}
		case 1: // Clear from start to cursor
			// Clear lines above
			for y := 0; y < p.buffer.cursorY; y++ {
				p.buffer.ClearLine(y)
			}
			// Clear current line to cursor
			for x := 0; x <= p.buffer.cursorX; x++ {
				p.buffer.SetCell(x, p.buffer.cursorY, ' ', p.currentFG, p.currentBG, Attributes{})
			}
		case 2: // Clear entire display
			p.buffer.Clear()
		}
	case 'K': // Erase line
		mode := 0
		if len(params) > 0 {
			mode = params[0]
		}
		switch mode {
		case 0: // Clear from cursor to end of line
			for x := p.buffer.cursorX; x < p.buffer.width; x++ {
				p.buffer.SetCell(x, p.buffer.cursorY, ' ', p.currentFG, p.currentBG, Attributes{})
			}
		case 1: // Clear from start of line to cursor
			for x := 0; x <= p.buffer.cursorX; x++ {
				p.buffer.SetCell(x, p.buffer.cursorY, ' ', p.currentFG, p.currentBG, Attributes{})
			}
		case 2: // Clear entire line
			p.buffer.ClearLine(p.buffer.cursorY)
		}
	case 'm': // SGR - Select Graphic Rendition
		p.handleSGR(params)
	case 's': // SCP - Save Cursor Position
		p.saveCursor()
	case 'u': // RCP - Restore Cursor Position
		p.restoreCursor()
	case 'L': // IL - Insert Lines
		n := 1
		if len(params) > 0 && params[0] > 0 {
			n = params[0]
		}
		p.buffer.InsertLines(p.buffer.cursorY, n)
	case 'M': // DL - Delete Lines
		n := 1
		if len(params) > 0 && params[0] > 0 {
			n = params[0]
		}
		p.buffer.DeleteLines(p.buffer.cursorY, n)
	case 'P': // DCH - Delete Characters
		n := 1
		if len(params) > 0 && params[0] > 0 {
			n = params[0]
		}
		p.buffer.DeleteChars(p.buffer.cursorX, p.buffer.cursorY, n)
	case '@': // ICH - Insert Characters
		n := 1
		if len(params) > 0 && params[0] > 0 {
			n = params[0]
		}
		p.buffer.InsertChars(p.buffer.cursorX, p.buffer.cursorY, n)
	case 'X': // ECH - Erase Characters
		n := 1
		if len(params) > 0 && params[0] > 0 {
			n = params[0]
		}
		for i := 0; i < n && p.buffer.cursorX+i < p.buffer.width; i++ {
			p.buffer.SetCell(p.buffer.cursorX+i, p.buffer.cursorY, ' ', p.currentFG, p.currentBG, Attributes{})
		}
	case 'G': // CHA - Cursor Horizontal Absolute
		col := 1
		if len(params) > 0 {
			col = params[0]
		}
		p.buffer.MoveCursor(col-1, p.buffer.cursorY)
	case 'd': // VPA - Vertical Position Absolute
		row := 1
		if len(params) > 0 {
			row = params[0]
		}
		p.buffer.MoveCursor(p.buffer.cursorX, row-1)
	case 'r': // DECSTBM - Set Top and Bottom Margins
		// TODO: Implement scrolling regions
	case 'h': // SM - Set Mode
		// TODO: Implement various modes
	case 'l': // RM - Reset Mode
		// TODO: Implement various modes
	case '?': // Private modes
		if len(p.escapeBuffer.String()) > 0 && p.escapeBuffer.String()[0] == '?' {
			// Handle private modes like ?25h (show cursor), ?25l (hide cursor)
		}
	}

	p.state = stateNormal
}

func (p *ANSIParser) handleOSC(b byte) {
	// OSC sequences are terminated by BEL or ST (ESC \)
	if b == 0x07 { // BEL
		// Process OSC command
		p.processOSC(p.escapeBuffer.String())
		p.state = stateNormal
	} else if b == 0x1B { // Might be start of ST
		// Look for \ to complete ST
		p.escapeBuffer.WriteByte(b)
	} else if b == '\\' && p.escapeBuffer.Len() > 0 && p.escapeBuffer.Bytes()[p.escapeBuffer.Len()-1] == 0x1B {
		// Found ST (ESC \)
		p.processOSC(p.escapeBuffer.String()[:p.escapeBuffer.Len()-1])
		p.state = stateNormal
	} else {
		p.escapeBuffer.WriteByte(b)
	}
}

func (p *ANSIParser) parseCSIParams(s string) []int {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ";")
	params := make([]int, 0, len(parts))
	
	for _, part := range parts {
		if part == "" {
			params = append(params, 0)
		} else {
			n, err := strconv.Atoi(part)
			if err == nil {
				params = append(params, n)
			}
		}
	}

	return params
}

func (p *ANSIParser) handleSGR(params []int) {
	if len(params) == 0 {
		params = []int{0}
	}

	for i := 0; i < len(params); i++ {
		switch params[i] {
		case 0: // Reset
			p.currentFG = Color{Default: true}
			p.currentBG = Color{Default: true}
			p.currentAttrs = Attributes{}
		case 1: // Bold
			p.currentAttrs.Bold = true
		case 3: // Italic
			p.currentAttrs.Italic = true
		case 4: // Underline
			p.currentAttrs.Underline = true
		case 5: // Blink
			p.currentAttrs.Blink = true
		case 7: // Reverse
			p.currentAttrs.Reverse = true
		case 8: // Hidden
			p.currentAttrs.Hidden = true
		case 22: // Not bold
			p.currentAttrs.Bold = false
		case 23: // Not italic
			p.currentAttrs.Italic = false
		case 24: // Not underline
			p.currentAttrs.Underline = false
		case 25: // Not blink
			p.currentAttrs.Blink = false
		case 27: // Not reverse
			p.currentAttrs.Reverse = false
		case 28: // Not hidden
			p.currentAttrs.Hidden = false
		case 30, 31, 32, 33, 34, 35, 36, 37: // Foreground colors
			p.currentFG = p.ansiToColor(params[i] - 30)
		case 39: // Default foreground
			p.currentFG = Color{Default: true}
		case 40, 41, 42, 43, 44, 45, 46, 47: // Background colors
			p.currentBG = p.ansiToColor(params[i] - 40)
		case 49: // Default background
			p.currentBG = Color{Default: true}
		case 38: // Extended foreground color
			if i+2 < len(params) && params[i+1] == 5 {
				// 256 color mode
				p.currentFG = p.ansi256ToColor(params[i+2])
				i += 2
			}
		case 48: // Extended background color
			if i+2 < len(params) && params[i+1] == 5 {
				// 256 color mode
				p.currentBG = p.ansi256ToColor(params[i+2])
				i += 2
			}
		}
	}
}

func (p *ANSIParser) ansiToColor(code int) Color {
	// Basic ANSI colors
	colors := []Color{
		{R: 0, G: 0, B: 0},       // Black
		{R: 170, G: 0, B: 0},     // Red
		{R: 0, G: 170, B: 0},     // Green
		{R: 170, G: 85, B: 0},    // Yellow
		{R: 0, G: 0, B: 170},     // Blue
		{R: 170, G: 0, B: 170},   // Magenta
		{R: 0, G: 170, B: 170},   // Cyan
		{R: 170, G: 170, B: 170}, // White
	}

	if code >= 0 && code < len(colors) {
		return colors[code]
	}
	return Color{Default: true}
}

func (p *ANSIParser) ansi256ToColor(code int) Color {
	if code < 16 {
		// Standard and bright colors
		if code < 8 {
			return p.ansiToColor(code)
		} else {
			// Bright colors (8-15)
			return p.ansiBrightToColor(code - 8)
		}
	} else if code < 232 {
		// 216 color cube (16-231)
		code -= 16
		r := (code / 36) * 51
		g := ((code / 6) % 6) * 51
		b := (code % 6) * 51
		return Color{R: uint8(r), G: uint8(g), B: uint8(b)}
	} else {
		// Grayscale (232-255)
		gray := 8 + (code-232)*10
		return Color{R: uint8(gray), G: uint8(gray), B: uint8(gray)}
	}
}

func (p *ANSIParser) ansiBrightToColor(code int) Color {
	// Bright ANSI colors
	colors := []Color{
		{R: 85, G: 85, B: 85},       // Bright Black (Gray)
		{R: 255, G: 85, B: 85},      // Bright Red
		{R: 85, G: 255, B: 85},      // Bright Green
		{R: 255, G: 255, B: 85},     // Bright Yellow
		{R: 85, G: 85, B: 255},      // Bright Blue
		{R: 255, G: 85, B: 255},     // Bright Magenta
		{R: 85, G: 255, B: 255},     // Bright Cyan
		{R: 255, G: 255, B: 255},    // Bright White
	}

	if code >= 0 && code < len(colors) {
		return colors[code]
	}
	return Color{Default: true}
}

// Additional helper methods

func (p *ANSIParser) handleDCS(b byte) {
	// DCS sequences are terminated by ST (ESC \)
	if b == 0x1B {
		p.escapeBuffer.WriteByte(b)
	} else if b == '\\' && p.escapeBuffer.Len() > 0 && p.escapeBuffer.Bytes()[p.escapeBuffer.Len()-1] == 0x1B {
		// Found ST, process DCS
		// For now, we just ignore DCS sequences
		p.state = stateNormal
	} else {
		p.escapeBuffer.WriteByte(b)
	}
}

func (p *ANSIParser) handleCharset(b byte) {
	// Handle character set selection
	// For now, we just ignore these
	p.state = stateNormal
}

func (p *ANSIParser) processOSC(command string) {
	// Process OSC commands (like setting window title)
	// Format: OSC Ps ; Pt BEL
	parts := strings.SplitN(command, ";", 2)
	if len(parts) < 1 {
		return
	}
	
	// Common OSC commands:
	// 0 - Set window title and icon
	// 1 - Set icon 
	// 2 - Set window title
	// We don't need to handle these for a terminal buffer
}

func (p *ANSIParser) saveCursor() {
	p.savedCursor = &cursorState{
		x:     p.buffer.cursorX,
		y:     p.buffer.cursorY,
		fg:    p.currentFG,
		bg:    p.currentBG,
		attrs: p.currentAttrs,
	}
}

func (p *ANSIParser) restoreCursor() {
	if p.savedCursor != nil {
		p.buffer.MoveCursor(p.savedCursor.x, p.savedCursor.y)
		p.currentFG = p.savedCursor.fg
		p.currentBG = p.savedCursor.bg
		p.currentAttrs = p.savedCursor.attrs
	}
}