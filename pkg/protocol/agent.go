package protocol

// AgentCard describes an agent's capabilities and metadata.
type AgentCard struct {
	// Basic information
	Name        string `json:"name" yaml:"name"`
	DisplayName string `json:"display_name,omitempty" yaml:"display_name"`
	Description string `json:"description,omitempty" yaml:"description"`
	Version     string `json:"version,omitempty" yaml:"version"`

	// Network information
	URL  string `json:"url" yaml:"url"`   // Base URL for A2A communication
	Port int    `json:"port" yaml:"port"` // A2A port

	// Capabilities (first-level index for tool lazy-loading)
	Capabilities []Capability `json:"capabilities,omitempty" yaml:"capabilities"`
	Tools        []Tool   `json:"tools,omitempty" yaml:"tools"`
	Skills       []Skill  `json:"skills,omitempty" yaml:"skills"`

	// Agent type
	Type AgentType `json:"type,omitempty" yaml:"type"`

	// Status
	Status AgentStatus `json:"status,omitempty" yaml:"status"`
}

// Capability is the first-level abstraction for grouping tools.
// It can be used as an index in LLM prompts.
type Capability struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description,omitempty" yaml:"description"`
	Keywords    []string `json:"keywords,omitempty" yaml:"keywords"`
}

// AgentType defines the type of agent.
type AgentType string

const (
	// AgentTypeNormal is a regular business agent.
	AgentTypeNormal AgentType = "normal"
	// AgentTypeOrchestrator is a special orchestrating agent.
	AgentTypeOrchestrator AgentType = "orchestrator"
)

// AgentStatus defines the status of an agent.
type AgentStatus string

const (
	AgentStatusOnline  AgentStatus = "online"
	AgentStatusOffline AgentStatus = "offline"
	AgentStatusBusy    AgentStatus = "busy"
)

// Tool describes a tool/function that an agent can execute.
type Tool struct {
	Name        string                 `json:"name" yaml:"name"`
	Description string                 `json:"description,omitempty" yaml:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty" yaml:"parameters"`
}

// Skill describes a high-level skill that an agent possesses.
type Skill struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description,omitempty" yaml:"description"`
	Keywords    []string `json:"keywords,omitempty" yaml:"keywords"`
}
