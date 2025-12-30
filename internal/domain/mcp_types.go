package domain

type Meta map[string]any

type Role string

type Annotations struct {
	Audience     []Role
	LastModified string
	Priority     float64
}

type ToolAnnotations struct {
	DestructiveHint *bool
	IdempotentHint  bool
	OpenWorldHint   *bool
	ReadOnlyHint    bool
	Title           string
}

type PromptArgument struct {
	Name        string
	Title       string
	Description string
	Required    bool
}
