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
	lexer.lineNumber = 1
	lines := strings.Split(source, "\n")
	for _, line := range lines {
		lexer.lexLine(line)
		lexer.lineNumber++
	}
}

func (lexer *Lexer) lexLine(line string) {
	for lexer.readPosition = 0; lexer.readPosition < len(line); {
		lexer.lexToken(line)
	}
}

func (lexer *Lexer) lexToken(line string) {
	if !lexer.findToken(line) {
		lexer.collector.Error(lexer.lineNumber, lexer.readPosition+1)
		lexer.readPosition++
	}
}

func (lexer *Lexer) findToken(line string) bool {
	return lexer.findSkipSpace(line) || lexer.findSingleCharacterToken(line) || lexer.findName(line)
}

func (lexer *Lexer) findSkipSpace(line string) bool {
	whiteSpacePattern := regexp.MustCompile("^\\s+")
	commentPattern := regexp.MustCompile("^//.*$")
	substring := string(line[lexer.readPosition:])

	if commentPattern.MatchString(substring) {
		lexer.readPosition += commentPattern.FindStringIndex(substring)[1]
		return true
	}

	if whiteSpacePattern.MatchString(substring) {
		lexer.readPosition += whiteSpacePattern.FindStringIndex(substring)[1]
		return true
	}

	return false
}

func (lexer *Lexer) findSingleCharacterToken(line string) bool {
	char := line[lexer.readPosition : lexer.readPosition+1]
	switch char {
	case tokens.OPEN_BRACE:
		lexer.collector.OpenBrace(lexer.lineNumber, lexer.readPosition)
		break
	case tokens.CLOSE_BRACE:
		lexer.collector.CloseBrace(lexer.lineNumber, lexer.readPosition)
		break
	case tokens.OPEN_PAREN:
		lexer.collector.OpenParen(lexer.lineNumber, lexer.readPosition)
		break
	case tokens.CLOSE_PAREN:
		lexer.collector.CloseParen(lexer.lineNumber, lexer.readPosition)
		break
	case tokens.ENTRY_STATE:
		lexer.collector.OpenAngle(lexer.lineNumber, lexer.readPosition)
		break
	case tokens.EXIT_STATE:
		lexer.collector.CloseAngle(lexer.lineNumber, lexer.readPosition)
		break
	case tokens.STAR:
		lexer.collector.Star(lexer.lineNumber, lexer.readPosition)
		break
	case tokens.COLON:
		lexer.collector.Colon(lexer.lineNumber, lexer.readPosition)
		break
	default:
		return false
	}
	lexer.readPosition++
	return true
}

func (lexer *Lexer) findName(line string) bool {
	namePattern := regexp.MustCompile("^\\w+")
	substring := line[lexer.readPosition:]
	if namePattern.MatchString(substring) {
		lexer.collector.Name(namePattern.FindString(substring), lexer.lineNumber, lexer.readPosition)
		lexer.readPosition += namePattern.FindStringIndex(substring)[1]
		return true
	}
	return false
}
