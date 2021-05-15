package lexer

type TokenCollector interface {
	OpenBrace(lineNumber, position int)
	CloseBrace(lineNumber, position int)
	OpenParen(lineNumber, position int)
	CloseParen(lineNumber, position int)
	OpenAngle(lineNumber, position int)
	CloseAngle(lineNumber, position int)
	Star(lineNumber, position int)
	Colon(lineNumber, position int)
	Name(name string, lineNumber, position int)
	Error(lineNumber, position int)
}
