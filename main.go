package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/looplab/fsm"
	"gopkg.in/yaml.v2"
)

// ------------------------------
// Data Structures and YAML Model
// ------------------------------

// Workflow defines the top-level structure.
type Workflow struct {
	Workflow struct {
		Tasks []Task `yaml:"tasks"`
	} `yaml:"workflow"`
}

// Task represents a single task.
type Task struct {
	Name      string    `yaml:"name"`
	Operator  string    `yaml:"operator"`  // e.g., "bash", "cli", "python", "map"
	Input     []string  `yaml:"input"`     // used for map operator
	Output    []string  `yaml:"output"`    // not used in this example
	Command   string    `yaml:"command"`   // main command (try part)
	Catch     *Catch    `yaml:"catch"`     // optional fallback command (catch part)
	DependsOn string    `yaml:"depends_on"`// not implemented in this example
	Resources Resources `yaml:"resources"` // not used in execution here
	Container string    `yaml:"container"` // not used in execution here
	Reduce    *Reduce   `yaml:"reduce"`    // used for map operator reduce step
}

// Catch holds the fallback command.
type Catch struct {
	Command string `yaml:"command"`
}

// Reduce holds the reduce command.
type Reduce struct {
	Command string `yaml:"command"`
}

// Resources defines resource requirements.
type Resources struct {
	CPU int    `yaml:"cpu"`
	Mem string `yaml:"mem"`
}

// ------------------------------
// State Machine for Task Execution
// ------------------------------

// TaskStateMachine wraps a task with a state machine.
type TaskStateMachine struct {
	Task  Task
	FSM   *fsm.FSM
	State string
}

// newTaskFSM creates a new state machine for a task.
func newTaskFSM(task Task) *TaskStateMachine {
	t := &TaskStateMachine{
		Task: task,
	}
	t.FSM = fsm.NewFSM(
		"ready", // initial state
		fsm.Events{
			{Name: "start", Src: []string{"ready"}, Dst: "running"},
			{Name: "succeed", Src: []string{"running", "recovering"}, Dst: "succeeded"},
			{Name: "fail", Src: []string{"running"}, Dst: "failed"},
			{Name: "recover", Src: []string{"failed"}, Dst: "recovering"},
		},
		fsm.Callbacks{
			"enter_state": func(e *fsm.Event) {
				t.State = e.Dst
				fmt.Printf("Task %s transitioned to state: %s\n", task.Name, t.State)
			},
		},
	)
	return t
}

// ------------------------------
// Command Runners
// ------------------------------

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

func runPython(cmdStr string) error {
	fullCmd := fmt.Sprintf("python -c \"%s\"", cmdStr)
	return runCommand(fullCmd)
}

// ------------------------------
// Workflow Executor (State Machine Based)
// ------------------------------

func executeTaskFSM(tsm *TaskStateMachine) {
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
		// Map operator
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

func executeWorkflow(wf Workflow) {
	for _, task := range wf.Workflow.Tasks {
		tsm := newTaskFSM(task)
		executeTaskFSM(tsm)
		fmt.Println("-----")
	}
}

// ------------------------------
// Command-Line Interface
// ------------------------------

func main() {
	// Define a --run flag.
	runFlag := flag.String("run", "", "Run the workflow specified by the given file (ErgoFile)")
	flag.Parse()

	if *runFlag == "" {
		fmt.Println("Usage: ergo --run <ErgoFile>")
		os.Exit(1)
	}

	// Read and parse the workflow file.
	data, err := ioutil.ReadFile(*runFlag)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	var wf Workflow
	if err = yaml.Unmarshal(data, &wf); err != nil {
		fmt.Println("Error parsing YAML:", err)
		os.Exit(1)
	}

	fmt.Println("Parsed Workflow:")
	parsed, _ := yaml.Marshal(&wf)
	fmt.Println(string(parsed))

	// Execute the workflow.
	executeWorkflow(wf)
}
