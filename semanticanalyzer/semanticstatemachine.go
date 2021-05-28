package semanticanalyzer

import (
	"fmt"
	"sort"
)

type SemanticStateMachine struct {
	Errors       []AnalysisError
	Warnings     []AnalysisError
	States       map[string]*SemanticState
	Events       map[string]bool
	Actions      map[string]bool
	InitialState SemanticState
	ActionClass  string
	FsmName      string
}

func NewSemanticStateMachine() *SemanticStateMachine {
	return &SemanticStateMachine{
		Errors:   []AnalysisError{},
		Warnings: []AnalysisError{},
		States:   make(map[string]*SemanticState),
		Events:   make(map[string]bool),
		Actions:  make(map[string]bool),
	}
}

func (ssm *SemanticStateMachine) addError(analysisError *AnalysisError) {
	ssm.Errors = append(ssm.Errors, *analysisError)
}

func (ssm *SemanticStateMachine) String() string {
	return fmt.Sprintf(
		""+
			"Actions: %s\n"+
			"FSM: %s\n"+
			"Initial: %s"+
			"%s",
		ssm.ActionClass, ssm.FsmName, ssm.InitialState.Name, ssm.statesToString(),
	)
}

func (ssm *SemanticStateMachine) statesToString() string {
	statesString := "{"
	semanticStates := []SemanticState{}
	for _, value := range ssm.States {
		semanticStates = append(semanticStates, *value)
	}
	sort.Slice(semanticStates, func(i, j int) bool {
		return semanticStates[i].Name < semanticStates[j].Name
	})

	for _, semanticState := range semanticStates {
		statesString += semanticState.String()
	}

	return statesString + "}\n"
}

type SemanticState struct {
	Name          string
	EntryActions  []string
	ExitActions   []string
	AbstractState bool
	SuperStates   map[*SemanticState]bool
	Transitions   []SemanticTransition
}

func NewSemanticState(name string) *SemanticState {
	return &SemanticState{Name: name, SuperStates: make(map[*SemanticState]bool)}
}

func (ss *SemanticState) String() string {
	return fmt.Sprintf("\n  %s {\n%s  }\n", ss.makeStateNameWithAdornments(), ss.makeTransitionsString())
}

func (ss *SemanticState) makeStateNameWithAdornments() string {
	stateName := ""
	if ss.AbstractState {
		stateName += "(" + ss.Name + ")"
	} else {
		stateName += ss.Name
	}

	sortedSuperStates := []*SemanticState{}
	for superState := range ss.SuperStates {
		sortedSuperStates = append(sortedSuperStates, superState)
	}
	sort.Slice(sortedSuperStates, func(i, j int) bool {
		return sortedSuperStates[i].Name < sortedSuperStates[j].Name
	})
	for _, superState := range sortedSuperStates {
		stateName += " :" + superState.Name
	}
	for _, entryAction := range ss.EntryActions {
		stateName += " <" + entryAction
	}
	for _, exitAction := range ss.ExitActions {
		stateName += " >" + exitAction
	}
	return stateName
}

func (ss *SemanticState) makeTransitionsString() string {
	transitionString := ""
	for _, semanticTransition := range ss.Transitions {
		transitionString += ss.makeTransitionString(&semanticTransition)
	}

	return transitionString
}

func (ss *SemanticState) makeTransitionString(st *SemanticTransition) string {
	return fmt.Sprintf("    %s %s {%s}\n", st.Event, ss.makeNextStateName(st), makeActions(st))
}

func (ss *SemanticState) makeNextStateName(st *SemanticTransition) string {
	if ss.Name == "" {
		return "null"
	}
	return st.NextState.Name
}

func makeActions(st *SemanticTransition) string {
	actions := ""
	firstAction := true
	for _, action := range st.Action {
		if firstAction {
			firstAction = false
		} else {
			actions += " "
		}
		actions += action
	}

	return actions
}

type ErrorId string

const (
	NO_FSM                            ErrorId = "NO_FSM"
	NO_INITIAL                        ErrorId = "NO_INITIAL"
	INVALID_HEADER                    ErrorId = "INVALID_HEADER"
	EXTRA_HEADER_IGNORED              ErrorId = "EXTRA_HEADER_IGNORED"
	UNDEFINED_STATE                   ErrorId = "UNDEFINED_STATE"
	UNDEFINED_SUPER_STATE             ErrorId = "UNDEFINED_SUPER_STATE"
	UNUSED_STATE                      ErrorId = "UNUSED_STATE"
	DUPLICATE_TRANSITION              ErrorId = "DUPLICATE_TRANSITION"
	ABSTRACT_STATE_USED_AS_NEXT_STATE ErrorId = "ABSTRACT_STATE_USED_AS_NEXT_STATE"
	INCONSISTENT_ABSTRACTION          ErrorId = "INCONSISTENT_ABSTRACTION"
	STATE_ACTIONS_MULTIPLY_DEFINED    ErrorId = "STATE_ACTIONS_MULTIPLY_DEFINED"
	CONFLICTING_SUPERSTATES           ErrorId = "CONFLICTING_SUPERSTATES"
)

type AnalysisError struct {
	errorId ErrorId
	extra   string
}

func NewAnalysisError(errorId ErrorId) *AnalysisError {
	return &AnalysisError{
		errorId: errorId,
		extra:   "",
	}
}

func NewAnalysisErrorWithExtra(errorId ErrorId, extra string) *AnalysisError {
	return &AnalysisError{
		errorId: errorId,
		extra:   extra,
	}
}

type SemanticTransition struct {
	Event     string
	NextState *SemanticState
	Action    []string
}
