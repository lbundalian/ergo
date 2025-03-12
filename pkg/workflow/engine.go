package workflow

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Task colors
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// Custom simple loading animation
func loadingAnimation(taskName string, stopChan chan bool) {
	frames := []string{"-", "\\", "|", "/"}
	i := 0
	for {
		select {
		case <-stopChan:
			return
		default:
			fmt.Printf("\r%s[WORKFLOW]%s Task %s: %s Running %s ", colorCyan, colorReset, taskName, colorYellow, frames[i%len(frames)])
			i++
			time.Sleep(150 * time.Millisecond)
		}
	}
}

// ExecuteTaskFSM runs a task and ensures workflow stops on failure.
func ExecuteTaskFSM(tsm *TaskStateMachine) error {
	fmt.Printf("\n%s[WORKFLOW]%s Task %s: %sStarting...%s\n", colorCyan, colorReset, tsm.Task.Name, colorYellow, colorReset)

	// Start the task FSM
	tsm.TransitionTask("start")

	// Start custom loading animation
	stopAnimation := make(chan bool)
	go loadingAnimation(tsm.Task.Name, stopAnimation)

	var err error
	switch strings.ToLower(tsm.Task.Operator) {
	case "python":
		err = runPython(tsm.Task.Command)
	case "cli":
		err = runCommand(tsm.Task.Command)
	default:
		err = runCommand(tsm.Task.Command)
	}

	// Stop the loading animation
	stopAnimation <- true
	fmt.Print("\r") // Clear the line

	// Handle success or failure
	if err != nil {
		tsm.TransitionTask("fail")
		fmt.Printf("%s[FAILED] Task %s failed!%s\n", colorRed, tsm.Task.Name, colorReset)

		// If a fallback exists, execute it
		if tsm.Task.Catch != nil {
			fmt.Printf("%s[RECOVERY]%s Task %s: Running fallback command: %s\n", colorYellow, colorReset, tsm.Task.Name, tsm.Task.Catch.Command)
			recErr := runCommand(tsm.Task.Catch.Command)
			if recErr != nil {
				fmt.Printf("%s[FALLBACK FAILED] Task %s: Fallback command also failed!%s\n", colorRed, tsm.Task.Name, colorReset)
				return fmt.Errorf("task %s failed, stopping workflow", tsm.Task.Name)
			} else {
				tsm.TransitionTask("succeed")
				fmt.Printf("%s[SUCCESS]%s Task %s recovered successfully!%s\n", colorGreen, colorReset, tsm.Task.Name, colorReset)
			}
		} else {
			fmt.Printf("%s[ERROR] No fallback found. Stopping workflow execution.%s\n", colorRed, colorReset)
			return fmt.Errorf("task %s failed, stopping workflow", tsm.Task.Name)
		}
	} else {
		tsm.TransitionTask("succeed")
		fmt.Printf("%s[SUCCESS]%s Task %s completed successfully!%s\n", colorGreen, colorReset, tsm.Task.Name, colorReset)
	}

	return nil
}

// ExecuteWorkflow runs tasks sequentially and stops on failure.
func ExecuteWorkflow(wf *Workflow) {
	fmt.Println("\n🚀 Starting Workflow Execution...\n")

	for _, task := range wf.Workflow.Tasks {
		tsm := NewTaskFSM(task)

		// Execute task and STOP if it fails
		if err := ExecuteTaskFSM(tsm); err != nil {
			fmt.Println("\n❌ Workflow execution halted due to task failure.\n")
			return
		}

		fmt.Println("-----") // Separator between tasks
	}

	fmt.Println("\n✅ Workflow Execution Completed Successfully! 🎉\n")
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
		fmt.Printf("%sError: %s%s\n", colorRed, stderr.String(), colorReset)
	}
	return err
}

// runPython executes Python commands.
func runPython(cmdStr string) error {
	fullCmd := fmt.Sprintf("python -c \"%s\"", cmdStr)
	return runCommand(fullCmd)
}
