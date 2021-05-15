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
	const sourceInput = "{ s e ns a }"
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
