package workflow

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// runCommand executes a bash command.
func runCommand(cmdStr string) error {
	cmd := exec.Command("bash", "-c", cmdStr)
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

// runPython executes a Python command using "python -c".
func runPython(cmdStr string) error {
	fullCmd := fmt.Sprintf("python -c \"%s\"", cmdStr)
	return runCommand(fullCmd)
}

// ExecuteTaskFSM executes a task using its state machine.
func ExecuteTaskFSM(tsm *TaskStateMachine) {
	task := tsm.Task
	operator := strings.ToLower(task.Operator)
	if operator == "" {
		operator = "bash"
	}

	// Start the state machine.
	if err := tsm.FSM.Event("start"); err != nil {
		fmt.Printf("Error starting task %s: %v\n", task.Name, err)
		return
	}

	if operator != "map" {
		var err error
		switch operator {
		case "python":
			fmt.Printf("Task %s: Running Python command: %s\n", task.Name, task.Command)
			err = runPython(task.Command)
		case "cli":
			fmt.Printf("Task %s: Running CLI command: %s\n", task.Name, task.Command)
			err = runCommand(task.Command)
		default:
			fmt.Printf("Task %s: Running Bash command: %s\n", task.Name, task.Command)
			err = runCommand(task.Command)
		}

		if err != nil {
			tsm.FSM.Event("fail")
			fmt.Printf("Task %s: Command failed.\n", task.Name)
			if task.Catch != nil && task.Catch.Command != "" {
				fmt.Printf("Task %s: Executing fallback command: %s\n", task.Name, task.Catch.Command)
				if err := tsm.FSM.Event("recover"); err != nil {
					fmt.Printf("Task %s: Failed to transition to recovering state: %v\n", task.Name, err)
				}
				var recErr error
				switch operator {
				case "python":
					recErr = runPython(task.Catch.Command)
				case "cli":
					recErr = runCommand(task.Catch.Command)
				default:
					recErr = runCommand(task.Catch.Command)
				}
				if recErr != nil {
					fmt.Printf("Task %s: Fallback command failed.\n", task.Name)
				} else {
					tsm.FSM.Event("succeed")
				}
			} else {
				fmt.Printf("Task %s: No fallback defined.\n", task.Name)
			}
		} else {
			tsm.FSM.Event("succeed")
		}
	} else {
		// Map operator.
		inputs := task.Input
		if len(inputs) == 0 {
			fmt.Printf("Task %s: No inputs for map operation.\n", task.Name)
			return
		}
		mapResults := []string{}
		for _, item := range inputs {
			cmd := strings.ReplaceAll(task.Command, "{item}", item)
			fmt.Printf("Task %s: Running map command for item '%s': %s\n", task.Name, item, cmd)
			err := runCommand(cmd)
			if err != nil {
				fmt.Printf("Task %s: Map command failed for item '%s'.\n", task.Name, item)
				if task.Catch != nil && task.Catch.Command != "" {
					fallback := strings.ReplaceAll(task.Catch.Command, "{item}", item)
					fmt.Printf("Task %s: Running fallback for item '%s': %s\n", task.Name, item, fallback)
					if err := runCommand(fallback); err != nil {
						fmt.Printf("Task %s: Fallback also failed for item '%s'.\n", task.Name, item)
					} else {
						mapResults = append(mapResults, item)
					}
				} else {
					fmt.Printf("Task %s: No fallback defined for item '%s'.\n", task.Name, item)
				}
			} else {
				mapResults = append(mapResults, item)
			}
		}
		if task.Reduce != nil && task.Reduce.Command != "" {
			resultsStr := strings.Join(mapResults, ",")
			reduceCmd := strings.ReplaceAll(task.Reduce.Command, "{results}", resultsStr)
			fmt.Printf("Task %s: Running reduce command: %s\n", task.Name, reduceCmd)
			if err := runCommand(reduceCmd); err != nil {
				fmt.Printf("Task %s: Reduce command failed.\n", task.Name)
			}
		}
		tsm.FSM.Event("succeed")
	}
}

// ExecuteWorkflow runs all tasks sequentially.
func ExecuteWorkflow(wf *Workflow) {
	for _, task := range wf.Workflow.Tasks {
		tsm := NewTaskFSM(task)
		ExecuteTaskFSM(tsm)
		fmt.Println("-----")
	}
}
