package parser

import "fmt"

type FsmSyntax struct {
	Headers []header
	Logic   []*transition
	Errors  []syntaxError
	Done    bool
}

type header struct {
	Name  string
	Value string
}

func nullHeader() header {
	return header{"", ""}
}

type transition struct {
	State          stateSpec
	SubTransitions []subTransition
}

type subTransition struct {
	Event     string
	NextState string
	Actions   []string
}

type stateSpec struct {
	Name          string
	SuperStates   []string
	EntryActions  []string
	ExitActions   []string
	AbstractState bool
}

type syntaxError struct {
	Type       string
	Message    string
	LineNumber int
	Position   int
}

func (fsmSyntax *FsmSyntax) String() (result string) {
	result = fsmSyntax.formatHeaders() + fsmSyntax.formatLogic()
	if fsmSyntax.Done {
		result += ".\n"
	} else {
		result += fsmSyntax.formatErrors()
	}

	return
}

func (fsmSyntax *FsmSyntax) GetErrors() string {
	return fsmSyntax.formatErrors()
}

func (fsmSyntax *FsmSyntax) formatHeaders() string {
	formattedHeaders := ""
	for _, h := range fsmSyntax.Headers {
		formattedHeaders += formatHeader(&h)
	}
	return formattedHeaders
}

func (fsmSyntax *FsmSyntax) formatLogic() string {
	if len(fsmSyntax.Logic) > 0 {
		return fmt.Sprintf("{\n%s}\n", fsmSyntax.formatTransitions())
	}
	return ""
}

func (fsmSyntax *FsmSyntax) formatTransitions() string {
	transitions := ""
	for _, trans := range fsmSyntax.Logic {
		transitions += formatTransition(trans)
	}
	return transitions
}

func formatTransition(trans *transition) string {
	return fmt.Sprintf("  %s %s\n", formatStateName(trans.State), formatSubTransitions(trans))
}

func (fsmSyntax *FsmSyntax) formatErrors() string {
	if len(fsmSyntax.Errors) > 0 {
		return formatError(fsmSyntax.Errors[0])
	}
	return ""
}

func formatError(error syntaxError) string {
	return fmt.Sprintf("Syntax error: %s. %s. line %d, position %d.\n", error.Type, error.Message, error.LineNumber, error.Position)
}

func formatHeader(h *header) string {
	return fmt.Sprintf("%s:%s\n", h.Name, h.Value)
}

func formatStateName(state stateSpec) string {
	stateName := fmt.Sprintf(getStateFormatter(state.AbstractState), state.Name)
	for _, superState := range state.SuperStates {
		stateName += ":" + superState
	}
	for _, entryAction := range state.EntryActions {
		stateName += " <" + entryAction
	}
	for _, exitAction := range state.ExitActions {
		stateName += " >" + exitAction
	}
	return stateName
}
func getStateFormatter(isAbstract bool) string {
	if isAbstract {
		return "(%s)"
	}
	return "%s"
}

func formatSubTransitions(trans *transition) string {
	if len(trans.SubTransitions) == 1 {
		return formatSubTransition(trans.SubTransitions[0])
	}

	formattedSubTransitions := "{\n"
	for _, subTrans := range trans.SubTransitions {
		formattedSubTransitions += "    " + formatSubTransition(subTrans) + "\n"
	}

	return formattedSubTransitions + "  }"
}

func formatSubTransition(subTrans subTransition) string {
	return fmt.Sprintf("%s %s %s", formatEventOrState(subTrans.Event), formatEventOrState(subTrans.NextState), formatActions(subTrans))
}

func formatEventOrState(eventOrState string) string {
	if eventOrState == "" {
		return "*"
	} else {
		return eventOrState
	}
}

func formatActions(subTrans subTransition) string {
	if len(subTrans.Actions) == 1 {
		return subTrans.Actions[0]
	} else {
		actions := "{"
		first := true
		for _, action := range subTrans.Actions {
			if first {
				actions += action
				first = false
			} else {
				actions += " " + action
			}
		}

		return actions + "}"
	}
}
