package optimizer

import (
	"reflect"

	"github.com/larkvincer/dsl-fsm/semanticanalyzer"
)

type Optimizer struct {
	optimizedStateMachine OptimizedStateMachine
	semanticStateMachine  semanticanalyzer.SemanticStateMachine
}

func New(semanticStateMachine semanticanalyzer.SemanticStateMachine) *Optimizer {
	return &Optimizer{
		semanticStateMachine: semanticStateMachine,
	}
}

func Optimize(ast semanticanalyzer.SemanticStateMachine) *OptimizedStateMachine {
	optimizer := Optimizer{
		semanticStateMachine:  ast,
		optimizedStateMachine: OptimizedStateMachine{},
	}

	optimizer.addHeader(ast)
	optimizer.addLists()
	optimizer.addTransitions()

	return &optimizer.optimizedStateMachine
}

func (optimizer *Optimizer) addHeader(ast semanticanalyzer.SemanticStateMachine) {
	optimizer.optimizedStateMachine.header = Header{}
	optimizer.optimizedStateMachine.header.Fsm = ast.FsmName
	optimizer.optimizedStateMachine.header.Initial = ast.InitialState.Name
	optimizer.optimizedStateMachine.header.Actions = ast.ActionClass
}

func (optimizer *Optimizer) addLists() {
	optimizer.addStates()
	optimizer.addEvents()
	optimizer.addActions()
}

func (optimizer *Optimizer) addStates() {
	for _, state := range optimizer.semanticStateMachine.States {
		if !state.AbstractState {
			optimizer.optimizedStateMachine.states = append(optimizer.optimizedStateMachine.states, state.Name)
		}
	}
}
func (optimizer *Optimizer) addEvents() {
	events := []string{}
	for event := range optimizer.semanticStateMachine.Events {
		events = append(events, event)
	}

	optimizer.optimizedStateMachine.events = append(
		optimizer.optimizedStateMachine.events,
		events...,
	)
}

func (optimizer *Optimizer) addActions() {
	actions := []string{}
	for action := range optimizer.semanticStateMachine.Actions {
		actions = append(actions, action)
	}

	optimizer.optimizedStateMachine.actions = append(
		optimizer.optimizedStateMachine.actions,
		actions...,
	)
}

func (optimizer *Optimizer) addTransitions() {
	for _, semanticState := range optimizer.semanticStateMachine.States {
		if !semanticState.AbstractState {
			NewStateOptimizer(optimizer, semanticState).addTransitionsForState()
		}
	}
}

func (optimizer *Optimizer) addAllStatesInHiearchyLeafFirst(
	semanticState *semanticanalyzer.SemanticState,
	hierarchy []*semanticanalyzer.SemanticState,
) []*semanticanalyzer.SemanticState {
	//Check this method, can be wrong
	for superState := range semanticState.SuperStates {
		contains := false
		for _, stateInHierarchy := range hierarchy {
			if reflect.DeepEqual(superState, stateInHierarchy) {
				contains = true
			}
		}

		if !contains {
			hierarchy = optimizer.addAllStatesInHiearchyLeafFirst(superState, hierarchy)
		}
	}
	return append(hierarchy, semanticState)
}

type StateOptimizer struct {
	optimizer          *Optimizer
	currentState       *semanticanalyzer.SemanticState
	eventsForThisState map[string]bool
}

func NewStateOptimizer(optimizer *Optimizer, currentState *semanticanalyzer.SemanticState) *StateOptimizer {
	return &StateOptimizer{
		optimizer:          optimizer,
		currentState:       currentState,
		eventsForThisState: make(map[string]bool),
	}
}

func (so *StateOptimizer) addTransitionsForState() {
	transition := Transition{}
	transition.CurrentState = so.currentState.Name
	so.addSubTransitions(&transition)
	so.optimizer.optimizedStateMachine.transitions = append(
		so.optimizer.optimizedStateMachine.transitions,
		transition,
	)
}

func (so *StateOptimizer) addSubTransitions(transition *Transition) {
	for _, stateInHierarchy := range so.makeRootFirstHierarchyOfStates() {
		so.addStateTransitions(transition, stateInHierarchy)
	}
}

func (so *StateOptimizer) addStateTransitions(transition *Transition, state *semanticanalyzer.SemanticState) {
	for _, semanticTransition := range state.Transitions {
		if so.eventExistsAndHasNotBeenOverridden(semanticTransition.Event) {
			so.addSubTransition(&semanticTransition, transition)
		}
	}
}

func (so *StateOptimizer) addSubTransition(
	semanticTransition *semanticanalyzer.SemanticTransition,
	transition *Transition,
) {
	so.eventsForThisState[semanticTransition.Event] = true
	subTransition := &SubTransition{}
	NewSubTransitionOptimizer(so, semanticTransition, subTransition).optimize()
	transition.SubTransitions = append(transition.SubTransitions, *subTransition)
}

func (so *StateOptimizer) makeRootFirstHierarchyOfStates() []*semanticanalyzer.SemanticState {
	hierarchy := []*semanticanalyzer.SemanticState{}
	hierarchy = so.optimizer.addAllStatesInHiearchyLeafFirst(so.currentState, hierarchy)
	reverserHierarchy(hierarchy)
	return hierarchy
}

func reverserHierarchy(hierarchy []*semanticanalyzer.SemanticState) []*semanticanalyzer.SemanticState {
	for i, j := 0, len(hierarchy)-1; i < j; i, j = i+1, j-1 {
		hierarchy[i], hierarchy[j] = hierarchy[j], hierarchy[i]
	}
	return hierarchy
}

func (so *StateOptimizer) eventExistsAndHasNotBeenOverridden(event string) bool {
	return event != "" && !so.eventsForThisState[event]
}

type SubTransitionOptimizer struct {
	stateOptimizer     *StateOptimizer
	semanticTransition *semanticanalyzer.SemanticTransition
	subTransition      *SubTransition
}

func NewSubTransitionOptimizer(
	stateOptimizer *StateOptimizer,
	semanticTransition *semanticanalyzer.SemanticTransition,
	subTransition *SubTransition,
) *SubTransitionOptimizer {
	return &SubTransitionOptimizer{
		stateOptimizer:     stateOptimizer,
		semanticTransition: semanticTransition,
		subTransition:      subTransition,
	}
}

func (sto *SubTransitionOptimizer) optimize() {
	sto.subTransition.Event = sto.semanticTransition.Event
	sto.subTransition.NextState = sto.semanticTransition.NextState.Name
	sto.addExitActions(sto.stateOptimizer.currentState)
	sto.addEntryActions(sto.semanticTransition.NextState)
	sto.subTransition.Actions = append(sto.subTransition.Actions, sto.semanticTransition.Action...)
}

func (sto *SubTransitionOptimizer) addExitActions(exitState *semanticanalyzer.SemanticState) {
	hierarchy := []*semanticanalyzer.SemanticState{}
	hierarchy = sto.stateOptimizer.optimizer.addAllStatesInHiearchyLeafFirst(exitState, hierarchy)
	reverserHierarchy(hierarchy)
	for _, superState := range hierarchy {
		sto.subTransition.Actions = append(sto.subTransition.Actions, superState.ExitActions...)
	}
}

func (sto *SubTransitionOptimizer) addEntryActions(entryState *semanticanalyzer.SemanticState) {
	hierarchy := []*semanticanalyzer.SemanticState{}
	hierarchy = sto.stateOptimizer.optimizer.addAllStatesInHiearchyLeafFirst(entryState, hierarchy)
	for _, superState := range hierarchy {
		sto.subTransition.Actions = append(sto.subTransition.Actions, superState.EntryActions...)
	}
}
