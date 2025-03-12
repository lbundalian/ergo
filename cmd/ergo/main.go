package main

import (
	"flag"
	"fmt"
	"os"
	"ergo/pkg/workflow"
)

func main() {
	runFlag := flag.String("run", "", "Run the workflow file")
	flag.Parse()

	if *runFlag == "" {
		fmt.Println("Usage: ergo --run <workflow.yaml>")
		os.Exit(1)
	}

	wf, err := workflow.LoadWorkflow(*runFlag)
	if err != nil {
		fmt.Println("Error loading workflow:", err)
		os.Exit(1)
	}

	fmt.Println("Executing workflow...")
	workflow.ExecuteWorkflow(wf)
}
