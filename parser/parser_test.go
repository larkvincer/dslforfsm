package parser

import (
	"testing"

	"github.com/larkvincer/dsl-fsm/lexer"
)

type parserTest struct {
	name     string
	source   string
	expected string
}

func TestHeaderParsing(t *testing.T) {
	testTable := []parserTest{
		{"one header", "N:V{}", "N:V\n.\n"},
		{"many headers", " N1 : V1\tN2 : V2\n{}", "N1:V1\nN2:V2\n.\n"},
		{"no header", "{}", ".\n"},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			got := parseSource(testCase.source)
			if got != testCase.expected {
				t.Fatalf("expected '%s' for %s, but got '%s'", testCase.expected, testCase.source, got)
			}
		})
	}
}

func TestBodyParsing(t *testing.T) {
	testTable := []parserTest{
		{"simple transition", "{ s e ns a}", "{\n  s e ns a\n}\n.\n"},
		{"transition with null action", "{ s e ns *}", "{\n  s e ns {}\n}\n.\n"},
		{"transition with many actions", "{ s e ns {a1 a2}}", "{\n  s e ns {a1 a2}\n}\n.\n"},
		{"state with subtransition", "{ s {e ns a}}", "{\n  s e ns a\n}\n.\n"},
		{"state with serveral subtransitions", "{ s {e1 ns a1 e2 ns a2}}", "{\n  s {\n    e1 ns a1\n    e2 ns a2\n  }\n}\n.\n"},
		{"many transitions", "{s1 e1 s2 a1 s2 e2 s3 a2}", "{\n  s1 e1 s2 a1\n  s2 e2 s3 a2\n}\n.\n"},
		{"many transitions", "{s1 e1 s2 a1 s2 e2 s3 a2}", "{\n  s1 e1 s2 a1\n  s2 e2 s3 a2\n}\n.\n"},
		{"super state", "{(ss) e s a}", "{\n  (ss) e s a\n}\n.\n"},
		{"entry action", "{s <ea e ns a}", "{\n  s <ea e ns a\n}\n.\n"},
		{"exit action", "{s >xa e ns a}", "{\n  s >xa e ns a\n}\n.\n"},
		{"derived state", "{s:ss e ns a}", "{\n  s:ss e ns a\n}\n.\n"},
		{"all state adornments", "{(s)<ea>xa:ss e ns a}", "{\n  (s):ss <ea >xa e ns a\n}\n.\n"},
		{"state with no subtransitions", "{s {}}", "{\n  s {\n  }\n}\n.\n"},
		{"state with all stars", "{s * * *}", "{\n  s * * {}\n}\n.\n"},
		{"multiple super states", "{s :x :y * * *}", "{\n  s:x:y * * {}\n}\n.\n"},
		{"multiple exit actions", "{s >x >y * * *}", "{\n  s >x >y * * {}\n}\n.\n"},
		{"multiple exit and entry actions with braces", "{s <{u v} >{w x} * * *}", "{\n  s <u <v >w >x * * {}\n}\n.\n"},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			got := parseSource(testCase.source)
			if got != testCase.expected {
				t.Fatalf("expected '%s' for %s, but got '%s'", testCase.expected, testCase.source, got)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	testTable := []parserTest{
		{"parse nothing", "", "Syntax error: HEADER. HEADER|EOF. line -1, position -1.\n"},
		{"header with no colon or value", "A {s e ns a}", "Syntax error: HEADER. HEADER_COLON|{. line 1, position 2.\n"},
		{"header with no value", "A: {s e ns a}", "Syntax error: HEADER. HEADER_VALUE|{. line 1, position 3.\n"},
		{"header with no value", "A: {s e ns a}", "Syntax error: HEADER. HEADER_VALUE|{. line 1, position 3.\n"},
		{"transition missing event, next state and action", "{s}", "Syntax error: STATE. STATE_MODIFIER|}. line 1, position 2.\n"},
		{"transition missing next state and action", "{s e}", "Syntax error: TRANSITION. SINGLE_EVENT|}. line 1, position 4.\n"},
		{"transition missing action", "{s e ns}", "Syntax error: TRANSITION. SINGLE_NEXT_STATE|}. line 1, position 7.\n"},
		{"no closing brace", "{", "Syntax error: STATE. STATE_SPEC|EOF. line -1, position -1.\n"},
		{"initial state skipped", "{* e ns a}", "Syntax error: STATE. STATE_SPEC|*. line 1, position 1.\n"},
		{"lexical error", "{. e ns a}", "Syntax error: SYNTAX. . line 1, position 2.\n"},
	}

	for _, testCase := range testTable {
		syntaxBuilder := NewFsmSyntaxBuilder()
		parser := NewParser(syntaxBuilder)
		lexer := lexer.New(parser)

		t.Run(testCase.name, func(t *testing.T) {
			lexer.Lex(testCase.source)
			parser.HandleEvent("EOF", -1, -1)
			got := syntaxBuilder.fsmSyntax.GetErrors()
			if got != testCase.expected {
				t.Fatalf("expected '%s' for %s, but got '%s'", testCase.expected, testCase.source, got)
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	testTable := []parserTest{
		{
			"simple one coin turnstile",
			"" +
				"Actions: Turnstile\n" +
				"FSM: OneCoinTurnstile\n" +
				"Initial: Locked\n" +
				"{\n" +
				"  Locked\tCoin\tUnlocked\t{alarmOff unlock}\n" +
				"  Locked \tPass\tLocked\t\talarmOn\n" +
				"  Unlocked\tCoin\tUnlocked\tthankyou\n" +
				"  Unlocked\tPass\tLocked\t\tlock\n" +
				"}",
			"" +
				"Actions:Turnstile\n" +
				"FSM:OneCoinTurnstile\n" +
				"Initial:Locked\n" +
				"{\n" +
				"  Locked Coin Unlocked {alarmOff unlock}\n" +
				"  Locked Pass Locked alarmOn\n" +
				"  Unlocked Coin Unlocked thankyou\n" +
				"  Unlocked Pass Locked lock\n" +
				"}\n" +
				".\n",
		},
		{
			"two coins turnstile without super state",
			"" +
				"Actions: Turnstile\n" +
				"FSM: TwoCoinTurnstile\n" +
				"Initial: Locked\n" +
				"{\n" +
				"\tLocked {\n" +
				"\t\tPass\tAlarming\talarmOn\n" +
				"\t\tCoin\tFirstCoin\t*\n" +
				"\t\tReset\tLocked\t{lock alarmOff}\n" +
				"\t}\n" +
				"\t\n" +
				"\tAlarming\tReset\tLocked {lock alarmOff}\n" +
				"\t\n" +
				"\tFirstCoin {\n" +
				"\t\tPass\tAlarming\t*\n" +
				"\t\tCoin\tUnlocked\tunlock\n" +
				"\t\tReset\tLocked {lock alarmOff}\n" +
				"\t}\n" +
				"\t\n" +
				"\tUnlocked {\n" +
				"\t\tPass\tLocked\tlock\n" +
				"\t\tCoin\t*\t\tthankyou\n" +
				"\t\tReset\tLocked {lock alarmOff}\n" +
				"\t}\n" +
				"}",
			"" +
				"Actions:Turnstile\n" +
				"FSM:TwoCoinTurnstile\n" +
				"Initial:Locked\n" +
				"{\n" +
				"  Locked {\n" +
				"    Pass Alarming alarmOn\n" +
				"    Coin FirstCoin {}\n" +
				"    Reset Locked {lock alarmOff}\n" +
				"  }\n" +
				"  Alarming Reset Locked {lock alarmOff}\n" +
				"  FirstCoin {\n" +
				"    Pass Alarming {}\n" +
				"    Coin Unlocked unlock\n" +
				"    Reset Locked {lock alarmOff}\n" +
				"  }\n" +
				"  Unlocked {\n" +
				"    Pass Locked lock\n" +
				"    Coin * thankyou\n" +
				"    Reset Locked {lock alarmOff}\n" +
				"  }\n" +
				"}\n" +
				".\n",
		},
		{
			"two coins turnstile with super state",
			"" +
				"Actions: Turnstile\n" +
				"FSM: TwoCoinTurnstile\n" +
				"Initial: Locked\n" +
				"{\n" +
				"    (Base)\tReset\tLocked\tlock\n" +
				"\n" +
				"\tLocked : Base {\n" +
				"\t\tPass\tAlarming\t*\n" +
				"\t\tCoin\tFirstCoin\t*\n" +
				"\t}\n" +
				"\t\n" +
				"\tAlarming : Base\t<alarmOn >alarmOff *\t*\t*\n" +
				"\t\n" +
				"\tFirstCoin : Base {\n" +
				"\t\tPass\tAlarming\t*\n" +
				"\t\tCoin\tUnlocked\tunlock\n" +
				"\t}\n" +
				"\t\n" +
				"\tUnlocked : Base {\n" +
				"\t\tPass\tLocked\tlock\n" +
				"\t\tCoin\t*\t\tthankyou\n" +
				"\t}\n" +
				"}",
			"" +
				"Actions:Turnstile\n" +
				"FSM:TwoCoinTurnstile\n" +
				"Initial:Locked\n" +
				"{\n" +
				"  (Base) Reset Locked lock\n" +
				"  Locked:Base {\n" +
				"    Pass Alarming {}\n" +
				"    Coin FirstCoin {}\n" +
				"  }\n" +
				"  Alarming:Base <alarmOn >alarmOff * * {}\n" +
				"  FirstCoin:Base {\n" +
				"    Pass Alarming {}\n" +
				"    Coin Unlocked unlock\n" +
				"  }\n" +
				"  Unlocked:Base {\n" +
				"    Pass Locked lock\n" +
				"    Coin * thankyou\n" +
				"  }\n" +
				"}\n" +
				".\n",
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			got := parseSource(testCase.source)
			if got != testCase.expected {
				t.Fatalf("expected '%s' for %s, but got '%s'", testCase.expected, testCase.source, got)
			}
		})
	}
}

func parseSource(source string) string {
	syntaxBuilder := NewFsmSyntaxBuilder()
	parser := NewParser(syntaxBuilder)
	lexer := lexer.New(parser)
	lexer.Lex(source)
	parser.HandleEvent("EOF", -1, -1)

	return syntaxBuilder.String()
}
