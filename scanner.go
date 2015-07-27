package lexer

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// NOTE: The code below is HEAVILY influenced by the InfluxQL parser and lexer.
// You can find the InfluxDB code at https://github.com/influxdb/influxdb/blob/master/influxql/scanner.go

// Scanner represents a lexical scanner for InfluxQL.
type Scanner struct {
	r *reader
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: &reader{r: bufio.NewReader(r)}}
}

// Scan returns the next token and position from the underlying reader.
// Also returns the literal text read for strings and numbers tokens
// since these token types can have different literal representations.
func (s *Scanner) Scan() (tok Token, pos Pos, lit string) {

	// Read next code point.
	ch0, pos := s.r.read()

	// If we see whitespace then consume all contiguous whitespace.
	// If we see a letter, or certain acceptable special characters, then consume
	// as an ident or reserved word.
	if isWhitespace(ch0) {
		return s.scanWhitespace()
	} else if isLetter(ch0) || ch0 == '_' {
		s.r.unread()
		return s.scanIdent()
	} else if isDigit(ch0) {
		return s.scanNumber()
	}

	// Otherwise parse individual characters.
	switch ch0 {
	case eof:
		return EOF, pos, ""
	case '"':
		s.r.unread()
		return s.scanIdent()
	case '\'':
		return s.scanString()
	case '.':
		ch1, _ := s.r.read()
		s.r.unread()
		if isDigit(ch1) {
			return s.scanNumber()
		}
		return DOT, pos, ""
	case '+', '-':
		return s.scanNumber()
	case '*':
		return MUL, pos, ""
	case '/':
		return DIV, pos, ""
	case '=':
		if ch1, _ := s.r.read(); ch1 == '~' {
			return EQREGEX, pos, ""
		}
		s.r.unread()
		return EQ, pos, ""
	case '!':
		if ch1, _ := s.r.read(); ch1 == '=' {
			return NEQ, pos, ""
		} else if ch1 == '~' {
			return NEQREGEX, pos, ""
		}
		s.r.unread()
	case '>':
		if ch1, _ := s.r.read(); ch1 == '=' {
			return GTE, pos, ""
		} else if ch1 == '>' {
			return RSHIFT, pos, ""
		}
		s.r.unread()
		return GT, pos, ""
	case '<':
		if ch1, _ := s.r.read(); ch1 == '=' {
			return LTE, pos, ""
		} else if ch1 == '>' {
			return NEQ, pos, ""
		} else if ch1 == '<' {
			return LSHIFT, pos, ""
		}
		s.r.unread()
		return LT, pos, ""
	case '(':
		return LPAREN, pos, ""
	case ')':
		return RPAREN, pos, ""
	case '[':
		return LBRACKET, pos, ""
	case ']':
		return RBRACKET, pos, ""
	case '{':
		return LCURLY, pos, ""
	case '}':
		return RCURLY, pos, ""
	case ',':
		return COMMA, pos, ""
	case ';':
		return SEMICOLON, pos, ""
	case ':':
		return COLON, pos, ""
	case '^':
		return XOR, pos, ""
	case '|':
		return PIPE, pos, ""
	case '&':
		return AMPERSAND, pos, ""
	case '%':
		return PERCENT, pos, ``
	case '$':
		return DOLLAR, pos, ""
	case '#':
		return HASH, pos, ""
	case '@':
		return ATSIGN, pos, ""
	}

	return ILLEGAL, pos, string(ch0)
}

// read reads the next rune from the bufferred reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return eof
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *Scanner) unread() { _ = s.r.UnreadRune() }

// scanWhitespace consumes the current rune and all contiguous whitespace.
func (s *Scanner) scanWhitespace() (tok Token, pos Pos, lit string) {
	// Create a buffer and read the current character into it.
	var buf bytes.Buffer
	ch, pos := s.r.curr()
	_, _ = buf.WriteRune(ch)

	// Read every subsequent whitespace character into the buffer.
	// Non-whitespace characters and EOF will cause the loop to exit.
	for {
		ch, _ = s.r.read()
		if ch == eof {
			break
		} else if !isWhitespace(ch) {
			s.r.unread()
			break
		} else {
			_, _ = buf.WriteRune(ch)
		}
	}

	return WS, pos, buf.String()
}

func (s *Scanner) scanIdent() (tok Token, pos Pos, lit string) {
	// Save the starting position of the identifier.
	_, pos = s.r.read()
	s.r.unread()

	var buf bytes.Buffer
	for {
		if ch, _ := s.r.read(); ch == eof {
			break
		} else if ch == '"' {
			tok0, pos0, lit0 := s.scanString()
			if tok0 == BADSTRING || tok0 == BADESCAPE {
				return tok0, pos0, lit0
			}
			return IDENT, pos, lit0
		} else if isIdentChar(ch) {
			s.r.unread()
			buf.WriteString(ScanBareIdent(s.r))
		} else {
			s.r.unread()
			break
		}
	}
	lit = buf.String()

	// If the literal matches a keyword then return that keyword.
	if tok = Lookup(lit); tok != IDENT {
		return tok, pos, ""
	}

	return IDENT, pos, lit
}

// scanString consumes a contiguous string of non-quote characters.
// Quote characters can be consumed if they're first escaped with a backslash.
func (s *Scanner) scanString() (tok Token, pos Pos, lit string) {
	s.r.unread()
	_, pos = s.r.curr()

	var err error
	lit, err = ScanString(s.r)
	if err == errBadString {
		return BADSTRING, pos, lit
	} else if err == errBadEscape {
		_, pos = s.r.curr()
		return BADESCAPE, pos, lit
	}
	return STRING, pos, lit
}

// scanNumber consumes anything that looks like the start of a number.
// Numbers start with a digit, full stop, plus sign or minus sign.
// This function can return non-number tokens if a scan is a false positive.
// For example, a minus sign followed by a letter will just return a minus sign.
func (s *Scanner) scanNumber() (tok Token, pos Pos, lit string) {
	var buf bytes.Buffer

	// Check if the initial rune is a "+" or "-".
	ch, pos := s.r.curr()
	if ch == '+' || ch == '-' {
		// Peek at the next two runes.
		ch1, _ := s.r.read()
		ch2, _ := s.r.read()
		s.r.unread()
		s.r.unread()

		// This rune must be followed by a digit or a full stop and a digit.
		if isDigit(ch1) || (ch1 == '.' && isDigit(ch2)) {
			_, _ = buf.WriteRune(ch)
		} else if ch == '+' {
			return PLUS, pos, ""
		} else if ch == '-' {
			return MINUS, pos, ""
		}
	} else if ch == '.' {
		// Peek and see if the next rune is a digit.
		ch1, _ := s.r.read()
		s.r.unread()
		if !isDigit(ch1) {
			return ILLEGAL, pos, "."
		}

		// Unread the full stop so we can read it later.
		s.r.unread()
	} else {
		s.r.unread()
	}

	// Read as many digits as possible.
	_, _ = buf.WriteString(s.scanDigits())

	// If next code points are a full stop and digit then consume them.
	if ch0, _ := s.r.read(); ch0 == '.' {
		if ch1, _ := s.r.read(); isDigit(ch1) {
			_, _ = buf.WriteRune(ch0)
			_, _ = buf.WriteRune(ch1)
			_, _ = buf.WriteString(s.scanDigits())
		} else {
			s.r.unread()
			s.r.unread()
		}
	} else {
		s.r.unread()
	}

	// Attempt to read as a duration if it doesn't have a fractional part.
	if !strings.Contains(buf.String(), ".") {
		// If the next rune is a duration unit (u,µ,ms,s) then return a duration token
		if ch0, _ := s.r.read(); ch0 == 'u' || ch0 == 'µ' || ch0 == 's' || ch0 == 'h' || ch0 == 'd' || ch0 == 'w' {
			_, _ = buf.WriteRune(ch0)
			return DURATION_VAL, pos, buf.String()
		} else if ch0 == 'm' {
			_, _ = buf.WriteRune(ch0)
			if ch1, _ := s.r.read(); ch1 == 's' {
				_, _ = buf.WriteRune(ch1)
			} else {
				s.r.unread()
			}
			return DURATION_VAL, pos, buf.String()
		}
		s.r.unread()
	}
	return NUMBER, pos, buf.String()
}

// scanDigits consume a contiguous series of digits.
func (s *Scanner) scanDigits() string {
	var buf bytes.Buffer
	for {
		ch, _ := s.r.read()
		if !isDigit(ch) {
			s.r.unread()
			break
		}
		_, _ = buf.WriteRune(ch)
	}
	return buf.String()
}

func (s *Scanner) ScanRegex() (tok Token, pos Pos, lit string) {
	_, pos = s.r.curr()

	// Start & end sentinels.
	start, end := '/', '/'
	// Valid escape chars.
	escapes := map[rune]rune{'/': '/'}

	b, err := ScanDelimited(s.r, start, end, escapes, true)

	if err == errBadEscape {
		_, pos = s.r.curr()
		return BADESCAPE, pos, lit
	} else if err != nil {
		return BADREGEX, pos, lit
	}
	return REGEX, pos, string(b)
}

// bufScanner represents a wrapper for scanner to add a buffer.
// It provides a fixed-length circular buffer that can be unread.
type bufScanner struct {
	s   *Scanner
	i   int // buffer index
	n   int // buffer size
	buf [3]struct {
		tok Token
		pos Pos
		lit string
	}
}

// newBufScanner returns a new buffered scanner for a reader.
func newBufScanner(r io.Reader) *bufScanner {
	return &bufScanner{s: NewScanner(r)}
}

// Scan reads the next token from the scanner.
func (s *bufScanner) Scan() (tok Token, pos Pos, lit string) {
	return s.scanFunc(s.s.Scan)
}

// ScanRegex reads a regex token from the scanner.
func (s *bufScanner) ScanRegex() (tok Token, pos Pos, lit string) {
	return s.scanFunc(s.s.ScanRegex)
}

// scanFunc uses the provided function to scan the next token.
func (s *bufScanner) scanFunc(scan func() (Token, Pos, string)) (tok Token, pos Pos, lit string) {
	// If we have unread tokens then read them off the buffer first.
	if s.n > 0 {
		s.n--
		return s.curr()
	}

	// Move buffer position forward and save the token.
	s.i = (s.i + 1) % len(s.buf)
	buf := &s.buf[s.i]
	buf.tok, buf.pos, buf.lit = scan()

	return s.curr()
}

// Unscan pushes the previously token back onto the buffer.
func (s *bufScanner) Unscan() { s.n++ }

// curr returns the last read token.
func (s *bufScanner) curr() (tok Token, pos Pos, lit string) {
	buf := &s.buf[(s.i-s.n+len(s.buf))%len(s.buf)]
	return buf.tok, buf.pos, buf.lit
}