package workflow

import (
	"fmt"
	"strings"

	"github.com/looplab/fsm"
)

// TaskStateMachine wraps a Task with a state machine.
type TaskStateMachine struct {
	Task  Task
	FSM   *fsm.FSM
	State string
}

// NewTaskFSM creates a new state machine for a Task.
func NewTaskFSM(task Task) *TaskStateMachine {
	tsm := &TaskStateMachine{
		Task: task,
	}
	tsm.FSM = fsm.NewFSM(
		"ready", // initial state
		fsm.Events{
			{Name: "start", Src: []string{"ready"}, Dst: "running"},
			{Name: "succeed", Src: []string{"running", "recovering"}, Dst: "succeeded"},
			{Name: "fail", Src: []string{"running"}, Dst: "failed"},
			{Name: "recover", Src: []string{"failed"}, Dst: "recovering"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				tsm.State = e.Dst
				fmt.Printf("Task %s transitioned to state: %s\n", task.Name, tsm.State)
			},
		},
	)
	return tsm
}
