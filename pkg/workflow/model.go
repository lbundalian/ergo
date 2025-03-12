package workflow

import "gopkg.in/yaml.v2"

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
	Command   string    `yaml:"command"`   // main command
	Catch     *Catch    `yaml:"catch"`     // optional fallback command
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

// LoadWorkflow loads and parses a YAML workflow file.
func LoadWorkflow(filename string) (*Workflow, error) {
	data, err := yaml.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var wf Workflow
	if err := yaml.Unmarshal(data, &wf); err != nil {
		return nil, err
	}

	return &wf, nil
}
