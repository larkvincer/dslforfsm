package nscgenerator

import "github.com/larkvincer/dsl-fsm/optimizer"

type NSCGenerator struct {
	stateEnumNode       *EnumNode
	eventEnumNode       *EnumNode
	eventDelegatorsNode *EventDelegatorsNode
	statePropertyNode   *StatePropertyNode
	handleEventNode     *HandleEventNode
	stateSwitch         *SwitchCaseNode
}

func (nsc *NSCGenerator) Generate(osm *optimizer.OptimizedStateMachine) *FSMClassNode {
	nsc.eventDelegatorsNode = NewEventDelegatorsNode(osm.Events)
	nsc.statePropertyNode = NewStatePropertyNode(osm.Header.Initial)
	nsc.stateEnumNode = NewEnumNode("State", osm.States)
	nsc.eventEnumNode = NewEnumNode("Event", osm.Events)
	nsc.stateSwitch = NewSwitchCaseNode("state")
	nsc.addStateCases(osm)
	nsc.handleEventNode = NewHandleEventNode(nsc.stateSwitch)
	return nsc.makeFsmNode(osm)
}

func (nsc *NSCGenerator) makeFsmNode(osm *optimizer.OptimizedStateMachine) *FSMClassNode {
	fsm := &FSMClassNode{}
	fsm.className = osm.Header.Fsm
	fsm.actionsName = osm.Header.Actions
	fsm.stateEnum = nsc.stateEnumNode
	fsm.eventEnum = nsc.eventEnumNode
	fsm.delegators = nsc.eventDelegatorsNode
	fsm.stateProperty = nsc.statePropertyNode
	fsm.handleEvent = nsc.handleEventNode
	fsm.actions = osm.Actions
	return fsm
}

func (nsc *NSCGenerator) addStateCases(osm *optimizer.OptimizedStateMachine) {
	for _, transition := range osm.Transitions {
		nsc.addStateCase(nsc.stateSwitch, &transition)
	}
}

func (nsc *NSCGenerator) addStateCase(
	stateSwitch *SwitchCaseNode,
	transition *optimizer.Transition,
) {
	stateCaseNode := NewCaseNode("State", transition.CurrentState)
	nsc.addEventCases(stateCaseNode, transition)
}

func (nsc *NSCGenerator) addEventCases(
	stateCaseNode *CaseNode,
	transition *optimizer.Transition,
) {
	eventSwitch := NewSwitchCaseNode("event")
	stateCaseNode.caseActionNode = eventSwitch
	for _, st := range transition.SubTransitions {
		nsc.addEventCase(eventSwitch, &st)
	}
	eventSwitch.caseNodes = append(eventSwitch.caseNodes, NewDefaultCaseNode(transition.CurrentState))
}

func (nsc *NSCGenerator) addEventCase(eventSwitch *SwitchCaseNode, st *optimizer.SubTransition) {
	eventCaseNode := NewCaseNode("Event", st.Event)
	nsc.addActions(st, eventCaseNode)
	eventSwitch.caseNodes = append(eventSwitch.caseNodes, eventCaseNode)
}

func (nsc *NSCGenerator) addActions(st *optimizer.SubTransition, eventCaseNode *CaseNode) {
	actions := &CompositeNode{}
	nsc.addSetStateNode(st.NextState, actions)
	for _, action := range st.Actions {
		functionCallNode := &FunctionCallNode{functionName: action}
		actions.Add(functionCallNode)
	}
}

func (nsc *NSCGenerator) addSetStateNode(stateName string, actions *CompositeNode) {
	enumeratorNode := NewEnumeratorNode("State", stateName)
	setStateNode := NewFunctionCallNode("setState", enumeratorNode)
	actions.Add(setStateNode)
}
