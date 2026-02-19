package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/AdrianWangs/ai-nexus/pkg/llm"
)

// Capability is the first-level abstraction for grouping tools.
// It is intentionally lightweight so it can be sent to LLM as an index.
type Capability struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
}

// ToolFunc is a function that can be called as a tool.
// The args map is decoded from JSON object arguments.
type ToolFunc func(ctx context.Context, args map[string]interface{}) (string, error)

// ToolDefinition holds the definition and implementation of a tool.
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Capability  string                 `json:"capability,omitempty"`
	Func        ToolFunc               `json:"-"`
}

// Registry manages capabilities and tools.
// It is concurrency-safe.
type Registry struct {
	mu sync.RWMutex

	capabilities      map[string]Capability
	toolsByName       map[string]*ToolDefinition
	toolsByCapability map[string][]*ToolDefinition
}

// NewRegistry creates a new tool registry.
func NewRegistry() *Registry {
	return &Registry{
		capabilities:      make(map[string]Capability),
		toolsByName:       make(map[string]*ToolDefinition),
		toolsByCapability: make(map[string][]*ToolDefinition),
	}
}

// RegisterCapability registers (or overwrites) a capability.
func (r *Registry) RegisterCapability(cap Capability) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.capabilities[cap.Name] = cap
}

// ListCapabilities returns all capabilities.
func (r *Registry) ListCapabilities() []Capability {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Capability, 0, len(r.capabilities))
	for _, c := range r.capabilities {
		out = append(out, c)
	}
	return out
}

// CapabilityIndex returns a capability list suitable for exposing to other components
// (includes implicit capabilities inferred from registered tools).
func (r *Registry) CapabilityIndex() []Capability {
	r.mu.RLock()
	defer r.mu.RUnlock()

	merged := make(map[string]Capability, len(r.capabilities)+len(r.toolsByCapability))
	for name, c := range r.capabilities {
		merged[name] = c
	}
	for name := range r.toolsByCapability {
		if _, ok := merged[name]; !ok {
			merged[name] = Capability{Name: name}
		}
	}

	out := make([]Capability, 0, len(merged))
	for _, c := range merged {
		out = append(out, c)
	}
	return out
}

// Register registers a tool definition.
// If capability is empty, it will be put into the default capability "default".
func (r *Registry) Register(capability string, tool *ToolDefinition) {
	if capability == "" {
		capability = "default"
	}
	tool.Capability = capability

	r.mu.Lock()
	defer r.mu.Unlock()

	r.toolsByName[tool.Name] = tool
	r.toolsByCapability[capability] = append(r.toolsByCapability[capability], tool)
}

// Get retrieves a tool by name.
func (r *Registry) Get(name string) (*ToolDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.toolsByName[name]
	return tool, ok
}

// List returns all registered tools.
func (r *Registry) List() []*ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]*ToolDefinition, 0, len(r.toolsByName))
	for _, tool := range r.toolsByName {
		tools = append(tools, tool)
	}
	return tools
}

// ListByCapability returns tools under a capability.
func (r *Registry) ListByCapability(capability string) []*ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := r.toolsByCapability[capability]
	out := make([]*ToolDefinition, 0, len(tools))
	out = append(out, tools...)
	return out
}

// ToLLMTools converts tools to LLM tool format.
// If capability is empty, returns all tools.
func (r *Registry) ToLLMTools(capability string) []llm.Tool {
	var src []*ToolDefinition
	if capability == "" {
		src = r.List()
	} else {
		src = r.ListByCapability(capability)
	}

	tools := make([]llm.Tool, 0, len(src))
	for _, tool := range src {
		tools = append(tools, llm.Tool{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		})
	}
	return tools
}

// Execute executes a tool call.
func (r *Registry) Execute(ctx context.Context, name string, argsJSON string) (string, error) {
	tool, ok := r.Get(name)
	if !ok {
		return "", fmt.Errorf("tool not found: %s", name)
	}

	var args map[string]interface{}
	if argsJSON != "" && argsJSON != "{}" {
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return "", fmt.Errorf("failed to parse arguments: %w", err)
		}
	}

	return tool.Func(ctx, args)
}

// RegisterFunc registers a Go function as a tool.
// The function should have the signature: func(ctx context.Context, args *T) (string, error)
// where T is a struct with JSON tags describing the parameters.
func (r *Registry) RegisterFunc(capability, name, description string, fn interface{}) error {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("expected function, got %s", fnType.Kind())
	}

	if fnType.NumIn() != 2 || fnType.NumOut() != 2 {
		return fmt.Errorf("function must have signature: func(ctx context.Context, args *T) (string, error)")
	}

	// Extract parameter schema from the args struct
	argsType := fnType.In(1)
	if argsType.Kind() == reflect.Ptr {
		argsType = argsType.Elem()
	}

	parameters := extractParametersSchema(argsType)
	fnValue := reflect.ValueOf(fn)

	r.Register(capability, &ToolDefinition{
		Name:        name,
		Description: description,
		Parameters:  parameters,
		Func: func(ctx context.Context, args map[string]interface{}) (string, error) {
			// Create args struct instance
			argsPtr := reflect.New(argsType)

			// Populate from map
			argsJSONBytes, _ := json.Marshal(args)
			if err := json.Unmarshal(argsJSONBytes, argsPtr.Interface()); err != nil {
				return "", fmt.Errorf("failed to parse arguments: %w", err)
			}

			// Call the function
			results := fnValue.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				argsPtr,
			})

			// Handle results
			if !results[1].IsNil() {
				return "", results[1].Interface().(error)
			}
			return results[0].String(), nil
		},
	})

	return nil
}

// extractParametersSchema extracts JSON schema from a struct type.
func extractParametersSchema(t reflect.Type) map[string]interface{} {
	properties := make(map[string]interface{})
	required := make([]string, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Parse json tag
		parts := strings.Split(jsonTag, ",")
		fieldName := parts[0]
		isOptional := false
		for _, part := range parts[1:] {
			if part == "omitempty" {
				isOptional = true
			}
		}

		// Get description from tag (support both "description" and "desc")
		description := field.Tag.Get("description")
		if description == "" {
			description = field.Tag.Get("desc")
		}

		// Determine type
		fieldType := "string"
		switch field.Type.Kind() {
		case reflect.String:
			fieldType = "string"
		case reflect.Int, reflect.Int32, reflect.Int64:
			fieldType = "integer"
		case reflect.Float32, reflect.Float64:
			fieldType = "number"
		case reflect.Bool:
			fieldType = "boolean"
		case reflect.Slice:
			fieldType = "array"
		case reflect.Map, reflect.Struct:
			fieldType = "object"
		}

		prop := map[string]interface{}{
			"type": fieldType,
		}
		if description != "" {
			prop["description"] = description
		}

		properties[fieldName] = prop
		if !isOptional {
			required = append(required, fieldName)
		}
	}

	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}

	return schema
}
