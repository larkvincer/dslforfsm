package semanticanalyzer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/larkvincer/dsl-fsm/parser"
)

type SemanticAnalyzer struct {
	semanticStateMachine *SemanticStateMachine
	fsmHeader            parser.Header
	actionsHeader        parser.Header
	initialHeader        parser.Header
}

func New() *SemanticAnalyzer {
	return &SemanticAnalyzer{
		semanticStateMachine: NewSemanticStateMachine(),
		fsmHeader:            parser.Header{Name: "", Value: ""},
		actionsHeader:        parser.Header{Name: "", Value: ""},
		initialHeader:        parser.Header{Name: "", Value: ""},
	}
}

func (sa *SemanticAnalyzer) Analyze(fsmSyntax *parser.FsmSyntax) *SemanticStateMachine {
	sa.semanticStateMachine = NewSemanticStateMachine()
	sa.analyzeHeaders(fsmSyntax)
	sa.checkSemanticValidity(fsmSyntax)
	sa.produceSemanticStateMachine(fsmSyntax)

	return sa.semanticStateMachine
}

func (sa *SemanticAnalyzer) analyzeHeaders(fsmSyntax *parser.FsmSyntax) {
	sa.setHeaders(fsmSyntax)
	sa.checkMissingHeaders()
}

func (sa *SemanticAnalyzer) setHeaders(fsmSyntax *parser.FsmSyntax) {
	for _, header := range fsmSyntax.Headers {
		if isNamed(header, "fsm") {
			sa.setHeader(&sa.fsmHeader, header)
		} else if isNamed(header, "actions") {
			sa.setHeader(&sa.actionsHeader, header)
		} else if isNamed(header, "initial") {
			sa.setHeader(&sa.initialHeader, header)
		} else {
			sa.semanticStateMachine.addError(NewAnalysisErrorWithExtra(INVALID_HEADER, header.String()))
		}
	}
}

func (sa *SemanticAnalyzer) setHeader(targetHeader *parser.Header, header parser.Header) {
	if isNullHeader(targetHeader) {
		targetHeader.Name = header.Name
		targetHeader.Value = header.Value
	} else {
		sa.semanticStateMachine.addError(NewAnalysisErrorWithExtra(EXTRA_HEADER_IGNORED, header.String()))
	}
}

func (sa *SemanticAnalyzer) checkMissingHeaders() {
	if isNullHeader(&sa.fsmHeader) {
		sa.semanticStateMachine.addError(NewAnalysisError(NO_FSM))
	}
	if isNullHeader(&sa.initialHeader) {
		sa.semanticStateMachine.addError(NewAnalysisError(NO_INITIAL))
	}
}

func (sa *SemanticAnalyzer) checkSemanticValidity(fsmSyntax *parser.FsmSyntax) {
	sa.createStateEventAndActionLists(fsmSyntax)
	sa.checkUndefinedStates(fsmSyntax)
	sa.checkForUnusedStates(fsmSyntax)
	sa.checkForDuplicateTransitions(fsmSyntax)
	sa.checkThatAbstractStatesAreNotTargets(fsmSyntax)
	sa.checkForInconsistentAbstraction(fsmSyntax)
	sa.checkForMultiplyDefinedStateActions(fsmSyntax)
}

func (sa *SemanticAnalyzer) checkForInconsistentAbstraction(fsmSyntax *parser.FsmSyntax) {
	abstractStates := sa.findAbstractStates(fsmSyntax)
	for _, transition := range fsmSyntax.Logic {
		if _, ok := abstractStates[transition.State.Name]; !transition.State.AbstractState && ok {
			sa.semanticStateMachine.Warnings = append(
				sa.semanticStateMachine.Warnings,
				*NewAnalysisErrorWithExtra(INCONSISTENT_ABSTRACTION, transition.State.Name),
			)
		}
	}
}

func (sa *SemanticAnalyzer) checkForMultiplyDefinedStateActions(fsmSyntax *parser.FsmSyntax) {
	firstActionsForState := make(map[string]string)
	for _, transition := range fsmSyntax.Logic {
		if specifiesStateActions(transition) {
			actionsKey := makeActionsKey(transition)
			if _, ok := firstActionsForState[transition.State.Name]; ok {
				if ac := firstActionsForState[transition.State.Name]; ac != actionsKey {
					sa.semanticStateMachine.Errors = append(
						sa.semanticStateMachine.Errors,
						*NewAnalysisErrorWithExtra(STATE_ACTIONS_MULTIPLY_DEFINED, transition.State.Name),
					)
				}
			} else {
				firstActionsForState[transition.State.Name] = actionsKey
			}
		}
	}
}

func (sa *SemanticAnalyzer) checkThatAbstractStatesAreNotTargets(fsmSyntax *parser.FsmSyntax) {
	abstractStates := sa.findAbstractStates(fsmSyntax)

	for _, transition := range fsmSyntax.Logic {
		for _, subTransition := range transition.SubTransitions {
			if _, ok := abstractStates[subTransition.NextState]; ok {
				sa.semanticStateMachine.Errors = append(
					sa.semanticStateMachine.Errors,
					*NewAnalysisErrorWithExtra(
						ABSTRACT_STATE_USED_AS_NEXT_STATE,
						fmt.Sprintf("%s(%s)->%s", transition.State.Name, subTransition.Event, subTransition.NextState),
					),
				)
			}
		}
	}
}

func (sa *SemanticAnalyzer) findAbstractStates(fsmSyntax *parser.FsmSyntax) map[string]bool {
	abstractStates := make(map[string]bool)

	for _, transition := range fsmSyntax.Logic {
		if transition.State.AbstractState {
			abstractStates[transition.State.Name] = true
		}
	}

	return abstractStates
}

func (sa *SemanticAnalyzer) checkForDuplicateTransitions(fsmSyntax *parser.FsmSyntax) {
	transitionKeys := make(map[string]bool)
	for _, transition := range fsmSyntax.Logic {
		for _, subTransition := range transition.SubTransitions {
			key := fmt.Sprintf("%s(%s)", transition.State.Name, subTransition.Event)
			if _, ok := transitionKeys[key]; ok {
				sa.semanticStateMachine.Errors = append(
					sa.semanticStateMachine.Errors,
					*NewAnalysisErrorWithExtra(DUPLICATE_TRANSITION, key),
				)
			} else {
				transitionKeys[key] = true
			}
		}
	}
}

func (sa *SemanticAnalyzer) checkForUnusedStates(fsmSyntax *parser.FsmSyntax) {
	sa.findStatesDefinedButNotUsed(sa.findUsedStates(fsmSyntax))
}

func (sa *SemanticAnalyzer) findStatesDefinedButNotUsed(usedStates map[string]bool) {
	definedStates := make([]string, 0, len(sa.semanticStateMachine.States))
	for key := range sa.semanticStateMachine.States {
		definedStates = append(definedStates, key)
	}
	sort.Strings(definedStates)

	for _, definedState := range definedStates {
		if _, ok := usedStates[definedState]; !ok {
			sa.semanticStateMachine.Errors = append(
				sa.semanticStateMachine.Errors,
				*NewAnalysisErrorWithExtra(UNUSED_STATE, definedState),
			)
		}
	}
}

func (sa *SemanticAnalyzer) findUsedStates(fsmSyntax *parser.FsmSyntax) map[string]bool {
	usedStates := make(map[string]bool)
	if sa.initialHeader.Value != "" {
		usedStates[sa.initialHeader.Value] = true
	}

	for superState := range sa.getSuperStates(fsmSyntax) {
		usedStates[superState] = true
	}
	for nextState := range sa.getNextStates(fsmSyntax) {
		usedStates[nextState] = true
	}
	return usedStates
}

func (sa *SemanticAnalyzer) getSuperStates(fsmSyntax *parser.FsmSyntax) map[string]bool {
	superStates := make(map[string]bool)
	for _, transition := range fsmSyntax.Logic {
		for _, superState := range transition.State.SuperStates {
			superStates[superState] = true
		}
	}

	return superStates
}

func (sa *SemanticAnalyzer) getNextStates(fsmSyntax *parser.FsmSyntax) map[string]bool {
	nextStates := make(map[string]bool)
	for _, transition := range fsmSyntax.Logic {
		for _, subTransition := range transition.SubTransitions {
			if subTransition.NextState == "" {
				nextStates[transition.State.Name] = true
			} else {
				nextStates[subTransition.NextState] = true
			}
		}
	}

	return nextStates
}

func (sa *SemanticAnalyzer) createStateEventAndActionLists(fsmSyntax *parser.FsmSyntax) {
	sa.addStateNamesToStateList(fsmSyntax)

	sa.addEntryAndExitActionsToActionList(fsmSyntax)
	sa.addEventsToEventList(fsmSyntax)
	sa.addTransitionActionsToActionList(fsmSyntax)
}

func (sa *SemanticAnalyzer) addStateNamesToStateList(fsmSyntax *parser.FsmSyntax) {
	for _, transition := range fsmSyntax.Logic {
		state := NewSemanticState(transition.State.Name)
		sa.semanticStateMachine.States[state.Name] = state
	}
}

func (sa *SemanticAnalyzer) addEntryAndExitActionsToActionList(fsmSyntax *parser.FsmSyntax) {
	for _, transition := range fsmSyntax.Logic {
		for _, entryAction := range transition.State.EntryActions {
			sa.semanticStateMachine.Actions[entryAction] = true
		}

		for _, exitAction := range transition.State.ExitActions {
			sa.semanticStateMachine.Actions[exitAction] = true
		}
	}
}

func (sa *SemanticAnalyzer) addEventsToEventList(fsmSyntax *parser.FsmSyntax) {
	for _, transition := range fsmSyntax.Logic {
		for _, subTransition := range transition.SubTransitions {
			if subTransition.Event != "" {
				sa.semanticStateMachine.Events[subTransition.Event] = true
			}
		}
	}
}

func (sa *SemanticAnalyzer) addTransitionActionsToActionList(fsmSyntax *parser.FsmSyntax) {
	for _, transition := range fsmSyntax.Logic {
		for _, subTransition := range transition.SubTransitions {
			for _, action := range subTransition.Actions {
				sa.semanticStateMachine.Actions[action] = true
			}
		}
	}
}

func (sa *SemanticAnalyzer) checkUndefinedStates(fsmSyntax *parser.FsmSyntax) {
	for _, transition := range fsmSyntax.Logic {
		for _, superState := range transition.State.SuperStates {
			sa.checkUndefinedState(superState, UNDEFINED_SUPER_STATE)
		}

		for _, subTransition := range transition.SubTransitions {
			sa.checkUndefinedState(subTransition.NextState, UNDEFINED_STATE)
		}
	}

	if _, ok := sa.semanticStateMachine.States[sa.initialHeader.Value]; !ok && sa.initialHeader.Value != "" {
		sa.semanticStateMachine.Errors = append(sa.semanticStateMachine.Errors,
			*NewAnalysisErrorWithExtra(UNDEFINED_STATE, "initial: "+sa.initialHeader.Value),
		)
	}
}

func (sa *SemanticAnalyzer) checkUndefinedState(referenceState string, errorCode ErrorId) {
	if _, ok := sa.semanticStateMachine.States[referenceState]; !ok && referenceState != "" {
		sa.semanticStateMachine.Errors = append(
			sa.semanticStateMachine.Errors,
			*NewAnalysisErrorWithExtra(errorCode, referenceState),
		)
	}
}

func (sa *SemanticAnalyzer) produceSemanticStateMachine(fsmSyntax *parser.FsmSyntax) {
	if len(sa.semanticStateMachine.Errors) == 0 {
		sa.compileHeaders()
		for _, transition := range fsmSyntax.Logic {
			state := sa.compileState(transition)
			sa.compileTransitions(transition, state)
		}

		newSuperClassCrawler(sa.semanticStateMachine).checkSuperClassTransitions()
	}
}

func (sa *SemanticAnalyzer) compileHeaders() {
	sa.semanticStateMachine.InitialState = *sa.semanticStateMachine.States[sa.initialHeader.Value]
	sa.semanticStateMachine.ActionClass = sa.actionsHeader.Value
	sa.semanticStateMachine.FsmName = sa.fsmHeader.Value
}

func (sa *SemanticAnalyzer) compileState(transition *parser.FsmTransition) *SemanticState {
	state := sa.semanticStateMachine.States[transition.State.Name]
	state.EntryActions = append(state.EntryActions, transition.State.EntryActions...)
	state.ExitActions = append(state.ExitActions, transition.State.ExitActions...)
	state.AbstractState = state.AbstractState || transition.State.AbstractState
	for _, superStateName := range transition.State.SuperStates {
		superState := sa.semanticStateMachine.States[superStateName]
		state.SuperStates[superState] = true
	}
	return state
}

func (sa *SemanticAnalyzer) compileTransitions(transition *parser.FsmTransition, state *SemanticState) {
	for _, subTransition := range transition.SubTransitions {
		sa.compileTransition(state, &subTransition)
	}
}

func (sa *SemanticAnalyzer) compileTransition(state *SemanticState, subTransition *parser.SubTransition) {
	semanticTransition := SemanticTransition{}
	semanticTransition.Event = subTransition.Event
	if subTransition.NextState == "" {
		semanticTransition.NextState = state
	} else {
		semanticTransition.NextState = sa.semanticStateMachine.States[subTransition.NextState]
	}
	semanticTransition.Action = append(semanticTransition.Action, subTransition.Actions...)
	state.Transitions = append(state.Transitions, semanticTransition)
}

func isNamed(header parser.Header, headerName string) bool {
	return strings.Compare(strings.ToLower(headerName), strings.ToLower(header.Name)) == 0
}

func isNullHeader(header *parser.Header) bool {
	return header.Name == "" && header.Value == ""
}

func specifiesStateActions(transition *parser.FsmTransition) bool {
	return len(transition.State.EntryActions) != 0 || len(transition.State.ExitActions) != 0
}

func makeActionsKey(transition *parser.FsmTransition) string {
	actions := []string{}
	actions = append(actions, transition.State.EntryActions...)
	actions = append(actions, transition.State.ExitActions...)
	return commaList(actions)
}

func commaList(list []string) string {
	commaList := ""
	for _, s := range list {
		commaList += s + ","
	}
	return commaList[:len(commaList)-1]
}
