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
	fsm.ClassName = osm.Header.Fsm
	fsm.ActionsName = osm.Header.Actions
	fsm.StateEnum = nsc.stateEnumNode
	fsm.EventEnum = nsc.eventEnumNode
	fsm.Delegators = nsc.eventDelegatorsNode
	fsm.StateProperty = nsc.statePropertyNode
	fsm.HandleEvent = nsc.handleEventNode
	fsm.Actions = osm.Actions
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
	stateSwitch.CaseNodes = append(stateSwitch.CaseNodes, stateCaseNode)
}

func (nsc *NSCGenerator) addEventCases(
	stateCaseNode *CaseNode,
	transition *optimizer.Transition,
) {
	eventSwitch := NewSwitchCaseNode("event")
	stateCaseNode.CaseActionNode = eventSwitch
	for _, st := range transition.SubTransitions {
		nsc.addEventCase(eventSwitch, &st)
	}
	eventSwitch.CaseNodes = append(eventSwitch.CaseNodes, NewDefaultCaseNode(transition.CurrentState))
}

func (nsc *NSCGenerator) addEventCase(eventSwitch *SwitchCaseNode, st *optimizer.SubTransition) {
	eventCaseNode := NewCaseNode("Event", st.Event)
	nsc.addActions(st, eventCaseNode)
	eventSwitch.CaseNodes = append(eventSwitch.CaseNodes, eventCaseNode)
}

func (nsc *NSCGenerator) addActions(st *optimizer.SubTransition, eventCaseNode *CaseNode) {
	actions := &CompositeNode{}
	nsc.addSetStateNode(st.NextState, actions)
	for _, action := range st.Actions {
		functionCallNode := &FunctionCallNode{FunctionName: action}
		actions.Add(functionCallNode)
	}

	eventCaseNode.CaseActionNode = actions
}

func (nsc *NSCGenerator) addSetStateNode(stateName string, actions *CompositeNode) {
	enumeratorNode := NewEnumeratorNode("State", stateName)
	setStateNode := NewFunctionCallNode("setState", enumeratorNode)
	actions.Add(setStateNode)
}
