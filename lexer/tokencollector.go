package lexer

type TokenCollector interface {
	openBrace(lineNumber int, position int)
	closeBrace(lineNumber int, position int)
	openParen(lineNumber int, position int)
	closeParen(lineNumber int, position int)
	openAngle(lineNumber int, position int)
	closeAngle(lineNumber int, position int)
	star(lineNumber int, position int)
	colon(lineNumber int, position int)
	name(name string, lineNumber int, position int)
	error(lineNumber int, position int)
}
