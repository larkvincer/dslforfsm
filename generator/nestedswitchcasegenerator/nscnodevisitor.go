package nscgenerator

type NSCNodeVisitor interface {
	VisitSwitchCaseNode(switchCaseNode *SwitchCaseNode)
	VisitCaseNode(caseNode *CaseNode)
	VisitFunctionalCallNode(functionCallNode *FunctionCallNode)
	VisitEnumNode(enumNode *EnumNode)
	VisitStatePropertyNode(statePropertyNode *StatePropertyNode)
	VisitEventDelegatorsNode(eventDelegatorsNode *EventDelegatorsNode)
	VisitFSMClassNode(fsmClassNode *FSMClassNode)
	VisitHandleEventNode(handleEventNode *HandleEventNode)
	VisitEnumeratorNode(enumeratorNode *EnumeratorNode)
	VisitDefaultCaseNode(defaultCaseNode *DefaultCaseNode)
}
