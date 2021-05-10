package lexer

import (
	"regexp"
	"strings"

	"github.com/larkvincer/dsl-fsm/tokens"
)

type Lexer struct {
	lineNumber   int
	readPosition int
	collector    TokenCollector
}

func New(collector TokenCollector) *Lexer {
	lexer := &Lexer{collector: collector}
	return lexer
}

func (lexer *Lexer) Lex(source string) {
	lineNumber := 1
	lines := strings.Split(source, "\n")

	for _, line := range lines {
		lexer.lexLine(line)
		lineNumber++
	}
}

func (lexer *Lexer) lexLine(line string) {
	for lexer.readPosition = 0; lexer.readPosition < len(line); {
		lexer.lexToken(line)
	}
}

func (lexer *Lexer) lexToken(line string) {
	if !lexer.findToken(line) {
		lexer.collector.error(lexer.lineNumber, lexer.readPosition+1)
		lexer.readPosition++
	}
}

func (lexer *Lexer) findToken(line string) bool {
	return lexer.findWhiteSpace(line) || lexer.findSingleCharacteToken(line) || lexer.findName(line)
}

func (lexer *Lexer) findWhiteSpace(line string) bool {
	whiteSpacePattern := regexp.MustCompile("^\\s+")
	commentPattern := regexp.MustCompile("^//.*$")
	substring := string(line[lexer.readPosition])

	return commentPattern.MatchString(substring) || whiteSpacePattern.MatchString(substring)
}

func (lexer *Lexer) findSingleCharacteToken(line string) bool {
	char := line[lexer.readPosition : lexer.readPosition+1]
	switch char {
	case tokens.LCURLY:
		lexer.collector.openBrace(lexer.lineNumber, lexer.readPosition)
		break
	case tokens.RCURLY:
		lexer.collector.closeBrace(lexer.lineNumber, lexer.readPosition)
		break
	case tokens.LPAREN:
		lexer.collector.openParen(lexer.lineNumber, lexer.readPosition)
		break
	case tokens.RPAREN:
		lexer.collector.closeParen(lexer.lineNumber, lexer.readPosition)
		break
	case tokens.ENTRYSTATE:
		lexer.collector.openAngle(lexer.lineNumber, lexer.readPosition)
		break
	case tokens.EXITSTATE:
		lexer.collector.closeAngle(lexer.lineNumber, lexer.readPosition)
		break
	case tokens.STAR:
		lexer.collector.star(lexer.lineNumber, lexer.readPosition)
		break
	case tokens.COLON:
		lexer.collector.colon(lexer.lineNumber, lexer.readPosition)
		break
	default:
		return false
	}
	return true
}

func (lexer *Lexer) findName(line string) bool {
	namePattern := regexp.MustCompile("^\\w+")
	if namePattern.MatchString(line) {
		lexer.collector.name(namePattern.FindString(line), lexer.lineNumber, lexer.readPosition)
		lexer.readPosition = namePattern.FindStringIndex(line)[1]
		return true
	}
	return false
}
