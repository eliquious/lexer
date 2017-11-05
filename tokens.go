package lexer

import "strings"

// NOTE: The code below was HEAVILY influenced by how InfluxDB's parser
// handles lexical tokens.
// You can find the code here: https://github.com/influxdb/influxdb/blob/master/influxql/token.go

// Pos specifies the line and character position of a token.
// The Char and Line are both zero-based indexes.
type Pos struct {
	Line int
	Char int
}

// Token represents each lexer symbol
type Token int

// Token enums
const (
	ILLEGAL Token = iota
	EOF
	WS

	// Punctuation

	LPAREN      // (
	RPAREN      // )
	LBRACKET    // [
	RBRACKET    // ]
	LCURLY      // {
	RCURLY      // }
	COMMA       // ,
	SEMICOLON   // ;
	COLON       // :
	DOT         // .
	SINGLEQUOTE // '
	DOUBLEQUOTE // "
	PERCENT     // %
	DOLLAR      // $
	HASH        // #
	ATSIGN      // @

	// Literals

	startLiterals
	IDENT
	INTEGER
	DECIMAL
	STRING
	BADSTRING
	BADESCAPE
	TRUE
	FALSE
	REGEX
	BADREGEX
	DURATION
	endLiterals

	// Operators

	startOperators
	PLUS      // +
	MINUS     // -
	MUL       // *
	DIV       // /
	AMPERSAND // &
	XOR       // ^
	PIPE      // |
	LSHIFT    // <<
	RSHIFT    // >>
	POW       // **
	
	ARROW       // ->
	EQARROW     // =>

	AND // AND
	OR  // OR

	EQ       // =
	NEQ      // !=
	EQREGEX  // =~
	NEQREGEX // !~
	LT       // <
	LTE      // <=
	GT       // >
	GTE      // >=
	endOperators
)

var tokens = map[Token]string{
	ILLEGAL: "ILLEGAL",
	EOF:     "EOF",
	WS:      "WS",

	LPAREN:      "(",
	RPAREN:      ")",
	LBRACKET:    "[",
	RBRACKET:    "]",
	LCURLY:      "{",
	RCURLY:      "}",
	COMMA:       ",",
	SEMICOLON:   ";",
	COLON:       ":",
	DOT:         ".",
	SINGLEQUOTE: "'",
	DOUBLEQUOTE: "\"",
	PERCENT:     "%%",
	DOLLAR:      "$",
	HASH:        "#",
	ATSIGN:      "@",

	IDENT:     "IDENT",
	INTEGER:   "INTEGER",
	DECIMAL:   "DECIMAL",
	STRING:    "TEXTUAL",
	DURATION:  "DURATION",
	BADSTRING: "BADSTRING",
	BADESCAPE: "BADESCAPE",
	REGEX:     "REGEX",
	BADREGEX:  "BADREGEX",

	PLUS:      "+",
	MINUS:     "-",
	MUL:       "*",
	DIV:       "/",
	AMPERSAND: "&",
	XOR:       "^",
	PIPE:      "|",
	RSHIFT:    ">>",
	LSHIFT:    "<<",
	POW:       "**",
	ARROW:     "->",
	EQARROW:   "=>",

	AND: "AND",
	OR:  "OR",

	TRUE:  "true",
	FALSE: "false",

	EQ:       "=",
	NEQ:      "!=",
	EQREGEX:  "=~",
	NEQREGEX: "!~",

	LT:  "<",
	LTE: "<=",
	GT:  ">",
	GTE: ">=",
}

var keywords = map[string]Token{}

func init() {
	keywords["and"] = AND
	keywords["or"] = OR
	keywords["true"] = TRUE
	keywords["false"] = FALSE
}

// LoadTokenMap allows for extra keywords to be added to the lexer
func LoadTokenMap(keywordTokens map[Token]string) {

	// Combine built-in tokens and keywords
	for k, v := range keywordTokens {
		tokens[k] = v
	}

	// Load Keywords
	for k, v := range keywordTokens {
		keywords[strings.ToLower(v)] = k
	}
}

// String returns the string representation of the token.
func (tok Token) String() string {
	if _, ok := tokens[tok]; ok {
		return tokens[tok]
	}
	return ""
}

// Precedence returns the operator precedence of the binary operator token.
func (tok Token) Precedence() int {
	switch tok {
	case OR:
		return 1
	case AND:
		return 2
	case EQ, NEQ, EQREGEX, NEQREGEX, LT, LTE, GT, GTE:
		return 3
	case PLUS, MINUS:
		return 4
	case MUL, DIV:
		return 5
	case PIPE, XOR, RSHIFT, LSHIFT, POW:
		return 6
	}
	return 0
}

// isOperator returns true for operator tokens.
func (tok Token) isOperator() bool { return tok > startOperators && tok < endOperators }

// tokstr returns a literal if provided, otherwise returns the token string.
func tokstr(tok Token, lit string) string {
	if lit != "" {
		return lit
	}
	return tok.String()
}

// Lookup returns the token associated with a given string.
func Lookup(ident string) Token {
	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return IDENT
}
