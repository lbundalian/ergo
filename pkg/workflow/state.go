package workflow

import (
	"fmt"
	"ergo/pkg/fsm"
	"time"
)

// TaskStateMachine wraps a Task with its current state, runtime, and an FSM.
type TaskStateMachine struct {
	Task     Task
	FSM      *fsm.FSM
	State    string        // Mirror of FSM's current state.
	Duration time.Duration // Runtime of the task.
}

// Define transitions for your FSM.
var transitions = map[string]map[string]string{
	"waiting": {"start": "running"},
	"running": {"succeed": "succeeded", "fail": "failed"},
	"failed":  {"recover": "recovering"},
	"recovering": {"succeed": "succeeded"},
}

// NewTaskFSM creates a new TaskStateMachine for a given task.
func NewTaskFSM(task Task) *TaskStateMachine {
	f := fsm.NewFSM("waiting", transitions)
	return &TaskStateMachine{
		Task:     task,
		FSM:      f,
		State:    f.GetState(),
		Duration: 0, // Initialize duration to zero.
	}
}

// TransitionTask updates the task state using your custom FSM.
func (tsm *TaskStateMachine) TransitionTask(event string) {
	if err := tsm.FSM.Transition(event); err != nil {
		fmt.Printf("Task %s: Invalid transition (%s): %v\n", tsm.Task.Name, event, err)
	} else {
		tsm.State = tsm.FSM.GetState()
		fmt.Printf("Task %s transitioned to state: %s %s\n", tsm.Task.Name, tsm.State, getStateIcon(tsm.State))
	}
}