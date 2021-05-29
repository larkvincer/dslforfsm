package optimizer

import "fmt"

type OptimizedStateMachine struct {
	States      []string
	Events      []string
	Actions     []string
	Header      Header
	Transitions []Transition
}

func (osm *OptimizedStateMachine) String() string {
	transitionsString := osm.transitionsToString()
	transitionsString = transitionsString[0 : len(transitionsString)-2]
	return fmt.Sprintf(
		"Initial %s\nFsm: %s\nActions:%s\n{\n  %s\n",
		osm.Header.Initial, osm.Header.Fsm, osm.Header.Actions, transitionsString,
	)
}

func (osm *OptimizedStateMachine) transitionsToString() string {
	result := ""
	for _, transition := range osm.Transitions {
		result += transition.String()
	}
	return result
}

type Header struct {
	Initial string
	Fsm     string
	Actions string
}

type Transition struct {
	CurrentState   string
	SubTransitions []SubTransition
}

func (t *Transition) String() string {
	result := fmt.Sprintf("%s {\n", t.CurrentState)
	for _, subTransition := range t.SubTransitions {
		result += subTransition.String()
	}
	result += "}\n"
	return result
}

type SubTransition struct {
	Event     string
	NextState string
	Actions   []string
}

func (st *SubTransition) String() string {
	return fmt.Sprintf("  %s %s {%s}\n", st.Event, st.NextState, st.actionsToString())
}

func (st *SubTransition) actionsToString() string {
	result := ""
	if len(st.Actions) == 0 {
		return result
	}

	for _, action := range st.Actions {
		result += action + " "
	}
	return result[0 : len(result)-1]
}
