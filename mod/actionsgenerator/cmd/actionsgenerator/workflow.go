package main

type PushTrigger struct {
	Branches []string `yaml:"branches,omitempty"`
	Tags     []string `yaml:"tags,omitempty"`
}

type Trigger struct {
	Push PushTrigger `yaml:"push,omitempty"`
}

type Args map[string]interface{}

type Step struct {
	Name string `yaml:"name,omitempty"`
	If   string `yaml:"if,omitempty"`
	Uses string `yaml:"uses,omitempty"`
	ID   string `yaml:"id,omitempty"`
	Run  string `yaml:"run,omitempty"`
	With Args   `yaml:"with,omitempty"`
}

type Job struct {
	RunsOn string `yaml:"runs-on"`
	Steps  []Step `yaml:"steps"`
}

type Workflow struct {
	Name string  `yaml:"name"`
	On   Trigger `yaml:"on,omitempty"`
	Jobs map[string]Job
}
