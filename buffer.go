package lexer

import (
	"io"
)

// TokenBuffer represents a wrapper for scanner to add a buffer.
// It provides a fixed-length circular buffer that can be unread.
type TokenBuffer struct {
	s   *Scanner
	i   int // buffer index
	n   int // buffer size
	buf [6]struct {
		tok Token
		pos Pos
		lit string
	}
}

// NewTokenBuffer returns a new buffered scanner for a reader.
func NewTokenBuffer(r io.Reader) *TokenBuffer {
	return &TokenBuffer{s: NewScanner(r)}
}

// Scan reads the next token from the scanner.
func (s *TokenBuffer) Scan() (tok Token, pos Pos, lit string) {
	return s.ScanFunc(s.s.Scan)
}

// ScanRegex reads a regex token from the scanner.
func (s *TokenBuffer) ScanRegex() (tok Token, pos Pos, lit string) {
	return s.ScanFunc(s.s.ScanRegex)
}

// ScanFunc uses the provided function to scan the next token.
func (s *TokenBuffer) ScanFunc(scan func() (Token, Pos, string)) (tok Token, pos Pos, lit string) {
	// If we have unread tokens then read them off the buffer first.
	if s.n > 0 {
		s.n--
		return s.Current()
	}

	// Move buffer position forward and save the token.
	s.i = (s.i + 1) % len(s.buf)
	buf := &s.buf[s.i]
	buf.tok, buf.pos, buf.lit = scan()

	return s.Current()
}

// Unscan pushes the previously token back onto the buffer.
func (s *TokenBuffer) Unscan() { s.n++ }

// curr returns the last read token.
func (s *TokenBuffer) Current() (tok Token, pos Pos, lit string) {
	buf := &s.buf[(s.i-s.n+len(s.buf))%len(s.buf)]
	return buf.tok, buf.pos, buf.lit
}

// Peek returns the next rune from the scanner.
func (s *TokenBuffer) Peek() rune {
	return s.s.Peek()
}
