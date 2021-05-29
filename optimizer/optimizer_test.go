package optimizer

import (
	"testing"

	"github.com/larkvincer/dsl-fsm/lexer"
	"github.com/larkvincer/dsl-fsm/parser"
	"github.com/larkvincer/dsl-fsm/semanticanalyzer"
)

func TestBasicOptimizationsFunctions(t *testing.T) {
	t.Run("header", func(t *testing.T) {

		osm := produceStateMachineWithHeader("{i e i *}")
		if osm.header.Fsm != "f" {
			t.Fatalf("expected header 'f', but got %s", osm.header.Fsm)
		}
		if osm.header.Initial != "i" {
			t.Fatalf("expected header 'i', but got %s", osm.header.Initial)
		}
		if osm.header.Actions != "a" {
			t.Fatalf("expected header 'a', but got %s", osm.header.Actions)
		}
	})

	t.Run("state as preserved", func(t *testing.T) {
		osm := produceStateMachineWithHeader("{i e s * s e i *}")
		if !contains(osm.states, "i", "s") {
			t.Fatalf("expected to have i, s states, but got %s", osm.states)
		}
	})

	t.Run("abstract states are removed", func(t *testing.T) {
		osm := produceStateMachineWithHeader("{(b) * * * i:b e i *}")
		if contains(osm.states, "b") {
			t.Fatalf("not expect abstact state 'b', but got %s", osm.states)
		}
	})

	t.Run("events are preserved", func(t *testing.T) {
		osm := produceStateMachineWithHeader("{i e1 s * s e2 i *}")
		if !contains(osm.events, "e1", "e2") {
			t.Fatalf("expect events 'e1', 'e2', but got %s", osm.events)
		}
	})

	t.Run("actions are preserved", func(t *testing.T) {
		osm := produceStateMachineWithHeader("{i e1 s a1 s e2 i a2}")
		if !contains(osm.actions, "a1", "a2") {
			t.Fatalf("expect actions 'a1', 'a2', but got %s", osm.actions)
		}
	})

	t.Run("simple state machine", func(t *testing.T) {
		source := "{i e i a1}"
		osm := produceStateMachineWithHeader(source)
		assertOptimization(
			t,
			source,
			""+
				"i {\n"+
				"  e i {a1}\n"+
				"}\n",
		)
		if len(osm.transitions) != 1 {
			t.Fatalf("expect to have 1 transition, but got %d", len(osm.transitions))
		}
	})
}

func TestEntryAndExitActions(t *testing.T) {
	type entryExitActionsTest struct {
		name     string
		source   string
		expected string
	}

	testTable := []entryExitActionsTest{
		{
			"entry functions added",
			"" +
				"{" +
				"  i e s a1" +
				"  i e2 s a2" +
				"  s <n1 <n2 e i *" +
				"}",
			"" +
				"i {\n" +
				"  e s {n1 n2 a1}\n" +
				"  e2 s {n1 n2 a2}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n",
		},
		{
			"exit functions added",
			"" +
				"{" +
				"  i >x2 >x1 e s a1" +
				"  i e2 s a2" +
				"  s e i *" +
				"}",
			"" +
				"i {\n" +
				"  e s {x2 x1 a1}\n" +
				"  e2 s {x2 x1 a2}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n",
		},
		{
			"first super state enry and exit actions added",
			"" +
				"{" +
				"  (ib) >ibx1 >ibx2 * * *" +
				"  (sb) <sbn1 <sbn2 * * *" +
				"  i:ib >x e s a" +
				"  s:sb <n e i *" +
				"}",
			"" +
				"i {\n" +
				"  e s {x ibx1 ibx2 sbn1 sbn2 n a}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n",
		},
		{
			"multiple super states entry and exit actions added",
			"" +
				"{" +
				"  (ib1) >ib1x * * *" +
				"  (ib2) : ib1 >ib2x * * *" +
				"  (sb1) <sb1n* * *" +
				"  (sb2) :sb1 <sb2n* * *" +
				"  i:ib2 >x e s a" +
				"  s:sb2 <n e i *" +
				"}",
			"" +
				"i {\n" +
				"  e s {x ib2x ib1x sb1n sb2n n a}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n",
		},
		{
			"diamond super states entry and exit actions added",
			"" +
				"{" +
				"  (ib1) >ib1x * * *" +
				"  (ib2) : ib1 >ib2x * * *" +
				"  (ib3) : ib1 >ib3x * * *" +
				"  (sb1) <sb1n * * *" +
				"  (sb2) :sb1 <sb2n * * *" +
				"  (sb3) :sb1 <sb3n * * *" +
				"  i:ib2 :ib3 >x e s a" +
				"  s :sb2 :sb3 <n e i *" +
				"}",
			"" +
				"i {\n" +
				"  e s {x ib3x ib2x ib1x sb1n sb2n sb3n n a}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n",
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			assertOptimization(
				t,
				testCase.source,
				testCase.expected,
			)
		})
	}
}

func TestSuperStatesTransition(t *testing.T) {
	type superStatesTransitionTest struct {
		name     string
		source   string
		expected string
	}

	testTable := []superStatesTransitionTest{
		{
			"simple inheritance of transitions",
			"" +
				"{" +
				"  (b) be s ba" +
				"  i:b e s a" +
				"  s e i *" +
				"}",
			"" +
				"i {\n" +
				"  e s {a}\n" +
				"  be s {ba}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n",
		},
		{
			"deep inheritance of transitions",
			"" +
				"{" +
				"  (b1) {" +
				"    b1e1 s b1a1" +
				"    b1e2 s b1a2" +
				"  }" +
				"  (b2):b1 b2e s b2a" +
				"  i:b2 e s a" +
				"  s e i *" +
				"}",
			"" +
				"i {\n" +
				"  e s {a}\n" +
				"  b2e s {b2a}\n" +
				"  b1e1 s {b1a1}\n" +
				"  b1e2 s {b1a2}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n",
		},
		{
			"multiple inheritance of transitions",
			"" +
				"{" +
				"  (b1) b1e s b1a" +
				"  (b2) b2e s b2a" +
				"  i:b1 :b2 e s a" +
				"  s e i *" +
				"}",
			"" +
				"i {\n" +
				"  e s {a}\n" +
				"  b2e s {b2a}\n" +
				"  b1e s {b1a}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n",
		},
		{
			"multiple super states entry and exit actions added",
			"" +
				"{" +
				"  (ib1) >ib1x * * *" +
				"  (ib2) : ib1 >ib2x * * *" +
				"  (sb1) <sb1n* * *" +
				"  (sb2) :sb1 <sb2n* * *" +
				"  i:ib2 >x e s a" +
				"  s:sb2 <n e i *" +
				"}",
			"" +
				"i {\n" +
				"  e s {x ib2x ib1x sb1n sb2n n a}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n",
		},
		{
			"diamond inheritance of transitions",
			"" +
				"{" +
				"  (b) be s ba" +
				"  (b1):b b1e s b1a" +
				"  (b2):b b2e s b2a" +
				"  i:b1 :b2 e s a" +
				"  s e i *" +
				"}",
			"" +
				"i {\n" +
				"  e s {a}\n" +
				"  b2e s {b2a}\n" +
				"  b1e s {b1a}\n" +
				"  be s {ba}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n",
		},
		{
			"overriding transitions",
			"" +
				"{" +
				"  (b) e s2 a2" +
				"  i:b e s a" +
				"  s e i *" +
				"  s2 e i *" +
				"}",
			"" +
				"i {\n" +
				"  e s {a}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n" +
				"s2 {\n" +
				"  e i {}\n" +
				"}\n",
		},
		{
			"elimination of duplicate transitions",
			"" +
				"{" +
				"  (b) e s a" +
				"  i:b e s a" +
				"  s e i *" +
				"}",
			"" +
				"i {\n" +
				"  e s {a}\n" +
				"}\n" +
				"s {\n" +
				"  e i {}\n" +
				"}\n",
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			assertOptimization(
				t,
				testCase.source,
				testCase.expected,
			)
		})
	}
}

func TestAcceptance(t *testing.T) {
	const source = "" +
		"Actions: Turnstile\n" +
		"FSM: TwoCoinTurnstile\n" +
		"Initial: Locked\n" +
		"{" +
		"    (Base)  Reset  Locked  lock" +
		"" +
		"  Locked : Base {" +
		"    Pass  Alarming  *" +
		"    Coin  FirstCoin *" +
		"  }" +
		"" +
		"  Alarming : Base <alarmOn >alarmOff *  *  *" +
		"" +
		"  FirstCoin : Base {" +
		"    Pass  Alarming  *" +
		"    Coin  Unlocked  unlock" +
		"  }" +
		"" +
		"  Unlocked : Base {" +
		"    Pass  Locked  lock" +
		"    Coin  *       thankyou" +
		"}"

	osm := produceStateMachine(source)

	const expected = "" +
		"Initial: Locked\n" +
		"Fsm: TwoCoinTurnstile\n" +
		"Actions:Turnstile\n" +
		"{\n" +
		"  Alarming {\n" +
		"    Reset Locked {alarmOff lock}\n" +
		"  }\n" +
		"  FirstCoin {\n" +
		"    Pass Alarming {alarmOn}\n" +
		"    Coin Unlocked {unlock}\n" +
		"    Reset Locked {lock}\n" +
		"  }\n" +
		"  Locked {\n" +
		"    Pass Alarming {alarmOn}\n" +
		"    Coin FirstCoin {}\n" +
		"    Reset Locked {lock}\n" +
		"  }\n" +
		"  Unlocked {\n" +
		"    Pass Locked {lock}\n" +
		"    Coin Unlocked {thankyou}\n" +
		"    Reset Locked {lock}\n" +
		"  }\n" +
		"}\n"

	if osm.String() != expected {
		t.Fatalf("expected %s, but got %s, for %s", expected, osm.String(), source)
	}
}

func produceStateMachineWithHeader(body string) OptimizedStateMachine {
	source := "fsm:f initial:i actions:a " + body
	return produceStateMachine(source)
}

func produceStateMachine(source string) OptimizedStateMachine {
	syntaxBuilder := parser.NewFsmSyntaxBuilder()
	parser := parser.NewParser(syntaxBuilder)
	lexer := lexer.New(parser)
	lexer.Lex(source)
	parser.HandleEvent("EOF", -1, -1)
	analyzer := semanticanalyzer.New()
	semanticStateMachine := analyzer.Analyze(syntaxBuilder.GetFSM())
	return *Optimize(*semanticStateMachine)
}

func assertOptimization(t *testing.T, fsmBody, expected string) {
	optimizedStateMachine := produceStateMachineWithHeader(fsmBody)
	transitions := optimizedStateMachine.transitionsToString()
	if removeWhiteSpaces(transitions) != removeWhiteSpaces(expected) {
		t.Fatalf("expected '%s' for '%s', but got '%s'", expected, fsmBody, transitions)
	}
}

func contains(target []string, toCheck ...string) bool {
	for _, itemToCheck := range toCheck {
		found := false
		for _, targetItem := range target {
			if itemToCheck == targetItem {
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

func removeWhiteSpaces(source string) string {
	return source
	// result := source
	// strings.ReplaceAll(result, "\\n", "\n")
	// strings.ReplaceAll(result, "[\t ]", " ")
	// strings.ReplaceAll(result, "\\n", "\n")
	// return result
}
