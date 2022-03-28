package matcher

type Lexer struct {
	input   string
	pos     int  // current position in input
	nextPos int  // next reading position after pos
	ch      byte // current char under examination
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input: input,
	}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespace()

	switch l.ch {
	case '=':
		tok = newToken(tokEQ, l.ch)
	case '+':
		tok = newToken(tokPlus, l.ch)
	case '-':
		tok = newToken(tokMinus, l.ch)
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = Token{
				typ: tokNE,
				lit: literal,
			}
		}
	case '*':
		tok = newToken(tokAsterisk, l.ch)
	case '<':
		tok = newToken(tokLT, l.ch)
	case '>':
		tok = newToken(tokGT, l.ch)
	case ',':
		tok = newToken(tokComma, l.ch)
	case ';':
		tok = newToken(tokSemicolon, l.ch)
	case ':':
		tok = newToken(tokColon, l.ch)
	case '(':
		tok = newToken(tokLParen, l.ch)
	case ')':
		tok = newToken(tokRParen, l.ch)
	case '{':
		tok = newToken(tokLBrace, l.ch)
	case '}':
		tok = newToken(tokRBrace, l.ch)
	case '[':
		tok = newToken(tokLBracket, l.ch)
	case ']':
		tok = newToken(tokRBracket, l.ch)
	case '"':
		tok.typ = tokString
		tok.lit = l.readString()
	case 0:
		tok.lit = ""
		tok.typ = tokEOF
	default:
		if isLetter(l.ch) {
			tok.lit = l.readIdentifier()
			tok.typ = tokIdent
			return tok
		} else if isDigit(l.ch) {
			tok.lit = l.readNumber()
			tok.typ = tokNumber
			return tok
		} else {
			tok = newToken(tokILLEGAL, l.ch)
		}
	}

	l.readChar()
	return tok
}

func newToken(tokenType tokType, ch byte) Token {
	return Token{
		typ: tokenType,
		lit: string(ch),
	}
}

func (l *Lexer) readChar() {
	l.ch = l.peekChar()
	l.pos = l.nextPos
	l.nextPos += 1
}

func (l *Lexer) readString() string {
	pos := l.pos + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readNumber() string {
	pos := l.pos
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func (l *Lexer) readIdentifier() string {
	pos := l.pos
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[pos:l.pos]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) peekChar() byte {
	if l.nextPos >= len(l.input) {
		return 0
	} else {
		return l.input[l.nextPos]
	}
}
