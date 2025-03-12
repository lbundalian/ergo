package workflow

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

// Workflow structure
type Workflow struct {
	Workflow struct {
		Tasks []Task `yaml:"tasks"`
	} `yaml:"workflow"`
}

// Task represents a single workflow task.
type Task struct {
	Name     string `yaml:"name"`
	Operator string `yaml:"operator"`
	Command  string `yaml:"command"`
	Catch    *Catch `yaml:"catch,omitempty"`
}

// Catch defines the fallback command.
type Catch struct {
	Command string `yaml:"command"`
}

// LoadWorkflow loads a YAML workflow file.
func LoadWorkflow(filename string) (*Workflow, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, err
	}

	return &wf, nil
}
