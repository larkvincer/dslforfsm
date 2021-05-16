package parser

import (
	"github.com/larkvincer/dsl-fsm/parser/events"
	"github.com/larkvincer/dsl-fsm/tokens"
)

type Parser struct {
	state         string
	syntaxBuilder *SyntaxBuilder
	transitions   []Transition
}

func NewParser(syntaxBuilder SyntaxBuilder) *Parser {
	return &Parser{state: events.HEADER, syntaxBuilder: &syntaxBuilder, transitions: buildTransitions()}
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
	case events.HEADER, events.HEADER_COLON, events.HEADER_VALUE:
		(*parser.syntaxBuilder).headerError(parser.state, event, lineNumber, position)

	case events.STATE_SPEC,
		events.SUPER_STATE_NAME,
		events.SUPER_STATE_CLOSE,
		events.STATE_MODIFIER,
		events.EXIT_ACTION,
		events.ENTRY_ACTION,
		events.STATE_BASE:
		(*parser.syntaxBuilder).stateSpecError(parser.state, event, lineNumber, position)

	case events.SINGLE_EVENT,
		events.SINGLE_NEXT_STATE,
		events.SINGLE_ACTION_GROUP,
		events.SINGLE_ACTION_GROUP_NAME:
		(*parser.syntaxBuilder).transitionError(parser.state, event, lineNumber, position)

	case events.SUBTRANSITION_GROUP,
		events.GROUP_EVENT,
		events.GROUP_NEXT_STATE,
		events.GROUP_ACTION_GROUP,
		events.GROUP_ACTION_GROUP_NAME:
		(*parser.syntaxBuilder).transitionGroupError(parser.state, event, lineNumber, position)

	case events.END:
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
		{events.HEADER, tokens.NAME, events.HEADER_COLON, func(sb *SyntaxBuilder) { (*sb).newHeaderWithName() }},
		{events.HEADER, tokens.OPEN_BRACE, events.STATE_SPEC, nil},
		{events.HEADER_COLON, tokens.COLON, events.HEADER_VALUE, nil},
		{events.HEADER_VALUE, tokens.NAME, events.HEADER, func(sb *SyntaxBuilder) { (*sb).addHeaderWithValue() }},

		{events.STATE_SPEC, tokens.OPEN_PAREN, events.SUPER_STATE_NAME, nil},
		{events.STATE_SPEC, tokens.NAME, events.STATE_MODIFIER, func(sb *SyntaxBuilder) { (*sb).setStateName() }},
		{events.STATE_SPEC, tokens.CLOSE_BRACE, events.END, func(sb *SyntaxBuilder) { (*sb).done() }},
		{events.SUPER_STATE_NAME, tokens.NAME, events.SUPER_STATE_CLOSE, func(sb *SyntaxBuilder) { (*sb).setSuperStateName() }},
		{events.SUPER_STATE_CLOSE, tokens.CLOSE_PAREN, events.STATE_MODIFIER, nil},

		{events.STATE_MODIFIER, tokens.ENTRY_STATE, events.ENTRY_ACTION, nil},
		{events.STATE_MODIFIER, tokens.EXIT_STATE, events.EXIT_ACTION, nil},
		{events.STATE_MODIFIER, tokens.COLON, events.STATE_BASE, nil},
		{events.STATE_MODIFIER, tokens.NAME, events.SINGLE_EVENT, func(sb *SyntaxBuilder) { (*sb).setEvent() }},
		{events.STATE_MODIFIER, tokens.STAR, events.SINGLE_EVENT, func(sb *SyntaxBuilder) { (*sb).setNullEvent() }},
		{events.STATE_MODIFIER, tokens.OPEN_BRACE, events.SUBTRANSITION_GROUP, nil},

		{events.ENTRY_ACTION, tokens.NAME, events.STATE_MODIFIER, func(sb *SyntaxBuilder) { (*sb).setEntryAction() }},
		{events.ENTRY_ACTION, tokens.OPEN_BRACE, events.MULTIPLE_ENTRY_ACTIONS, nil},
		{events.MULTIPLE_ENTRY_ACTIONS, tokens.NAME, events.MULTIPLE_ENTRY_ACTIONS, func(sb *SyntaxBuilder) { (*sb).setEntryAction() }},
		{events.MULTIPLE_ENTRY_ACTIONS, tokens.CLOSE_BRACE, events.STATE_MODIFIER, nil},

		{events.EXIT_ACTION, tokens.NAME, events.STATE_MODIFIER, func(sb *SyntaxBuilder) { (*sb).setExitAction() }},
		{events.EXIT_ACTION, tokens.OPEN_BRACE, events.MULTIPLE_EXIT_ACTIONS, nil},
		{events.MULTIPLE_EXIT_ACTIONS, tokens.NAME, events.MULTIPLE_EXIT_ACTIONS, func(sb *SyntaxBuilder) { (*sb).setExitAction() }},
		{events.MULTIPLE_EXIT_ACTIONS, tokens.CLOSE_BRACE, events.STATE_MODIFIER, nil},

		{events.STATE_BASE, tokens.NAME, events.STATE_MODIFIER, func(sb *SyntaxBuilder) { (*sb).setStateBase() }},

		{events.SINGLE_EVENT, tokens.NAME, events.SINGLE_NEXT_STATE, func(sb *SyntaxBuilder) { (*sb).setNextState() }},
		{events.SINGLE_EVENT, tokens.STAR, events.SINGLE_NEXT_STATE, func(sb *SyntaxBuilder) { (*sb).setNullNextState() }},
		{events.SINGLE_NEXT_STATE, tokens.NAME, events.STATE_SPEC, func(sb *SyntaxBuilder) { (*sb).transitionWithAction() }},
		{events.SINGLE_NEXT_STATE, tokens.STAR, events.STATE_SPEC, func(sb *SyntaxBuilder) { (*sb).transitionNullAction() }},
		{events.SINGLE_NEXT_STATE, tokens.OPEN_BRACE, events.SINGLE_ACTION_GROUP, nil},
		{events.SINGLE_ACTION_GROUP, tokens.NAME, events.SINGLE_ACTION_GROUP_NAME, func(sb *SyntaxBuilder) { (*sb).addAction() }},
		{events.SINGLE_ACTION_GROUP, tokens.CLOSE_BRACE, events.STATE_SPEC, func(sb *SyntaxBuilder) { (*sb).transitionNullAction() }},
		{events.SINGLE_ACTION_GROUP_NAME, tokens.NAME, events.SINGLE_ACTION_GROUP_NAME, func(sb *SyntaxBuilder) { (*sb).addAction() }},
		{events.SINGLE_ACTION_GROUP_NAME, tokens.CLOSE_BRACE, events.STATE_SPEC, func(sb *SyntaxBuilder) { (*sb).transitionWithActions() }},

		{events.SUBTRANSITION_GROUP, tokens.CLOSE_BRACE, events.STATE_SPEC, nil},
		{events.SUBTRANSITION_GROUP, tokens.NAME, events.GROUP_EVENT, func(sb *SyntaxBuilder) { (*sb).setEvent() }},
		{events.SUBTRANSITION_GROUP, tokens.STAR, events.GROUP_EVENT, func(sb *SyntaxBuilder) { (*sb).setNullEvent() }},

		{events.GROUP_EVENT, tokens.NAME, events.GROUP_NEXT_STATE, func(sb *SyntaxBuilder) { (*sb).setNextState() }},
		{events.GROUP_EVENT, tokens.STAR, events.GROUP_NEXT_STATE, func(sb *SyntaxBuilder) { (*sb).setNullNextState() }},
		{events.GROUP_NEXT_STATE, tokens.NAME, events.SUBTRANSITION_GROUP, func(sb *SyntaxBuilder) { (*sb).transitionWithAction() }},
		{events.GROUP_NEXT_STATE, tokens.STAR, events.SUBTRANSITION_GROUP, func(sb *SyntaxBuilder) { (*sb).transitionNullAction() }},
		{events.GROUP_NEXT_STATE, tokens.OPEN_BRACE, events.GROUP_ACTION_GROUP, nil},
		{events.GROUP_ACTION_GROUP, tokens.NAME, events.GROUP_ACTION_GROUP_NAME, func(sb *SyntaxBuilder) { (*sb).addAction() }},
		{events.GROUP_ACTION_GROUP, tokens.CLOSE_BRACE, events.SUBTRANSITION_GROUP, func(sb *SyntaxBuilder) { (*sb).transitionNullAction() }},
		{events.GROUP_ACTION_GROUP_NAME, tokens.NAME, events.GROUP_ACTION_GROUP_NAME, func(sb *SyntaxBuilder) { (*sb).addAction() }},
		{events.GROUP_ACTION_GROUP_NAME, tokens.CLOSE_BRACE, events.SUBTRANSITION_GROUP, func(sb *SyntaxBuilder) { (*sb).transitionWithActions() }},

		{events.END, tokens.EOF, events.END, nil},
	}
}
