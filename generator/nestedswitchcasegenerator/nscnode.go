package nscgenerator

type NSCNode interface {
	Accept(visitor NSCNodeVisitor)
}

type SwitchCaseNode struct {
	VariableName string
	CaseNodes    []NSCNode
}

func NewSwitchCaseNode(variableName string) *SwitchCaseNode {
	return &SwitchCaseNode{
		VariableName: variableName,
	}
}

func (scn *SwitchCaseNode) Accept(visitor NSCNodeVisitor) {
	visitor.VisitSwitchCaseNode(scn)
}

func (scn *SwitchCaseNode) GenerateCases(visitor NSCNodeVisitor) {
	for _, cn := range scn.CaseNodes {
		cn.Accept(visitor)
	}
}

type CaseNode struct {
	SwitchName     string
	CaseName       string
	CaseActionNode NSCNode
}

func NewCaseNode(switchName, caseName string) *CaseNode {
	return &CaseNode{
		SwitchName: switchName,
		CaseName:   caseName,
	}
}

func (cn *CaseNode) Accept(visitor NSCNodeVisitor) {
	visitor.VisitCaseNode(cn)
}

type FunctionCallNode struct {
	FunctionName string
	Argument     NSCNode
}

func NewFunctionCallNode(functionName string, argument NSCNode) *FunctionCallNode {
	return &FunctionCallNode{
		FunctionName: functionName,
		Argument:     argument,
	}
}

func (fcn *FunctionCallNode) Accept(visitor NSCNodeVisitor) {
	visitor.VisitFunctionalCallNode(fcn)
}

type CompositeNode struct {
	nodes []NSCNode
}

func (cn *CompositeNode) Accept(visitor NSCNodeVisitor) {
	for _, node := range cn.nodes {
		node.Accept(visitor)
	}
}

func (cn *CompositeNode) Add(node NSCNode) {
	cn.nodes = append(cn.nodes, node)
}

type EnumNode struct {
	Name        string
	Enumerators []string
}

func NewEnumNode(name string, enumerators []string) *EnumNode {
	return &EnumNode{
		Name:        name,
		Enumerators: enumerators,
	}
}

func (en *EnumNode) Accept(visitor NSCNodeVisitor) {
	visitor.VisitEnumNode(en)
}

type StatePropertyNode struct {
	InitialState string
}

func NewStatePropertyNode(initialState string) *StatePropertyNode {
	return &StatePropertyNode{
		InitialState: initialState,
	}
}

func (spn *StatePropertyNode) Accept(visitor NSCNodeVisitor) {
	visitor.VisitStatePropertyNode(spn)
}

type EventDelegatorsNode struct {
	Events []string
}

func NewEventDelegatorsNode(events []string) *EventDelegatorsNode {
	return &EventDelegatorsNode{
		Events: events,
	}
}

func (evn *EventDelegatorsNode) Accept(visitor NSCNodeVisitor) {
	visitor.VisitEventDelegatorsNode(evn)
}

type FSMClassNode struct {
	Delegators    *EventDelegatorsNode
	EventEnum     *EnumNode
	StateEnum     *EnumNode
	StateProperty *StatePropertyNode
	HandleEvent   *HandleEventNode
	ClassName     string
	ActionsName   string
	Actions       []string
}

func (fsmcn *FSMClassNode) Accept(visitor NSCNodeVisitor) {
	visitor.VisitFSMClassNode(fsmcn)
}

type HandleEventNode struct {
	SwitchCase *SwitchCaseNode
}

func NewHandleEventNode(switchCaseNode *SwitchCaseNode) *HandleEventNode {
	return &HandleEventNode{
		SwitchCase: switchCaseNode,
	}
}

func (hen *HandleEventNode) Accept(visitor NSCNodeVisitor) {
	visitor.VisitHandleEventNode(hen)
}

type EnumeratorNode struct {
	Enumeration string
	Enumerator  string
}

func NewEnumeratorNode(enumeration, enumerator string) *EnumeratorNode {
	return &EnumeratorNode{
		Enumeration: enumeration,
		Enumerator:  enumerator,
	}
}

func (en *EnumeratorNode) Accept(visitor NSCNodeVisitor) {
	visitor.VisitEnumeratorNode(en)
}

type DefaultCaseNode struct {
	state string
}

func NewDefaultCaseNode(state string) *DefaultCaseNode {
	return &DefaultCaseNode{
		state: state,
	}
}

func (dcn *DefaultCaseNode) Accept(visitor NSCNodeVisitor) {
	visitor.VisitDefaultCaseNode(dcn)
}
