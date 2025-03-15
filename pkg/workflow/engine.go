package workflow

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
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

// Icons for states
const (
	iconWaiting    = "⏳" // waiting state
	iconRunning    = "🏃"
	iconSucceeded  = "✅"
	iconFailed     = "❌"
	iconRecovering = "♻️"
)

// getStateIcon returns an icon for a given state.
func getStateIcon(state string) string {
	switch state {
	case "waiting":
		return iconWaiting
	case "running":
		return iconRunning
	case "succeeded":
		return iconSucceeded
	case "failed":
		return iconFailed
	case "recovering":
		return iconRecovering
	default:
		return ""
	}
}

// loadingAnimation shows a simple spinner until stopChan is signaled.
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

// PrintStateTable prints a table listing all tasks, their current state (with icon), and runtime.
func PrintStateTable(tsms []*TaskStateMachine) {
	fmt.Println("\nCurrent Task States:")
	fmt.Println("| Tasks             | States                  | Runtime      |")
	fmt.Println("|-------------------|-------------------------|--------------|")
	for _, tsm := range tsms {
		runtimeStr := fmt.Sprintf("%.2f", tsm.Duration.Seconds())
		fmt.Printf("| %-17s | %-23s | %-12s |\n", tsm.Task.Name, tsm.State+" "+getStateIcon(tsm.State), runtimeStr)
	}
	fmt.Println()
}

// runCommand executes a shell command using OS-appropriate shell.
func runCommand(cmdStr string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("bash", "-c", cmdStr)
	}
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

// runPython executes a Python command using "python -c".
func runPython(cmdStr string) error {
    if strings.HasSuffix(cmdStr, ".py") {
        // Assume cmdStr is a Python script file.
        return runCommand(fmt.Sprintf("python %s", cmdStr))
    }
    // Otherwise, treat it as inline Python code.
    fullCmd := fmt.Sprintf("python -c \"%s\"", cmdStr)
    return runCommand(fullCmd)
}


// ExecuteTaskFSM runs a single task, updating its FSM and tracking runtime.
func ExecuteTaskFSM(tsm *TaskStateMachine) error {
	fmt.Printf("\n%s[WORKFLOW]%s Task %s: %sStarting...%s\n", colorCyan, colorReset, tsm.Task.Name, colorYellow, colorReset)
	tsm.TransitionTask("start")
	startTime := time.Now()
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

	stopAnimation <- true
	fmt.Print("\r")
	tsm.Duration = time.Since(startTime)

	if err != nil {
		tsm.TransitionTask("fail")
		fmt.Printf("%s[FAILED] Task %s failed!%s\n", colorRed, tsm.Task.Name, colorReset)
		if tsm.Task.Catch != nil {
			fmt.Printf("%s[RECOVERY]%s Task %s: Running fallback command: %s\n", colorYellow, colorReset, tsm.Task.Name, tsm.Task.Catch.Command)
			tsm.TransitionTask("recover")
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

// ExecuteWorkflow runs tasks sequentially and prints the state table after each task.
func ExecuteWorkflow(wf *Workflow) {
	fmt.Println("\n🚀 Starting Workflow Execution...\n")
	var tsms []*TaskStateMachine
	for _, task := range wf.Workflow.Tasks {
		tsms = append(tsms, NewTaskFSM(task))
	}
	for i, tsm := range tsms {
		if err := ExecuteTaskFSM(tsm); err != nil {
			PrintStateTable(tsms)
			fmt.Println("\n❌ Workflow execution halted due to task failure.\n")
			return
		}
		PrintStateTable(tsms)
		fmt.Printf("----- Completed task %d/%d -----\n\n", i+1, len(tsms))
	}
	fmt.Println("\n✅ Workflow Execution Completed Successfully! 🎉\n")
}
