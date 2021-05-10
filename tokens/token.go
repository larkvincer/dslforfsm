package tokens

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"
	IDENT   = "IDENT"

	LPAREN = "("
	RPAREN = ")"
	LCURLY = "{"
	RCURLY = "}"

	ENTRYSTATE = "<"
	EXITSTATE  = ">"

	STAR    = "*"
	COLON = ":"
)

var keywords = map[string]Type{
	"Initial": "INITIAL",
	"FSM":     "NAME",
}

type Type string

type Token struct {
	Type     Type
	Literal  string
	Line     int64
	Position int64
}

func New(tokenType Type, char byte) Token {
	return Token{Type: tokenType, Literal: string(char)}
}

func LookupIdent(ident string) Type {
	if token, ok := keywords[ident]; ok {
		return token
	}

	return IDENT
}
