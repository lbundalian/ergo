package workflow

import (
	"fmt"
	"ergo/pkg/fsm"
)

// TaskStateMachine wraps a Task with a state machine.
type TaskStateMachine struct {
	Task  Task
	FSM   *fsm.FSM
	State string
}

// NewTaskFSM initializes a state machine for a task.
func NewTaskFSM(task Task) *TaskStateMachine {
	tsm := &TaskStateMachine{
		Task: task,
	}

	tsm.FSM = fsm.NewFSM(
		"ready", // initial state
		map[string]map[string]string{
			"ready":     {"start": "running"},
			"running":   {"succeed": "succeeded", "fail": "failed"},
			"failed":    {"recover": "recovering"},
			"recovering": {"succeed": "succeeded", "fail": "failed"},
		},
	)

	return tsm
}

// TransitionTask changes the task state
func (tsm *TaskStateMachine) TransitionTask(event string) {
	err := tsm.FSM.Transition(event)
	if err != nil {
		fmt.Printf("Task %s: Invalid transition (%s)\n", tsm.Task.Name, event)
	}
}
