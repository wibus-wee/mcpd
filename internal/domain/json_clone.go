package domain

// CloneJSONValue deep-copies JSON-like values.
func CloneJSONValue(value any) any {
	switch typed := value.(type) {
	case nil:
		return nil
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, val := range typed {
			out[key] = CloneJSONValue(val)
		}
		return out
	case map[string]string:
		out := make(map[string]string, len(typed))
		for key, val := range typed {
			out[key] = val
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, val := range typed {
			out[i] = CloneJSONValue(val)
		}
		return out
	case []string:
		out := make([]string, len(typed))
		copy(out, typed)
		return out
	case []byte:
		out := make([]byte, len(typed))
		copy(out, typed)
		return out
	default:
		return typed
	}
}

// CloneToolDefinition deep-copies a tool definition.
func CloneToolDefinition(tool ToolDefinition) ToolDefinition {
	out := tool
	out.InputSchema = CloneJSONValue(tool.InputSchema)
	out.OutputSchema = CloneJSONValue(tool.OutputSchema)
	out.Meta = cloneMeta(tool.Meta)
	out.Annotations = cloneToolAnnotations(tool.Annotations)
	return out
}

// CloneToolSnapshot deep-copies a tool snapshot.
func CloneToolSnapshot(snapshot ToolSnapshot) ToolSnapshot {
	tools := make([]ToolDefinition, 0, len(snapshot.Tools))
	for _, tool := range snapshot.Tools {
		tools = append(tools, CloneToolDefinition(tool))
	}
	return ToolSnapshot{ETag: snapshot.ETag, Tools: tools}
}

// CloneResourceDefinition deep-copies a resource definition.
func CloneResourceDefinition(resource ResourceDefinition) ResourceDefinition {
	out := resource
	out.Meta = cloneMeta(resource.Meta)
	out.Annotations = cloneAnnotations(resource.Annotations)
	return out
}

// CloneResourceSnapshot deep-copies a resource snapshot.
func CloneResourceSnapshot(snapshot ResourceSnapshot) ResourceSnapshot {
	resources := make([]ResourceDefinition, 0, len(snapshot.Resources))
	for _, resource := range snapshot.Resources {
		resources = append(resources, CloneResourceDefinition(resource))
	}
	return ResourceSnapshot{ETag: snapshot.ETag, Resources: resources}
}

// ClonePromptDefinition deep-copies a prompt definition.
func ClonePromptDefinition(prompt PromptDefinition) PromptDefinition {
	out := prompt
	out.Meta = cloneMeta(prompt.Meta)
	out.Arguments = clonePromptArguments(prompt.Arguments)
	return out
}

// ClonePromptSnapshot deep-copies a prompt snapshot.
func ClonePromptSnapshot(snapshot PromptSnapshot) PromptSnapshot {
	prompts := make([]PromptDefinition, 0, len(snapshot.Prompts))
	for _, prompt := range snapshot.Prompts {
		prompts = append(prompts, ClonePromptDefinition(prompt))
	}
	return PromptSnapshot{ETag: snapshot.ETag, Prompts: prompts}
}

func cloneMeta(meta Meta) Meta {
	if meta == nil {
		return nil
	}
	cloned := CloneJSONValue(map[string]any(meta))
	if typed, ok := cloned.(map[string]any); ok {
		return Meta(typed)
	}
	return nil
}

func cloneToolAnnotations(ann *ToolAnnotations) *ToolAnnotations {
	if ann == nil {
		return nil
	}
	out := *ann
	if ann.DestructiveHint != nil {
		val := *ann.DestructiveHint
		out.DestructiveHint = &val
	}
	if ann.OpenWorldHint != nil {
		val := *ann.OpenWorldHint
		out.OpenWorldHint = &val
	}
	return &out
}

func cloneAnnotations(ann *Annotations) *Annotations {
	if ann == nil {
		return nil
	}
	out := *ann
	if len(ann.Audience) > 0 {
		out.Audience = append([]Role(nil), ann.Audience...)
	}
	return &out
}

func clonePromptArguments(args []PromptArgument) []PromptArgument {
	if len(args) == 0 {
		return nil
	}
	out := make([]PromptArgument, len(args))
	copy(out, args)
	return out
}
