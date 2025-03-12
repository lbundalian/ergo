package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/lbundalian/ergo/pkg/workflow"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func main() {
	// Define a --run flag to specify the ErgoFile.
	runFile := flag.String("run", "", "Run the workflow specified by the given ErgoFile (YAML format)")
	flag.Parse()

	if *runFile == "" {
		fmt.Println("Usage: ergo --run <ErgoFile>")
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(*runFile)
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	var wf workflow.Workflow
	err = yaml.Unmarshal(data, &wf)
	if err != nil {
		fmt.Println("Error parsing YAML:", err)
		os.Exit(1)
	}

	fmt.Println("Parsed Workflow:")
	parsed, _ := yaml.Marshal(&wf)
	fmt.Println(string(parsed))

	workflow.ExecuteWorkflow(&wf)
}
