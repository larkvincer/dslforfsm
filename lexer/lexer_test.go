package lexer

import (
	"testing"
)

type testInputs struct {
	input string
	want  string
}

func TestSingleTokens(t *testing.T) {
	testTable := []testInputs{
		{input: "{", want: "openBrace"},
		{input: "}", want: "closeBrace"},
		{input: "(", want: "openParen"},
		{input: ")", want: "closeParen"},
		{input: "<", want: "openAngle"},
		{input: ">", want: "closeAngle"},
		{input: "*", want: "star"},
		{input: ":", want: "colon"},
		{input: "mystate", want: "#mystate#"},
		{input: "state_with_numbers_222", want: "#state_with_numbers_222#"},
		{input: " ", want: ""},
		{input: " \r  \t\n", want: ""},
		{input: " \r  \t *\n", want: "star"},
		{input: ".", want: "error"},
	}

	runLexerTestTable(testTable, t)
}

func TestComments(t *testing.T) {
	testTable := []testInputs{
		{input: "*//comment", want: "star"},
		{input: "*//comment1\n//comment2\n*//comment3", want: "star,star"},
	}

	runLexerTestTable(testTable, t)
}

func TestIntegration(t *testing.T) {
	testTable := []testInputs{
		{input: "{}", want: "openBrace,closeBrace"},
		{input: "{name name *}()<> .",
			want: "openBrace,#name#,#name#,star,closeBrace,openParen,closeParen,openAngle,closeAngle,error"},
	}

	runLexerTestTable(testTable, t)
}

func runLexerTestTable(testTable []testInputs, t *testing.T) {
	for _, test := range testTable {
		tokenCollector := NewTestCollector()
		testLexer := New(tokenCollector)
		testLexer.Lex(test.input)
		got := tokenCollector.tokens
		if got != test.want {
			t.Errorf("Expected '%s', but got '%s'", test.want, got)
		}
	}
}
