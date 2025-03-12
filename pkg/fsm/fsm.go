package fsm

import (
	"errors"
	"fmt"
)

// FSM represents a finite state machine.
type FSM struct {
	CurrentState string
	Transitions  map[string]map[string]string
}

// NewFSM initializes a new FSM.
func NewFSM(initialState string, transitions map[string]map[string]string) *FSM {
	return &FSM{
		CurrentState: initialState,
		Transitions:  transitions,
	}
}

// CanTransition checks if a transition is possible.
func (f *FSM) CanTransition(event string) bool {
	_, exists := f.Transitions[f.CurrentState][event]
	return exists
}

// Transition performs a state change.
func (f *FSM) Transition(event string) error {
	if nextState, ok := f.Transitions[f.CurrentState][event]; ok {
		fmt.Printf("Transitioning from %s → %s using event: %s\n", f.CurrentState, nextState, event)
		f.CurrentState = nextState
		return nil
	}
	return errors.New("invalid transition")
}

// GetState returns the current state.
func (f *FSM) GetState() string {
	return f.CurrentState
}
