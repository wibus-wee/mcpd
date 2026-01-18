package domain

// Meta carries extension metadata for MCP entities.
type Meta map[string]any

// Role identifies an MCP audience role.
type Role string

// Annotations captures shared annotation fields.
type Annotations struct {
	Audience     []Role
	LastModified string
	Priority     float64
}

// ToolAnnotations captures tool-specific annotation fields.
type ToolAnnotations struct {
	DestructiveHint *bool
	IdempotentHint  bool
	OpenWorldHint   *bool
	ReadOnlyHint    bool
	Title           string
}

// PromptArgument describes a prompt argument.
type PromptArgument struct {
	Name        string
	Title       string
	Description string
	Required    bool
}
