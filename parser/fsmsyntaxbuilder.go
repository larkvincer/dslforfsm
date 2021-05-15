package parser

import (
	"github.com/larkvincer/dsl-fsm/parser/errortypes"
)

type FsmSyntaxBuilder struct {
	fsmSyntax     *FsmSyntax
	header        header
	parsedName    string
	transition    transition
	subTransition subTransition
}

func NewFsmSyntaxBuilder() *FsmSyntaxBuilder {
	return &FsmSyntaxBuilder{fsmSyntax: &FsmSyntax{}}
}

func (fsm *FsmSyntaxBuilder) newHeaderWithName() {
	fsm.header = header{Name: fsm.parsedName}
}

func (fsm *FsmSyntaxBuilder) addHeaderWithValue() {
	fsm.header.Value = fsm.parsedName
	fsm.fsmSyntax.Headers = append(fsm.fsmSyntax.Headers, fsm.header)
}

func (fsm *FsmSyntaxBuilder) setStateName() {
	fsm.transition = transition{State: stateSpec{Name: fsm.parsedName}}
	fsm.fsmSyntax.Logic = append(fsm.fsmSyntax.Logic, &fsm.transition)
}

func (fsm *FsmSyntaxBuilder) done() {
	fsm.fsmSyntax.Done = true
}

func (fsm *FsmSyntaxBuilder) setSuperStateName() {
	fsm.setStateName()
	fsm.transition.State.AbstractState = true
}

func (fsm *FsmSyntaxBuilder) setEvent() {
	fsm.subTransition = subTransition{Event: fsm.parsedName}
}

func (fsm *FsmSyntaxBuilder) setNullEvent() {
	fsm.subTransition = subTransition{Event: ""}
}

func (fsm *FsmSyntaxBuilder) setEntryAction() {
	fsm.transition.State.EntryActions = append(fsm.transition.State.EntryActions, fsm.parsedName)
}

func (fsm *FsmSyntaxBuilder) setExitAction() {
	fsm.transition.State.ExitActions = append(fsm.transition.State.ExitActions, fsm.parsedName)
}

func (fsm *FsmSyntaxBuilder) setStateBase() {
	fsm.transition.State.SuperStates = append(fsm.transition.State.SuperStates, fsm.parsedName)
}

func (fsm *FsmSyntaxBuilder) setNextState() {
	fsm.subTransition.NextState = fsm.parsedName
}

func (fsm *FsmSyntaxBuilder) setNullNextState() {
	fsm.subTransition.NextState = ""
}

func (fsm *FsmSyntaxBuilder) transitionWithAction() {
	fsm.subTransition.Actions = append(fsm.subTransition.Actions, fsm.parsedName)
	fsm.transition.SubTransitions = append(fsm.transition.SubTransitions, fsm.subTransition)
	// fmt.Printf("%v", fsm.fsmSyntax)
}

func (fsm *FsmSyntaxBuilder) transitionNullAction() {
	fsm.transition.SubTransitions = append(fsm.transition.SubTransitions, fsm.subTransition)
}

func (fsm *FsmSyntaxBuilder) addAction() {
	fsm.subTransition.Actions = append(fsm.subTransition.Actions, fsm.parsedName)
}

func (fsm *FsmSyntaxBuilder) transitionWithActions() {
	fsm.transition.SubTransitions = append(fsm.transition.SubTransitions, fsm.subTransition)
}

func (fsm *FsmSyntaxBuilder) headerError(state, event string, lineNumber, position int) {
	fsm.fsmSyntax.Errors = append(fsm.fsmSyntax.Errors, syntaxError{errortypes.HEADER, state + "|" + event, lineNumber, position})
}

func (fsm *FsmSyntaxBuilder) stateSpecError(state, event string, lineNumber, position int) {
	fsm.fsmSyntax.Errors = append(fsm.fsmSyntax.Errors, syntaxError{errortypes.STATE, state + "|" + event, lineNumber, position})
}

func (fsm *FsmSyntaxBuilder) transitionError(state, event string, lineNumber, position int) {
	fsm.fsmSyntax.Errors = append(fsm.fsmSyntax.Errors, syntaxError{errortypes.TRANSITION, state + "|" + event, lineNumber, position})
}

func (fsm *FsmSyntaxBuilder) transitionGroupError(state, event string, lineNumber, position int) {
	fsm.fsmSyntax.Errors = append(fsm.fsmSyntax.Errors, syntaxError{errortypes.TRANSITION_GROUP, state + "|" + event, lineNumber, position})
}

func (fsm *FsmSyntaxBuilder) endError(state, event string, lineNumber, position int) {
	fsm.fsmSyntax.Errors = append(fsm.fsmSyntax.Errors, syntaxError{errortypes.END, state + "|" + event, lineNumber, position})
}
func (fsm *FsmSyntaxBuilder) syntaxError(lineNumber, position int) {
	fsm.fsmSyntax.Errors = append(fsm.fsmSyntax.Errors, syntaxError{errortypes.SYNTAX, "", lineNumber, position})
}

func (fsm *FsmSyntaxBuilder) setName(name string) {
	fsm.parsedName = name
}

func (fsm *FsmSyntaxBuilder) String() string {
	return fsm.fsmSyntax.String()
}
