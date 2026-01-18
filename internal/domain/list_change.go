package domain

// ListChangeKind identifies the kind of list change.
type ListChangeKind string

const (
	// ListChangeTools indicates a tool list change.
	ListChangeTools ListChangeKind = "tools"
	// ListChangeResources indicates a resource list change.
	ListChangeResources ListChangeKind = "resources"
	// ListChangePrompts indicates a prompt list change.
	ListChangePrompts ListChangeKind = "prompts"
)

// ListChangeEvent describes a list change for a server.
type ListChangeEvent struct {
	Kind       ListChangeKind
	ServerType string
	SpecKey    string
}

// ListChangeEmitter emits list change events.
type ListChangeEmitter interface {
	EmitListChange(event ListChangeEvent)
}
