package orchestrator

// ExecutionPlan represents a plan to execute a user request.
type ExecutionPlan struct {
	DirectResponse bool       `json:"direct_response"`
	Steps          []PlanStep `json:"steps"`
	Reasoning      string     `json:"reasoning"`
}

// PlanStep represents a single step in the execution plan.
type PlanStep struct {
	Agent string `json:"agent"`
	Task  string `json:"task"`
	Order int    `json:"order"`
}

// StepResult represents the result of executing a step.
type StepResult struct {
	Agent   string
	Task    string
	Content string
	Success bool
	Error   string
}

// Config extends agent config with orchestrator-specific settings.
type Config struct {
	WebUIURL string
}

