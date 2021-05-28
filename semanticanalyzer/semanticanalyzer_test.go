package semanticanalyzer

import (
	"reflect"
	"testing"

	"github.com/larkvincer/dsl-fsm/lexer"
	"github.com/larkvincer/dsl-fsm/parser"
)

type semanticanalyzerTest struct {
	name              string
	source            string
	expectedErrors    []AnalysisError
	notExpectedErrors []AnalysisError
}

var emptyErrors = []AnalysisError{}

func TestHeaderParsing(t *testing.T) {
	testTable := []semanticanalyzerTest{
		{"no headers", "{}", []AnalysisError{*NewAnalysisError(NO_FSM), *NewAnalysisError(NO_INITIAL)}, emptyErrors},
		{"missing actions", "FSM:f Initial:i {}", emptyErrors,
			[]AnalysisError{*NewAnalysisError(NO_FSM), *NewAnalysisError(NO_INITIAL)},
		},
		{"missing fsm", "actions:a Initial:i {}",
			[]AnalysisError{*NewAnalysisError(NO_FSM)},
			[]AnalysisError{*NewAnalysisError(NO_INITIAL)},
		},
		{"missing initial", "actions:a FSM:f {}",
			[]AnalysisError{*NewAnalysisError(NO_INITIAL)},
			[]AnalysisError{*NewAnalysisError(NO_FSM)},
		},
		{"nothing is missing", "Initial:f Actions:a Fsm:f {}", emptyErrors, emptyErrors},
		{"unexpected header", "X: x{s * * *}", []AnalysisError{
			*NewAnalysisErrorWithExtra(INVALID_HEADER, (&parser.Header{Name: "X", Value: "x"}).String()),
			*NewAnalysisError(NO_FSM),
			*NewAnalysisError(NO_INITIAL),
		}, emptyErrors},
		{"duplicated header", "fsm:f fsm:x {s * * *}", []AnalysisError{
			*NewAnalysisErrorWithExtra(EXTRA_HEADER_IGNORED, (&parser.Header{Name: "fsm", Value: "x"}).String()),
			*NewAnalysisError(NO_INITIAL),
		}, emptyErrors},
		{"initial state is undefined", "Initial: i {s * * *}",
			[]AnalysisError{*NewAnalysisErrorWithExtra(UNDEFINED_STATE, "initial: i")},
			emptyErrors,
		},
	}

	runSemanticTests(t, testTable)
}

func TestStateErrors(t *testing.T) {
	testTable := []semanticanalyzerTest{
		{"null next state is not undefined", "{s * * *}", emptyErrors, []AnalysisError{*NewAnalysisError(UNDEFINED_STATE)}},
		{"undefined state", "{s * s2 *}", []AnalysisError{*NewAnalysisErrorWithExtra(UNDEFINED_STATE, "s2")}, emptyErrors},
		{"no undefined states", "{s * s *}", emptyErrors, []AnalysisError{*NewAnalysisErrorWithExtra(UNDEFINED_STATE, "s2")}},
		{"undefined super state", "{s:ss * * *}",
			[]AnalysisError{*NewAnalysisErrorWithExtra(UNDEFINED_SUPER_STATE, "ss")},
			emptyErrors,
		},
		{"super state defined", "{ss * * * s:ss * * *}",
			emptyErrors,
			[]AnalysisError{*NewAnalysisErrorWithExtra(UNDEFINED_SUPER_STATE, "s2")},
		},
		{"unused states", "{s e n *}", []AnalysisError{*NewAnalysisErrorWithExtra(UNUSED_STATE, "s")}, emptyErrors},
		{"no unused states", "{s e s *}", emptyErrors, []AnalysisError{*NewAnalysisErrorWithExtra(UNUSED_STATE, "s")}},
		{"next state is null is in implicit use", "{s e * *}", emptyErrors, []AnalysisError{*NewAnalysisErrorWithExtra(UNUSED_STATE, "s")}},
		{"used as a base is valid usage", "{b e n * s:b e2 s *}",
			emptyErrors,
			[]AnalysisError{*NewAnalysisErrorWithExtra(UNUSED_STATE, "b")},
		},
		{"used as initial is valid usage", "initial: b {b e n *}",
			emptyErrors,
			[]AnalysisError{*NewAnalysisErrorWithExtra(UNUSED_STATE, "b")},
		},
		{"error if super states have conflicting transactions",
			"" +
				"FSM: f Actions: act Initial: s" +
				"{" +
				"  (ss1) e1 s1 *" +
				"  (ss2) e1 s2 *" +
				"  s :ss1 :ss2 e2 s3 a" +
				"  s2 e s *" +
				"  s1 e s *" +
				"  s3 e s *" +
				"}",
			[]AnalysisError{*NewAnalysisErrorWithExtra(CONFLICTING_SUPERSTATES, "s|e1")},
			emptyErrors,
		},
		{"no error for overridden transition",
			"" +
				"FSM: f Actions: act Initial: s" +
				"{" +
				"  (ss1) e1 s1 *" +
				"  s :ss1 e1 s3 a" +
				"  s1 e s *" +
				"  s3 e s *" +
				"}",
			emptyErrors,
			[]AnalysisError{*NewAnalysisErrorWithExtra(CONFLICTING_SUPERSTATES, "s|e1")},
		},
		{"no error if super states have different actions in same transitions",
			"" +
				"FSM: f Actions: act Initial: s" +
				"{" +
				"  (ss1) e1 s1 ax" +
				"  (ss2) e1 s1 ax" +
				"  s :ss1 :ss2 e2 s3 a" +
				"  s1 e s *" +
				"  s3 e s *" +
				"}",
			emptyErrors,
			[]AnalysisError{*NewAnalysisErrorWithExtra(CONFLICTING_SUPERSTATES, "s|e1")},
		},
		{"no error if super states have identical transitions",
			"" +
				"FSM: f Actions: act Initial: s" +
				"{" +
				"  (ss1) e1 s1 ax" +
				"  (ss2) e1 s1 ax" +
				"  s :ss1 :ss2 e2 s3 a" +
				"  s1 e s *" +
				"  s3 e s *" +
				"}",
			emptyErrors,
			[]AnalysisError{*NewAnalysisErrorWithExtra(CONFLICTING_SUPERSTATES, "s|e1")},
		},
		{"error if super states have different actions in same transitions",
			"" +
				"FSM: f Actions: act Initial: s" +
				"{" +
				"  (ss1) e1 s1 a1" +
				"  (ss2) e1 s1 a2" +
				"  s :ss1 :ss2 e2 s3 a" +
				"  s1 e s *" +
				"  s3 e s *" +
				"}",
			[]AnalysisError{*NewAnalysisErrorWithExtra(CONFLICTING_SUPERSTATES, "s|e1")},
			emptyErrors,
		},
	}

	runSemanticTests(t, testTable)
}

func TestTransitionErrors(t *testing.T) {
	testTable := []semanticanalyzerTest{
		{"duplicate transitions", "{s e * * s e * *}",
			[]AnalysisError{*NewAnalysisErrorWithExtra(DUPLICATE_TRANSITION, "s(e)")},
			emptyErrors,
		},
		{"no duplicate transitions", "{s e * *}",
			emptyErrors,
			[]AnalysisError{*NewAnalysisErrorWithExtra(DUPLICATE_TRANSITION, "s(e)")},
		},
		{"abstract states can not be target", "{(as) e * * s e as *}",
			[]AnalysisError{*NewAnalysisErrorWithExtra(ABSTRACT_STATE_USED_AS_NEXT_STATE, "s(e)->as")},
			emptyErrors,
		},
		{"entry and exit actions are not multiply defined",
			"" +
				"{" +
				"  s * * * " +
				"  s * * *" +
				"  es * * *" +
				"  es <x * * * " +
				"  es <x * * *" +
				"  xs >x * * *" +
				"  xs >{x} * * *" +
				"}",
			emptyErrors,
			[]AnalysisError{
				*NewAnalysisErrorWithExtra(STATE_ACTIONS_MULTIPLY_DEFINED, "s"),
				*NewAnalysisErrorWithExtra(STATE_ACTIONS_MULTIPLY_DEFINED, "es"),
				*NewAnalysisErrorWithExtra(STATE_ACTIONS_MULTIPLY_DEFINED, "xs"),
			},
		},
		{"error if state has multiple entry action definitions",
			"{s * * * ds <x * * * ds <y * * *}",
			[]AnalysisError{
				*NewAnalysisErrorWithExtra(STATE_ACTIONS_MULTIPLY_DEFINED, "ds"),
			},
			[]AnalysisError{
				*NewAnalysisErrorWithExtra(STATE_ACTIONS_MULTIPLY_DEFINED, "s"),
			},
		},
		{"error if state has multiple exit action definitions",
			"{ds >x * * * ds >y * *}",
			[]AnalysisError{
				*NewAnalysisErrorWithExtra(STATE_ACTIONS_MULTIPLY_DEFINED, "ds"),
			},
			emptyErrors,
		},
		{"error if state has multiple defined entry and exit actions",
			"{ds >x * * * ds >y * *}",
			[]AnalysisError{
				*NewAnalysisErrorWithExtra(STATE_ACTIONS_MULTIPLY_DEFINED, "ds"),
			},
			emptyErrors,
		},
	}

	runSemanticTests(t, testTable)
}

func TestWarnings(t *testing.T) {
	testTable := []semanticanalyzerTest{
		{
			"warn if state used as both abstract and concrete",
			"{(ias) e * * ias e * * (cas) e * *}",
			[]AnalysisError{*NewAnalysisErrorWithExtra(INCONSISTENT_ABSTRACTION, "ias")},
			[]AnalysisError{*NewAnalysisErrorWithExtra(INCONSISTENT_ABSTRACTION, "cas")},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			got := produceSemanticStateMachine(testCase.source).Warnings
			if len(testCase.expectedErrors) > 0 && notContains(got, testCase.expectedErrors) {
				t.Fatalf("expected '%v' for %s, but got '%v'", testCase.expectedErrors, testCase.source, got)
			}
			if len(testCase.notExpectedErrors) > 0 && contains(got, testCase.notExpectedErrors) {
				t.Fatalf("does not expect '%v' for %s, but got '%v'", testCase.notExpectedErrors, testCase.source, got)
			}
		})
	}
}

func TestFSMElements(t *testing.T) {

	t.Run("states", func(t *testing.T) {
		type stateTest struct {
			name           string
			source         string
			expectedStates []SemanticState
		}

		testTable := []stateTest{
			{
				"one state",
				"{s * * *}",
				[]SemanticState{
					*NewSemanticState("s"),
				},
			},
			{
				"many states",
				"{s1 * * * s2 * * * s3 * * *}",
				[]SemanticState{
					*NewSemanticState("s1"),
					*NewSemanticState("s2"),
					*NewSemanticState("s3"),
				},
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				statesMap := produceSemanticStateMachine(testCase.source).States
				gottenStates := []SemanticState{}
				for _, value := range statesMap {
					gottenStates = append(gottenStates, *value)
				}
				if !containsStates(gottenStates, testCase.expectedStates) {
					t.Fatalf("expected '%v' for %s, but got '%v'", testCase.expectedStates, testCase.source, gottenStates)
				}
			})
		}
	})

	t.Run("events", func(t *testing.T) {
		type eventTests struct {
			name           string
			source         string
			expectedEvents []string
		}

		testTable := []eventTests{
			{
				"many events",
				"{s1 e1 * * s2 e2 * * s3 e3 * *}",
				[]string{"e1", "e2", "e3"},
			},
			{
				"many events, but no duplicates",
				"{s1 e1 * * s2 e2 * * s3 e1 * *}",
				[]string{"e1", "e2"},
			},
			{
				"no null events",
				"{(s1) * * *}",
				[]string{},
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				got := produceSemanticStateMachine(testCase.source).Events
				gottenEvents := []string{}
				for event := range got {
					gottenEvents = append(gottenEvents, event)
				}

				if !containsStrings(gottenEvents, testCase.expectedEvents) {
					t.Fatalf("expected '%v' for %s, but got '%v'", testCase.expectedEvents, testCase.source, gottenEvents)
				}
			})
		}
	})

	t.Run("actions", func(t *testing.T) {
		type eventTests struct {
			name            string
			source          string
			expectedActions []string
		}

		testTable := []eventTests{
			{
				"many actions but no duplicates",
				"{s1 e1 * {a1 a2} s2 e2 * {a3 a1}}",
				[]string{"a1", "a2", "a3"},
			},
			{
				"entry and exit actions are counted as actions",
				"{s <ea >xa * * a}",
				[]string{"ea", "xa"},
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				got := produceSemanticStateMachine(testCase.source).Actions
				gottenActions := []string{}
				for action := range got {
					gottenActions = append(gottenActions, action)
				}

				if !containsStrings(gottenActions, testCase.expectedActions) {
					t.Fatalf("expected '%v' for %s, but got '%v'", testCase.expectedActions, testCase.source, gottenActions)
				}
			})
		}
	})
}

func TestLogic(t *testing.T) {
	type logicTest struct {
		name                    string
		source                  string
		expectAstRepresentation string
	}

	testTable := []logicTest{
		{
			"one transition",
			"{s e s a}",
			"" +
				"{\n" +
				"  s {\n" +
				"    e s {a}\n" +
				"  }\n" +
				"}\n",
		},
		{
			"two transitions are aggregated",
			"{s e1 s a s e2 s a}",
			"" +
				"{\n" +
				"  s {\n" +
				"    e1 s {a}\n" +
				"    e2 s {a}\n" +
				"  }\n" +
				"}\n",
		},
		{
			"super states are aggregated",
			"{s:b1 e1 s a s:b2 e2 s a (b1) e s * (b2) e s *}",
			"" +
				"{\n" +
				"  (b1) {\n" +
				"    e s {}\n" +
				"  }\n" +
				"\n" +
				"  (b2) {\n" +
				"    e s {}\n" +
				"  }\n" +
				"\n" +
				"  s :b1 :b2 {\n" +
				"    e1 s {a}\n" +
				"    e2 s {a}\n" +
				"  }\n" +
				"}\n",
		},
		{
			"null next state refer to itself",
			"{s e * a}",
			"" +
				"{\n" +
				"  s {\n" +
				"    e s {a}\n" +
				"  }\n" +
				"}\n",
		},
		{
			"actions remain in the same order",
			"{s e s {some text here}}",
			"" +
				"{\n" +
				"  s {\n" +
				"    e s {some text here}\n" +
				"  }\n" +
				"}\n",
		},
		{
			"entry and exit actions remain in order",
			"{s <{d o} <g >{c a} >t e s a}",
			"" +
				"{\n" +
				"  s <d <o <g >c >a >t {\n" +
				"    e s {a}\n" +
				"  }\n" +
				"}\n",
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			got := produceSemanticStateMachine(attachHeader(testCase.source)).statesToString()
			if got != testCase.expectAstRepresentation {
				t.Fatalf("expected '%s' for %s, but got '%s'", testCase.expectAstRepresentation, testCase.source, got)
			}
		})
	}
}

func attachHeader(states string) string {
	return "initial: s fsm: f actions:a " + states
}

func TestIntegration(t *testing.T) {
	type integrationTest struct {
		name     string
		source   string
		expected string
	}

	testTable := []integrationTest{
		{
			"subway turnstile one",
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
				"Actions: Turnstile\n" +
				"FSM: OneCoinTurnstile\n" +
				"Initial: Locked{\n" +
				"  Locked {\n" +
				"    Coin Unlocked {alarmOff unlock}\n" +
				"    Pass Locked {alarmOn}\n" +
				"  }\n" +
				"\n" +
				"  Unlocked {\n" +
				"    Coin Unlocked {thankyou}\n" +
				"    Pass Locked {lock}\n" +
				"  }\n" +
				"}\n",
		},
		{
			"subway turnstile two",
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
				"Actions: Turnstile\n" +
				"FSM: TwoCoinTurnstile\n" +
				"Initial: Locked{\n" +
				"  Alarming {\n" +
				"    Reset Locked {lock alarmOff}\n" +
				"  }\n" +
				"\n" +
				"  FirstCoin {\n" +
				"    Pass Alarming {}\n" +
				"    Coin Unlocked {unlock}\n" +
				"    Reset Locked {lock alarmOff}\n" +
				"  }\n" +
				"\n" +
				"  Locked {\n" +
				"    Pass Alarming {alarmOn}\n" +
				"    Coin FirstCoin {}\n" +
				"    Reset Locked {lock alarmOff}\n" +
				"  }\n" +
				"\n" +
				"  Unlocked {\n" +
				"    Pass Locked {lock}\n" +
				"    Coin Unlocked {thankyou}\n" +
				"    Reset Locked {lock alarmOff}\n" +
				"  }\n" +
				"}\n",
		},
		// {
		// 	"subway turnstile three",
		// 	"" +
		// 		"Actions: Turnstile\n" +
		// 		"FSM: TwoCoinTurnstile\n" +
		// 		"Initial: Locked\n" +
		// 		"{\n" +
		// 		"    (Base)\tReset\tLocked\tlock\n" +
		// 		"\n" +
		// 		"\tLocked : Base {\n" +
		// 		"\t\tPass\tAlarming\t*\n" +
		// 		"\t\tCoin\tFirstCoin\t*\n" +
		// 		"\t}\n" +
		// 		"\t\n" +
		// 		"\tAlarming : Base\t<alarmOn >alarmOff *\t*\t*\n" +
		// 		"\t\n" +
		// 		"\tFirstCoin : Base {\n" +
		// 		"\t\tPass\tAlarming\t*\n" +
		// 		"\t\tCoin\tUnlocked\tunlock\n" +
		// 		"\t}\n" +
		// 		"\t\n" +
		// 		"\tUnlocked : Base {\n" +
		// 		"\t\tPass\tLocked\tlock\n" +
		// 		"\t\tCoin\t*\t\tthankyou\n" +
		// 		"\t}\n" +
		// 		"}",
		// 	"" +
		// 		"Actions: Turnstile\n" +
		// 		"FSM: TwoCoinTurnstile\n" +
		// 		"Initial: Locked{\n" +
		// 		"  Alarming :Base <alarmOn >alarmOff {\n" +
		// 		"    null Alarming {}\n" +
		// 		"  }\n" +
		// 		"\n" +
		// 		"  (Base) {\n" +
		// 		"    Reset Locked {lock}\n" +
		// 		"  }\n" +
		// 		"\n" +
		// 		"  FirstCoin :Base {\n" +
		// 		"    Pass Alarming {}\n" +
		// 		"    Coin Unlocked {unlock}\n" +
		// 		"  }\n" +
		// 		"\n" +
		// 		"  Locked :Base {\n" +
		// 		"    Pass Alarming {}\n" +
		// 		"    Coin FirstCoin {}\n" +
		// 		"  }\n" +
		// 		"\n" +
		// 		"  Unlocked :Base {\n" +
		// 		"    Pass Locked {lock}\n" +
		// 		"    Coin Unlocked {thankyou}\n" +
		// 		"  }\n" +
		// 		"}\n",
		// },
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			got := produceSemanticStateMachine(testCase.source)
			if !assertSemanticResult(testCase.source, testCase.expected) {
				t.Fatalf("expected '%s' for %s, but got '%s'", testCase.expected, testCase.source, got)
			}
		})
	}
}

func assertSemanticResult(source, expected string) bool {
	ast := produceSemanticStateMachine(source)
	got := ast.String()
	return got == expected
}

func runSemanticTests(t *testing.T, testTable []semanticanalyzerTest) {
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			got := produceSemanticStateMachine(testCase.source).Errors
			if len(testCase.expectedErrors) > 0 && notContains(got, testCase.expectedErrors) {
				t.Fatalf("expected '%v' for %s, but got '%v'", testCase.expectedErrors, testCase.source, got)
			}
			if len(testCase.notExpectedErrors) > 0 && contains(got, testCase.notExpectedErrors) {
				t.Fatalf("does not expect '%v' for %s, but got '%v'", testCase.notExpectedErrors, testCase.source, got)
			}
		})
	}

}

func produceSemanticStateMachine(source string) *SemanticStateMachine {
	syntaxBuilder := parser.NewFsmSyntaxBuilder()
	parser := parser.NewParser(syntaxBuilder)
	lexer := lexer.New(parser)
	lexer.Lex(source)
	parser.HandleEvent("EOF", -1, -1)
	analyzer := New()

	return analyzer.Analyze(syntaxBuilder.GetFSM())
}

func notContains(target, toCheck []AnalysisError) bool {
	return !contains(target, toCheck)
}

func contains(target, toCheck []AnalysisError) bool {
	if len(toCheck) == 0 {
		return true
	}
	for _, error := range toCheck {
		found := false
		for _, targetError := range target {
			if reflect.DeepEqual(error, targetError) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func containsStates(target, toCheck []SemanticState) bool {
	for _, expectedState := range toCheck {
		found := false
		for _, actualState := range target {
			if actualState.Name == expectedState.Name {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func containsStrings(target, toCheck []string) bool {
	for _, expectedString := range toCheck {
		found := false
		for _, actualString := range target {
			if expectedString == actualString {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
