package main

import "github.com/larkvincer/dsl-fsm/lexer"

func main() {
	const sourceInput = `Initial: Locked
FSM: Turnstile
{
Locked    Coin    Unlocked    unlock
Locked    Pass    Locked      alarm
Unlocked  Coin    Unlocked    thankyou
Unlocked  Pass    Locked      lock
}`
	testCollector := lexer.NewTestCollector()
	l := lexer.New(testCollector)
	l.Lex(sourceInput)
}
