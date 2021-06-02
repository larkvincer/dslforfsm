package main

import (
	"fmt"

	"github.com/larkvincer/dsl-fsm/generator"
	"github.com/larkvincer/dsl-fsm/generator/implementors"
	"github.com/larkvincer/dsl-fsm/lexer"
	"github.com/larkvincer/dsl-fsm/optimizer"
	"github.com/larkvincer/dsl-fsm/parser"
	"github.com/larkvincer/dsl-fsm/semanticanalyzer"
)

func main() {
	const sourceInput = `Initial: Locked
	Actions: Turnstile
	FSM: TurnstileFSM
	{
	  Locked    {
	    Coin    Unlocked    unlock
	    Pass    Locked      alarm
	  }
	  Unlocked  {
	    Coin    Unlocked    thankyou
	    Pass    Locked      lock
	  }
	}`
	syntaxBuilder := parser.NewFsmSyntaxBuilder()
	parser := parser.NewParser(syntaxBuilder)
	lexer := lexer.New(parser)
	lexer.Lex(sourceInput)
	parser.HandleEvent("EOF", -1, -1)
	fsm := syntaxBuilder.GetFSM()

	if len(fsm.Errors) == 0 {
		semanticStateMachine := semanticanalyzer.New().Analyze(fsm)
		optimizedStateMachine := optimizer.Optimize(*semanticStateMachine)
		flags := make(map[string]string)
		flags["package"] = "firsttry"
		javaImplementor := implementors.NewJavaNestedSwitchCaseImplementor(flags)
		javaCodeGenerator := generator.NewJavaCodeGenerator(javaImplementor)
		codeGenerator := generator.NewCodeGenerator(optimizedStateMachine, javaCodeGenerator)
		codeGenerator.Generate()
	} else {
		fmt.Print(fsm.Errors)
	}
}
