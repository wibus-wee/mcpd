package domain

type ListChangeKind string

const (
	ListChangeTools     ListChangeKind = "tools"
	ListChangeResources ListChangeKind = "resources"
	ListChangePrompts   ListChangeKind = "prompts"
)

type ListChangeEvent struct {
	Kind       ListChangeKind
	ServerType string
	SpecKey    string
}

type ListChangeEmitter interface {
	EmitListChange(event ListChangeEvent)
}
