package parser

import (
	"github.com/larkvincer/dsl-fsm/parser/states"
	"github.com/larkvincer/dsl-fsm/tokens"
)

type Parser struct {
	state         string
	syntaxBuilder *SyntaxBuilder
	transitions   []Transition
}

func NewParser(syntaxBuilder SyntaxBuilder) *Parser {
	return &Parser{state: states.HEADER, syntaxBuilder: &syntaxBuilder, transitions: buildTransitions()}
}

func (parser *Parser) OpenBrace(lineNumber, position int) {
	parser.HandleEvent(tokens.OPEN_BRACE, lineNumber, position)
}
func (parser *Parser) CloseBrace(lineNumber, position int) {
	parser.HandleEvent(tokens.CLOSE_BRACE, lineNumber, position)
}
func (parser *Parser) OpenParen(lineNumber, position int) {
	parser.HandleEvent(tokens.OPEN_PAREN, lineNumber, position)
}
func (parser *Parser) CloseParen(lineNumber, position int) {
	parser.HandleEvent(tokens.CLOSE_PAREN, lineNumber, position)
}
func (parser *Parser) OpenAngle(lineNumber, position int) {
	parser.HandleEvent(tokens.ENTRY_STATE, lineNumber, position)
}
func (parser *Parser) CloseAngle(lineNumber, position int) {
	parser.HandleEvent(tokens.EXIT_STATE, lineNumber, position)
}
func (parser *Parser) Star(lineNumber, position int) {
	parser.HandleEvent(tokens.STAR, lineNumber, position)
}
func (parser *Parser) Colon(lineNumber, position int) {
	parser.HandleEvent(tokens.COLON, lineNumber, position)
}
func (parser *Parser) Name(name string, lineNumber, position int) {
	(*parser.syntaxBuilder).setName(name)
	parser.HandleEvent(tokens.NAME, lineNumber, position)
}
func (parser *Parser) Error(lineNumber, position int) {
	(*parser.syntaxBuilder).syntaxError(lineNumber, position)
}

func (parser *Parser) HandleEvent(event string, line, position int) {
	for _, transition := range parser.transitions {
		if transition.currentState == parser.state && transition.event == event {
			parser.state = transition.newState
			if transition.action != nil {
				transition.action(parser.syntaxBuilder)
			}
			return
		}
	}
	parser.handleEventError(event, line, position)
}

func (parser *Parser) handleEventError(event string, lineNumber, position int) {
	switch parser.state {
	case states.HEADER, states.HEADER_COLON, states.HEADER_VALUE:
		(*parser.syntaxBuilder).headerError(parser.state, event, lineNumber, position)

	case states.STATE_SPEC,
		states.SUPER_STATE_NAME,
		states.SUPER_STATE_CLOSE,
		states.STATE_MODIFIER,
		states.EXIT_ACTION,
		states.ENTRY_ACTION,
		states.STATE_BASE:
		(*parser.syntaxBuilder).stateSpecError(parser.state, event, lineNumber, position)

	case states.SINGLE_EVENT,
		states.SINGLE_NEXT_STATE,
		states.SINGLE_ACTION_GROUP,
		states.SINGLE_ACTION_GROUP_NAME:
		(*parser.syntaxBuilder).transitionError(parser.state, event, lineNumber, position)

	case states.SUBTRANSITION_GROUP,
		states.GROUP_EVENT,
		states.GROUP_NEXT_STATE,
		states.GROUP_ACTION_GROUP,
		states.GROUP_ACTION_GROUP_NAME:
		(*parser.syntaxBuilder).transitionGroupError(parser.state, event, lineNumber, position)

	case states.END:
		(*parser.syntaxBuilder).endError(parser.state, event, lineNumber, position)
	}
}

type Transition struct {
	currentState string
	event        string
	newState     string
	action       func(builder *SyntaxBuilder)
}

func buildTransitions() []Transition {
	return []Transition{
		{states.HEADER, tokens.NAME, states.HEADER_COLON, func(sb *SyntaxBuilder) { (*sb).newHeaderWithName() }},
		{states.HEADER, tokens.OPEN_BRACE, states.STATE_SPEC, nil},
		{states.HEADER_COLON, tokens.COLON, states.HEADER_VALUE, nil},
		{states.HEADER_VALUE, tokens.NAME, states.HEADER, func(sb *SyntaxBuilder) { (*sb).addHeaderWithValue() }},

		{states.STATE_SPEC, tokens.OPEN_PAREN, states.SUPER_STATE_NAME, nil},
		{states.STATE_SPEC, tokens.NAME, states.STATE_MODIFIER, func(sb *SyntaxBuilder) { (*sb).setStateName() }},
		{states.STATE_SPEC, tokens.CLOSE_BRACE, states.END, func(sb *SyntaxBuilder) { (*sb).done() }},
		{states.SUPER_STATE_NAME, tokens.NAME, states.SUPER_STATE_CLOSE, func(sb *SyntaxBuilder) { (*sb).setSuperStateName() }},
		{states.SUPER_STATE_CLOSE, tokens.CLOSE_PAREN, states.STATE_MODIFIER, nil},

		{states.STATE_MODIFIER, tokens.ENTRY_STATE, states.ENTRY_ACTION, nil},
		{states.STATE_MODIFIER, tokens.EXIT_STATE, states.EXIT_ACTION, nil},
		{states.STATE_MODIFIER, tokens.COLON, states.STATE_BASE, nil},
		{states.STATE_MODIFIER, tokens.NAME, states.SINGLE_EVENT, func(sb *SyntaxBuilder) { (*sb).setEvent() }},
		{states.STATE_MODIFIER, tokens.STAR, states.SINGLE_EVENT, func(sb *SyntaxBuilder) { (*sb).setNullEvent() }},
		{states.STATE_MODIFIER, tokens.OPEN_BRACE, states.SUBTRANSITION_GROUP, nil},

		{states.ENTRY_ACTION, tokens.NAME, states.STATE_MODIFIER, func(sb *SyntaxBuilder) { (*sb).setEntryAction() }},
		{states.ENTRY_ACTION, tokens.OPEN_BRACE, states.MULTIPLE_ENTRY_ACTIONS, nil},
		{states.MULTIPLE_ENTRY_ACTIONS, tokens.NAME, states.MULTIPLE_ENTRY_ACTIONS, func(sb *SyntaxBuilder) { (*sb).setEntryAction() }},
		{states.MULTIPLE_ENTRY_ACTIONS, tokens.CLOSE_BRACE, states.STATE_MODIFIER, nil},

		{states.EXIT_ACTION, tokens.NAME, states.STATE_MODIFIER, func(sb *SyntaxBuilder) { (*sb).setExitAction() }},
		{states.EXIT_ACTION, tokens.OPEN_BRACE, states.MULTIPLE_EXIT_ACTIONS, nil},
		{states.MULTIPLE_EXIT_ACTIONS, tokens.NAME, states.MULTIPLE_EXIT_ACTIONS, func(sb *SyntaxBuilder) { (*sb).setExitAction() }},
		{states.MULTIPLE_EXIT_ACTIONS, tokens.CLOSE_BRACE, states.STATE_MODIFIER, nil},

		{states.STATE_BASE, tokens.NAME, states.STATE_MODIFIER, func(sb *SyntaxBuilder) { (*sb).setStateBase() }},

		{states.SINGLE_EVENT, tokens.NAME, states.SINGLE_NEXT_STATE, func(sb *SyntaxBuilder) { (*sb).setNextState() }},
		{states.SINGLE_EVENT, tokens.STAR, states.SINGLE_NEXT_STATE, func(sb *SyntaxBuilder) { (*sb).setNullNextState() }},
		{states.SINGLE_NEXT_STATE, tokens.NAME, states.STATE_SPEC, func(sb *SyntaxBuilder) { (*sb).transitionWithAction() }},
		{states.SINGLE_NEXT_STATE, tokens.STAR, states.STATE_SPEC, func(sb *SyntaxBuilder) { (*sb).transitionNullAction() }},
		{states.SINGLE_NEXT_STATE, tokens.OPEN_BRACE, states.SINGLE_ACTION_GROUP, nil},
		{states.SINGLE_ACTION_GROUP, tokens.NAME, states.SINGLE_ACTION_GROUP_NAME, func(sb *SyntaxBuilder) { (*sb).addAction() }},
		{states.SINGLE_ACTION_GROUP, tokens.CLOSE_BRACE, states.STATE_SPEC, func(sb *SyntaxBuilder) { (*sb).transitionNullAction() }},
		{states.SINGLE_ACTION_GROUP_NAME, tokens.NAME, states.SINGLE_ACTION_GROUP_NAME, func(sb *SyntaxBuilder) { (*sb).addAction() }},
		{states.SINGLE_ACTION_GROUP_NAME, tokens.CLOSE_BRACE, states.STATE_SPEC, func(sb *SyntaxBuilder) { (*sb).transitionWithActions() }},

		{states.SUBTRANSITION_GROUP, tokens.CLOSE_BRACE, states.STATE_SPEC, nil},
		{states.SUBTRANSITION_GROUP, tokens.NAME, states.GROUP_EVENT, func(sb *SyntaxBuilder) { (*sb).setEvent() }},
		{states.SUBTRANSITION_GROUP, tokens.STAR, states.GROUP_EVENT, func(sb *SyntaxBuilder) { (*sb).setNullEvent() }},

		{states.GROUP_EVENT, tokens.NAME, states.GROUP_NEXT_STATE, func(sb *SyntaxBuilder) { (*sb).setNextState() }},
		{states.GROUP_EVENT, tokens.STAR, states.GROUP_NEXT_STATE, func(sb *SyntaxBuilder) { (*sb).setNullNextState() }},
		{states.GROUP_NEXT_STATE, tokens.NAME, states.SUBTRANSITION_GROUP, func(sb *SyntaxBuilder) { (*sb).transitionWithAction() }},
		{states.GROUP_NEXT_STATE, tokens.STAR, states.SUBTRANSITION_GROUP, func(sb *SyntaxBuilder) { (*sb).transitionNullAction() }},
		{states.GROUP_NEXT_STATE, tokens.OPEN_BRACE, states.GROUP_ACTION_GROUP, nil},
		{states.GROUP_ACTION_GROUP, tokens.NAME, states.GROUP_ACTION_GROUP_NAME, func(sb *SyntaxBuilder) { (*sb).addAction() }},
		{states.GROUP_ACTION_GROUP, tokens.CLOSE_BRACE, states.SUBTRANSITION_GROUP, func(sb *SyntaxBuilder) { (*sb).transitionNullAction() }},
		{states.GROUP_ACTION_GROUP_NAME, tokens.NAME, states.GROUP_ACTION_GROUP_NAME, func(sb *SyntaxBuilder) { (*sb).addAction() }},
		{states.GROUP_ACTION_GROUP_NAME, tokens.CLOSE_BRACE, states.SUBTRANSITION_GROUP, func(sb *SyntaxBuilder) { (*sb).transitionWithActions() }},

		{states.END, tokens.EOF, states.END, nil},
	}
}
