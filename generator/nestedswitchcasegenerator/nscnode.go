package nscgenerator

type NSCNode interface {
	Accept(visitor *NSCNodeVisitor)
}

type SwitchCaseNode struct {
	variableName string
	caseNodes    []NSCNode
}

func NewSwitchCaseNode(variableName string) *SwitchCaseNode {
	return &SwitchCaseNode{
		variableName: variableName,
	}
}

func (scn *SwitchCaseNode) Accept(visitor *NSCNodeVisitor) {
	(*visitor).VisitSwitchCaseNode(scn)
}

func (scn *SwitchCaseNode) GenerateCases(visitor NSCNodeVisitor) {
	for _, cn := range scn.caseNodes {
		cn.Accept(&visitor)
	}
}

type CaseNode struct {
	switchName     string
	caseName       string
	caseActionNode NSCNode
}

func NewCaseNode(switchName, caseName string) *CaseNode {
	return &CaseNode{
		switchName: switchName,
		caseName:   caseName,
	}
}

func (cn *CaseNode) Accept(visitor *NSCNodeVisitor) {
	(*visitor).VisitCaseNode(cn)
}

type FunctionCallNode struct {
	functionName string
	argument     NSCNode
}

func NewFunctionCallNode(functionName string, argument NSCNode) *FunctionCallNode {
	return &FunctionCallNode{
		functionName: functionName,
		argument:     argument,
	}
}

func (fcn *FunctionCallNode) Accept(visitor *NSCNodeVisitor) {
	(*visitor).VisitFunctionalCallNode(fcn)
}

type CompositeNode struct {
	nodes []NSCNode
}

func (cn *CompositeNode) Accept(visitor *NSCNodeVisitor) {
	for _, node := range cn.nodes {
		node.Accept(visitor)
	}
}

func (cn *CompositeNode) Add(node NSCNode) {
	cn.nodes = append(cn.nodes, node)
}

type EnumNode struct {
	name        string
	enumerators []string
}

func NewEnumNode(name string, enumerators []string) *EnumNode {
	return &EnumNode{
		name:        name,
		enumerators: enumerators,
	}
}

func (en *EnumNode) Accept(visitor *NSCNodeVisitor) {
	(*visitor).VisitEnumNode(en)
}

type StatePropertyNode struct {
	initialState string
}

func NewStatePropertyNode(initialState string) *StatePropertyNode {
	return &StatePropertyNode{
		initialState: initialState,
	}
}

func (spn *StatePropertyNode) Accept(visitor *NSCNodeVisitor) {
	(*visitor).VisitStatePropertyNode(spn)
}

type EventDelegatorsNode struct {
	events []string
}

func NewEventDelegatorsNode(events []string) *EventDelegatorsNode {
	return &EventDelegatorsNode{
		events: events,
	}
}

func (evn *EventDelegatorsNode) Accept(visitor *NSCNodeVisitor) {
	(*visitor).VisitEventDelegatorsNode(evn)
}

type FSMClassNode struct {
	delegators    *EventDelegatorsNode
	eventEnum     *EnumNode
	stateEnum     *EnumNode
	stateProperty *StatePropertyNode
	handleEvent   *HandleEventNode
	className     string
	actionsName   string
	actions       []string
}

func (fsmcn *FSMClassNode) Accept(visitor *NSCNodeVisitor) {
	(*visitor).VisitFSMClassNode(fsmcn)
}

type HandleEventNode struct {
	switchCase *SwitchCaseNode
}

func NewHandleEventNode(switchCaseNode *SwitchCaseNode) *HandleEventNode {
	return &HandleEventNode{
		switchCase: switchCaseNode,
	}
}

func (hen *HandleEventNode) Accept(visitor *NSCNodeVisitor) {
	(*visitor).VisitHandleEventNode(hen)
}

type EnumeratorNode struct {
	enumeration string
	enumerator  string
}

func NewEnumeratorNode(enumeration, enumerator string) *EnumeratorNode {
	return &EnumeratorNode{
		enumeration: enumeration,
		enumerator:  enumerator,
	}
}

func (en *EnumeratorNode) Accept(visitor *NSCNodeVisitor) {
	(*visitor).VisitEnumeratorNode(en)
}

type DefaultCaseNode struct {
	state string
}

func NewDefaultCaseNode(state string) *DefaultCaseNode {
	return &DefaultCaseNode{
		state: state,
	}
}

func (dcn *DefaultCaseNode) Accept(visitor *NSCNodeVisitor) {
	(*visitor).VisitDefaultCaseNode(dcn)
}
