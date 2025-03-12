package workflow

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ExecuteTaskFSM runs a task using FSM for state tracking.
func ExecuteTaskFSM(tsm *TaskStateMachine) {
	fmt.Printf("Task %s: Starting...\n", tsm.Task.Name)

	// Start the task FSM
	tsm.TransitionTask("start")

	var err error
	switch strings.ToLower(tsm.Task.Operator) {
	case "python":
		err = runPython(tsm.Task.Command)
	case "cli":
		err = runCommand(tsm.Task.Command)
	default:
		err = runCommand(tsm.Task.Command)
	}

	// Handle success or failure
	if err != nil {
		tsm.TransitionTask("fail")
		fmt.Printf("Task %s: Failed.\n", tsm.Task.Name)

		if tsm.Task.Catch != nil {
			fmt.Printf("Task %s: Running fallback: %s\n", tsm.Task.Name, tsm.Task.Catch.Command)
			runCommand(tsm.Task.Catch.Command)
			tsm.TransitionTask("succeed")
		}
	} else {
		tsm.TransitionTask("succeed")
	}
}

// ExecuteWorkflow reads and runs tasks in sequence.
func ExecuteWorkflow(wf *Workflow) {
	for _, task := range wf.Workflow.Tasks {
		tsm := NewTaskFSM(task)
		ExecuteTaskFSM(tsm)
		fmt.Println("-----")
	}
}

// runCommand executes shell commands.
func runCommand(cmdStr string) error {
	cmd := exec.Command("cmd", "/C", cmdStr) // Windows-compatible
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	fmt.Print(out.String())
	if err != nil {
		fmt.Printf("Error: %s\n", stderr.String())
	}
	return err
}

// runPython executes Python commands.
func runPython(cmdStr string) error {
	fullCmd := fmt.Sprintf("python -c \"%s\"", cmdStr)
	return runCommand(fullCmd)
}
