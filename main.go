package main

import (
	"fmt"

	// "reflect"

	"github.com/larkvincer/dsl-fsm/lexer"
	"github.com/larkvincer/dsl-fsm/parser"
)

// "fmt"

func main() {
	// const sourceInput = `Initial: Locked
	// 	FSM: Turnstile
	// 	{
	// 	Locked    Coin    Unlocked    unlock
	// 	Locked    Pass    Locked      alarm
	// 	Unlocked  Coin    Unlocked    thankyou
	// 	Unlocked  Pass    Locked      lock
	// }`

	// const sourceInput = `Initial: Locked
	// FSM: Turnstile
	// {
	//   Locked    {
	//     Coin    Unlocked    unlock
	//     Pass    Locked      alarm
	//   }
	//   Unlocked  {
	//     Coin    Unlocked    thankyou
	//     Pass    Locked      lock
	//   }
	// }`
	const sourceInput = "" +
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
		"}"
	syntaxBuilder := parser.NewFsmSyntaxBuilder()
	parser := parser.NewParser(syntaxBuilder)
	lexer := lexer.New(parser)
	// fmt.Println(reflect.ValueOf(*parser).FieldByName("state"))
	// testCollector := lexer.NewTestCollector()
	// l := lexer.New(testCollector)
	// l.Lex(sourceInput)
	lexer.Lex(sourceInput)
	parser.HandleEvent("EOF", -1, -1)
	// fmt.Printf("Result: %v", reflect.ValueOf(*parser).FieldByName("transitions"))
	fmt.Println(syntaxBuilder.String())
}
