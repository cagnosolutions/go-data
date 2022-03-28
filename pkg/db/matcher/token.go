package matcher

import (
	"fmt"
)

type tokType int

type Token struct {
	typ tokType
	lit string
}

func (t Token) String() string {
	return fmt.Sprintf("tok.type=%s, tok.literal=%s", tokStrMap[t.typ], t.lit)
}

func (t Token) GetType() tokType {
	return t.typ
}

func (t Token) GetLiteral() string {
	return t.lit
}

const TokEOF = tokEOF

const (
	_ tokType = iota
	tokILLEGAL
	tokEOF
	tokCR
	tokLF

	tokIdent
	tokString
	tokNumber
	tokField
	tokValue

	tokAsterisk
	tokPlus
	tokMinus
	tokPipe

	tokEQ
	tokNE
	tokLT
	tokLE
	tokGT
	tokGE
	tokTrue
	tokFalse
	tokNULL

	tokLBrace
	tokRBrace
	tokLParen
	tokRParen
	tokLBracket
	tokRBracket
	tokDoubleQuote
	tokSingleQuote
	tokBackQuote
	tokColon
	tokSemicolon
	tokComma
)

var tokStrMap = map[tokType]string{
	tokILLEGAL: "ILLEGAL",
	tokEOF:     "EOF",
	tokCR:      "CARRIAGE_RETURN",
	tokLF:      "LINE_FEED",

	tokIdent:  "IDENT",
	tokString: "STRING",
	tokNumber: "NUMBER",
	tokField:  "FIELD",
	tokValue:  "VALUE",

	tokAsterisk: "ASTERISK",
	tokPlus:     "PLUS",
	tokMinus:    "MINUS",
	tokPipe:     "PIPE",

	tokEQ:    "EQUAL",
	tokNE:    "NOT_EQUAL",
	tokLT:    "LESS",
	tokLE:    "LESS_OR_EQUAL",
	tokGT:    "GREATER",
	tokGE:    "GREATER_OR_EQUAL",
	tokTrue:  "TRUE",
	tokFalse: "FALSE",
	tokNULL:  "NULL",

	tokLBrace:      "LEFT_BRACE",
	tokRBrace:      "RIGHT_BRACE",
	tokLParen:      "LEFT_PAREN",
	tokRParen:      "RIGHT_PAREN",
	tokLBracket:    "LEFT_BRACKET",
	tokRBracket:    "RIGHT_BRACKET",
	tokDoubleQuote: "DOUBLE_QUOTE",
	tokSingleQuote: "SINGLE_QUOTE",
	tokBackQuote:   "BACK_QUOTE",
	tokColon:       "COLON",
	tokSemicolon:   "SEMI_COLON",
	tokComma:       "COMMA",
}

// func LookupIdentifierType(identifier string) tokenType {
// 	if tok, ok := keywords[identifier]; ok {
// 		return tok
// 	}
// 	return tokIDENT
// }
