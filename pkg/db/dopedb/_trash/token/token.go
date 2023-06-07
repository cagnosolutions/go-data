package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
}

const (

	// Special
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers
	IDENT = "IDENT"

	// Literals
	INT    = "INT"
	STRING = "STRING"

	// Operators
	ASSIGN = "="
	PLUS   = "+"
	MINUS  = "-"

	// Delimiters
	COMMA     = ","
	SEMICOLON = ";"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"

	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
	VAR      = "VAR"

	SET = "SET"
	GET = "GET"
	DEL = "DEL"

	ZSET = "ZSET"
	ZGET = "ZGET"

	HSET = "HSET"
	HGET = "HGET"

	INCR = "INCR"
	DECR = "DECR"
)

var keywords = map[string]TokenType{
	"fn":  FUNCTION,
	"let": LET,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
