# lexer
This is a generic tokenizer written in Go. It is meant to be used by higher level parsers for writing DSLs. It is extensible by allowing keywords to be added to the default identifiers.

One example project is [PrefixDB](https://github.com/eliquious/prefixdb) which is a toy database project that has it's own query language.



## Extensibility

Custom tokens can be added to the lexer. Here's how you could add several tokens.

```
package lexer

import (
	"github.com/eliquious/lexer"
)

func init() {
	// Loads keyword tokens into lexer
	lexer.LoadTokenMap(tokenMap)
}

const (
	// Starts the keywords with an offset from the built in tokens
	startKeywords lexer.Token = iota + 1000

	// CREATE starts a CREATE KEYSPACE query.
	CREATE

	// SELECT starts a SELECT FROM query.
	SELECT

	// UPSERT inserts or replaces a key-value pair.
	UPSERT

	// DELETE deletes keys from a keyspace.
	DELETE
	endKeywords
)

var tokenMap = map[lexer.Token]string{

	CREATE:   "CREATE",
	SELECT:   "SELECT",
	DELETE:   "DELETE",
	UPSERT:   "UPSERT",
}

// IsKeyword returns true if the token is a custom keyword.
func IsKeyword(tok lexer.Token) bool {
	return tok > startKeywords && tok < endKeywords
}
```

## Example output

A simple test of the lexer is to write a command line tool that outputs the tokens that come from STDIN. This tool is the the [PrefixDB repository](https://github.com/eliquious/prefixdb). By importing the PrefixDB package, all the custom tokens are registered with the lexer and will be output during scanning.

```
package main

import (
        "fmt"
        "os"
        "strconv"

        "github.com/eliquious/lexer"
        _ "github.com/eliquious/prefixdb/lexer"
)

func main() {
        l := lexer.NewScanner(os.Stdin)
        for {
                tok, pos, lit := l.Scan()

                // exit if EOF
                if tok == lexer.EOF {
                        break
                }

                // skip whitespace tokens
                if tok == lexer.WS {
                        continue
                }

                // Print token
                if len(lit) > 0 {
                        fmt.Printf("[%4d:%-3d] %10s - %s\n", pos.Line, pos.Char, tok, strconv.QuoteToASCII(lit))
                } else {
                        fmt.Printf("[%4d:%-3d] %10s\n", pos.Line, pos.Char, tok)
                }
        }
}
```

An example output would look like this:

```
$ echo 'SELECT FROM users WHERE username = "bugs.bunny" OR "daffy.duck" AND timestamp BETWEEN "2015-01-01" AND "2016-01-01" AND topic = "hunting"' | go run main.go
[   0:0  ]     SELECT
[   0:7  ]       FROM
[   0:12 ]      IDENT - "users"
[   0:18 ]      WHERE
[   0:24 ]      IDENT - "username"
[   0:33 ]          =
[   0:34 ]    TEXTUAL - "bugs.bunny"
[   0:48 ]         OR
[   0:50 ]    TEXTUAL - "daffy.duck"
[   0:64 ]        AND
[   0:68 ]      IDENT - "timestamp"
[   0:78 ]    BETWEEN
[   0:85 ]    TEXTUAL - "2015-01-01"
[   0:99 ]        AND
[   0:102]    TEXTUAL - "2016-01-01"
[   0:116]        AND
[   0:120]      IDENT - "topic"
[   0:126]          =
[   0:127]    TEXTUAL - "hunting"
```

As you can see, the custom tokens are output such as `SELECT`, `FROM` and `WHERE`, etc.
