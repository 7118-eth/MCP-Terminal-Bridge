package terminal

import (
	"bytes"
	"strconv"
	"strings"
)

type ANSIParser struct {
	buffer       *ScreenBuffer
	state        parserState
	escapeBuffer bytes.Buffer
	currentFG    Color
	currentBG    Color
	currentAttrs Attributes
}

type parserState int

const (
	stateNormal parserState = iota
	stateEscape
	stateCSI
	stateOSC
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
	case 'c': // Reset
		p.buffer.Clear()
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
	}

	p.state = stateNormal
}

func (p *ANSIParser) handleOSC(b byte) {
	// For now, just consume OSC sequences without processing
	if b == 0x07 || b == 0x1B { // BEL or ESC
		p.state = stateNormal
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
	// Simplified 256 color support
	// For now, just map to basic colors
	if code < 8 {
		return p.ansiToColor(code)
	}
	return Color{R: uint8(code), G: uint8(code), B: uint8(code)}
}