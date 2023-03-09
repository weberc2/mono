package main

// PushTrigger models the GitHub Actions `workflow.on.trigger` field.
type PushTrigger struct {
	// Branches holds the branches on which the trigger will fire.
	Branches []string `yaml:"branches,omitempty"`

	// Tags holds the tags on which the trigger will fire.
	Tags []string `yaml:"tags,omitempty"`
}

// Trigger models the GitHub Actions `workflow.on` field.
type Trigger struct {
	Push PushTrigger `yaml:"push,omitempty"`
}

// Args models arbitrary key value pairs used for `workflow.jobs.with` field.
type Args map[string]interface{}

// Step models a step in a GitHub Actions `Job` resource.
type Step struct {
	// Name is the optional name of the step. Optional.
	Name string `yaml:"name,omitempty"`

	// If holds a condition--the step runs if the condition is satisfied, else
	// it skips. Optional.
	If string `yaml:"if,omitempty"`

	// Uses holds the name of a GitHub Action to invoke for this step.
	// Optional.
	Uses string `yaml:"uses,omitempty"`

	// ID uniquely identifies the ID in the job. Optional.
	ID string `yaml:"id,omitempty"`

	// Run contains a script to run for the step. Optional.
	Run string `yaml:"run,omitempty"`

	// With contains the parameters to pass into a designated GitHub Action
	// (see `Uses` above). Optional.
	With Args `yaml:"with,omitempty"`
}

// Job models a GitHub Actions `Job` resource for the `workflow.jobs` field.
type Job struct {
	// RunsOn contains the identifier for a container image that will be used
	// to run the job.
	RunsOn string `yaml:"runs-on"`

	// Steps holds the list of steps to run.
	Steps []Step `yaml:"steps"`
}

// Workflow models a GitHub Actions workflow.
type Workflow struct {
	// Name is the name of the workflow.
	Name string `yaml:"name"`

	// On is the trigger condition for the workflow. Optional.
	On Trigger `yaml:"on,omitempty"`

	// Jobs is the list of Jobs to execute which constitute the workflow.
	Jobs map[string]Job
}

func NewWorkflow(name string) *Workflow {
	return &Workflow{
		Name: name,
		On: Trigger{Push: PushTrigger{
			Branches: []string{"*"},
			Tags:     []string{"*"},
		}},
		Jobs: map[string]Job{},
	}
}

func (workflow *Workflow) WithJob(name string, job Job) *Workflow {
	workflow.Jobs[name] = job
	return workflow
}
