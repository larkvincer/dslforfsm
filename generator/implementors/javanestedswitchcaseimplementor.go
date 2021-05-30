package implementors

import (
	"fmt"

	nscgenerator "github.com/larkvincer/dsl-fsm/generator/nestedswitchcasegenerator"
)

type JavaNestedSwitchCaseImplementor struct {
	Output      string
	flags       map[string]string
	javaPackage string
}

func NewJavaNestedSwitchCaseImplementor(flags map[string]string) *JavaNestedSwitchCaseImplementor {
	obj := &JavaNestedSwitchCaseImplementor{
		flags: flags,
	}
	if _, ok := flags["package"]; ok {
		obj.javaPackage = flags["package"]
	}
	return obj
}

func (javaImplementor *JavaNestedSwitchCaseImplementor) VisitSwitchCaseNode(
	switchCaseNode *nscgenerator.SwitchCaseNode,
) {
	javaImplementor.Output += fmt.Sprintf("switch(%s) {\n", switchCaseNode.VariableName)
	switchCaseNode.GenerateCases(javaImplementor)
	javaImplementor.Output += "}\n"
}

func (javaImplementor *JavaNestedSwitchCaseImplementor) VisitCaseNode(caseNode *nscgenerator.CaseNode) {
	javaImplementor.Output += fmt.Sprintf("case %s:\n", caseNode.CaseName)
	caseNode.CaseActionNode.Accept(javaImplementor)
	javaImplementor.Output += "break;\n"
}

func (javaImplementor *JavaNestedSwitchCaseImplementor) VisitFunctionalCallNode(
	functionCallNode *nscgenerator.FunctionCallNode,
) {
	javaImplementor.Output += fmt.Sprintf("%s(", functionCallNode.FunctionName)
	if functionCallNode.Argument != nil {
		functionCallNode.Argument.Accept(javaImplementor)
	}
	javaImplementor.Output += ");\n"
}

func (javaImplementor *JavaNestedSwitchCaseImplementor) VisitEnumNode(enumNode *nscgenerator.EnumNode) {
	commaListEnumerators := ""
	first := true
	for _, enumerator := range enumNode.Enumerators {
		if first {
			first = false
		} else {
			commaListEnumerators += ","
		}
		commaListEnumerators += enumerator
	}
	javaImplementor.Output += fmt.Sprintf(
		"private enum %s {%s}\n",
		enumNode.Name,
		commaListEnumerators,
	)
}

func (javaImplementor *JavaNestedSwitchCaseImplementor) VisitStatePropertyNode(
	statePropertyNode *nscgenerator.StatePropertyNode,
) {
	javaImplementor.Output += fmt.Sprintf("private State state = State.%s;\n", statePropertyNode.InitialState)
	javaImplementor.Output += "private void setState(State s) { state = s; }\n"
}

func (javaImplementor *JavaNestedSwitchCaseImplementor) VisitEventDelegatorsNode(
	eventDelegatorsNode *nscgenerator.EventDelegatorsNode,
) {
	for _, event := range eventDelegatorsNode.Events {
		javaImplementor.Output += fmt.Sprintf("public void %s() {handleEvent(Event.%s);}\n", event, event)
	}
}

func (javaImplementor *JavaNestedSwitchCaseImplementor) VisitFSMClassNode(fsmClassNode *nscgenerator.FSMClassNode) {
	if javaImplementor.javaPackage != "" {
		javaImplementor.Output += "package " + javaImplementor.javaPackage + ";\n"
	}

	actionsName := fsmClassNode.ActionsName
	if actionsName != "" {
		javaImplementor.Output += fmt.Sprintf("public abstract class %s {\n", fsmClassNode.ClassName)
	} else {
		javaImplementor.Output += fmt.Sprintf(
			"public abstract class %s implements %s {\n",
			fsmClassNode.ClassName, actionsName,
		)
	}

	javaImplementor.Output += "public abstract void unhandledTransition(String state, String event);\n"
	fsmClassNode.StateEnum.Accept(javaImplementor)
	fsmClassNode.EventEnum.Accept(javaImplementor)
	fsmClassNode.StateProperty.Accept(javaImplementor)
	fsmClassNode.Delegators.Accept(javaImplementor)
	fsmClassNode.HandleEvent.Accept(javaImplementor)
	if actionsName == "" {
		for _, action := range fsmClassNode.Actions {
			javaImplementor.Output += fmt.Sprintf("protected abstract void %s();\n", action)
		}
	}
	javaImplementor.Output += "}\n"
}

func (javaImplementor *JavaNestedSwitchCaseImplementor) VisitHandleEventNode(
	handleEventNode *nscgenerator.HandleEventNode,
) {
	javaImplementor.Output += "private void handleEvent(Event event) {\n"
	handleEventNode.SwitchCase.Accept(javaImplementor)
	javaImplementor.Output += "}\n"
}

func (javaImplementor *JavaNestedSwitchCaseImplementor) VisitEnumeratorNode(
	enumeratorNode *nscgenerator.EnumeratorNode,
) {
	javaImplementor.Output += fmt.Sprintf("%s.%s", enumeratorNode.Enumeration, enumeratorNode.Enumerator)
}

func (javaImplementor *JavaNestedSwitchCaseImplementor) VisitDefaultCaseNode(
	defaultCaseNode *nscgenerator.DefaultCaseNode,
) {
	javaImplementor.Output += "default: unhandledTransition(state.name(), event.name()); break;\n"
}
