package semanticanalyzer

import "reflect"

type superClassCrawler struct {
	concreteState    SemanticState
	transitionTuples map[string]transitionTuple
	ssm              *SemanticStateMachine
}

func newSuperClassCrawler(ssm *SemanticStateMachine) *superClassCrawler {
	return &superClassCrawler{
		transitionTuples: make(map[string]transitionTuple),
		ssm:              ssm,
	}
}

func (sc *superClassCrawler) checkSuperClassTransitions() {
	values := []SemanticState{}
	for _, value := range sc.ssm.States {
		values = append(values, *value)
	}
	for _, state := range values {
		if !state.AbstractState {
			sc.concreteState = state
			// sc.transitionTuples = make(map[string]transitionTuple)
			sc.checkTransitionsForState(&sc.concreteState)
		}
	}
}

func (sc *superClassCrawler) checkTransitionsForState(state *SemanticState) {
	for superState := range state.SuperStates {
		sc.checkTransitionsForState(superState)
	}
	sc.checkStateForPreviouslyDefinedTransition(state)
}

func (sc *superClassCrawler) checkStateForPreviouslyDefinedTransition(state *SemanticState) {
	for _, semanticTrasition := range state.Transitions {
		sc.checkTransitionForPreviousDefinition(state, semanticTrasition)
	}
}

func (sc *superClassCrawler) checkTransitionForPreviousDefinition(
	state *SemanticState,
	semanticTransition SemanticTransition,
) {
	thisTuple := newTransitionTuple(
		state.Name, semanticTransition.Event,
		semanticTransition.NextState.Name, semanticTransition.Action,
	)
	if _, ok := sc.transitionTuples[thisTuple.event]; ok {
		sc.determineIfthePreviousDefinitionIsAnError(state, *thisTuple)
	} else {
		sc.transitionTuples[thisTuple.event] = *thisTuple
	}
}

func (sc *superClassCrawler) determineIfthePreviousDefinitionIsAnError(
	state *SemanticState,
	thisTuple transitionTuple,
) {
	previousTuple := sc.transitionTuples[thisTuple.event]
	if !sc.transitionsHaveSameOutcomes(thisTuple, previousTuple) {
		sc.checkForOverriddenTransition(state, thisTuple, previousTuple)
	}
}

func (sc *superClassCrawler) transitionsHaveSameOutcomes(t1, t2 transitionTuple) bool {
	return t1.nextState == t2.nextState && reflect.DeepEqual(t1.actions, t2.actions)
}

func (sc *superClassCrawler) checkForOverriddenTransition(
	state *SemanticState,
	thisTuple,
	previousTuple transitionTuple,
) {
	definingState := sc.ssm.States[previousTuple.currentState]
	if !isSuperStateOf(definingState, state) {
		sc.ssm.Errors = append(
			sc.ssm.Errors,
			*NewAnalysisErrorWithExtra(CONFLICTING_SUPERSTATES, sc.concreteState.Name+"|"+thisTuple.event),
		)
	} else {
		sc.transitionTuples[thisTuple.event] = thisTuple
	}
}

func isSuperStateOf(possibleSuperState, state *SemanticState) bool {
	if reflect.DeepEqual(possibleSuperState, state) {
		return true
	}
	for superState := range state.SuperStates {
		if isSuperStateOf(possibleSuperState, superState) {
			return true
		}
	}
	return false
}

type transitionTuple struct {
	currentState string
	event        string
	nextState    string
	actions      []string
}

func newTransitionTuple(
	currentState, event, nextState string, actions []string,
) *transitionTuple {
	return &transitionTuple{
		currentState: currentState,
		event:        event,
		nextState:    nextState,
		actions:      actions,
	}
}
