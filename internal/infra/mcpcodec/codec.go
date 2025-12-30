package mcpcodec

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"mcpd/internal/domain"
)

func ToolFromMCP(tool *mcp.Tool) domain.ToolDefinition {
	if tool == nil {
		return domain.ToolDefinition{}
	}
	return domain.ToolDefinition{
		Name:         tool.Name,
		Description:  tool.Description,
		InputSchema:  domain.CloneJSONValue(tool.InputSchema),
		OutputSchema: domain.CloneJSONValue(tool.OutputSchema),
		Title:        tool.Title,
		Annotations:  toolAnnotationsFromMCP(tool.Annotations),
		Meta:         metaFromMCP(tool.Meta),
	}
}

func ResourceFromMCP(resource *mcp.Resource) domain.ResourceDefinition {
	if resource == nil {
		return domain.ResourceDefinition{}
	}
	return domain.ResourceDefinition{
		URI:         resource.URI,
		Name:        resource.Name,
		Title:       resource.Title,
		Description: resource.Description,
		MIMEType:    resource.MIMEType,
		Size:        resource.Size,
		Annotations: annotationsFromMCP(resource.Annotations),
		Meta:        metaFromMCP(resource.Meta),
	}
}

func PromptFromMCP(prompt *mcp.Prompt) domain.PromptDefinition {
	if prompt == nil {
		return domain.PromptDefinition{}
	}
	return domain.PromptDefinition{
		Name:        prompt.Name,
		Title:       prompt.Title,
		Description: prompt.Description,
		Arguments:   promptArgumentsFromMCP(prompt.Arguments),
		Meta:        metaFromMCP(prompt.Meta),
	}
}

func MarshalToolDefinition(tool domain.ToolDefinition) ([]byte, error) {
	wire := toolToMCP(tool)
	return json.Marshal(&wire)
}

func MarshalResourceDefinition(resource domain.ResourceDefinition) ([]byte, error) {
	wire := resourceToMCP(resource)
	return json.Marshal(&wire)
}

func MarshalPromptDefinition(prompt domain.PromptDefinition) ([]byte, error) {
	wire := promptToMCP(prompt)
	return json.Marshal(&wire)
}

func MustMarshalToolDefinition(tool domain.ToolDefinition) []byte {
	raw, _ := MarshalToolDefinition(tool)
	return raw
}

func MustMarshalResourceDefinition(resource domain.ResourceDefinition) []byte {
	raw, _ := MarshalResourceDefinition(resource)
	return raw
}

func MustMarshalPromptDefinition(prompt domain.PromptDefinition) []byte {
	raw, _ := MarshalPromptDefinition(prompt)
	return raw
}

func HashToolDefinition(tool domain.ToolDefinition) string {
	raw, err := MarshalToolDefinition(tool)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func HashResourceDefinition(resource domain.ResourceDefinition) string {
	raw, err := MarshalResourceDefinition(resource)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func HashPromptDefinition(prompt domain.PromptDefinition) string {
	raw, err := MarshalPromptDefinition(prompt)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func HashToolDefinitions(tools []domain.ToolDefinition) string {
	hasher := sha256.New()
	for _, tool := range tools {
		raw, err := MarshalToolDefinition(tool)
		if err != nil {
			continue
		}
		_, _ = hasher.Write(raw)
		_, _ = hasher.Write([]byte{0})
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func HashResourceDefinitions(resources []domain.ResourceDefinition) string {
	hasher := sha256.New()
	for _, resource := range resources {
		raw, err := MarshalResourceDefinition(resource)
		if err != nil {
			continue
		}
		_, _ = hasher.Write(raw)
		_, _ = hasher.Write([]byte{0})
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func HashPromptDefinitions(prompts []domain.PromptDefinition) string {
	hasher := sha256.New()
	for _, prompt := range prompts {
		raw, err := MarshalPromptDefinition(prompt)
		if err != nil {
			continue
		}
		_, _ = hasher.Write(raw)
		_, _ = hasher.Write([]byte{0})
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func toolToMCP(tool domain.ToolDefinition) mcp.Tool {
	return mcp.Tool{
		Meta:         metaToMCP(tool.Meta),
		Annotations:  toolAnnotationsToMCP(tool.Annotations),
		Description:  tool.Description,
		InputSchema:  tool.InputSchema,
		Name:         tool.Name,
		OutputSchema: tool.OutputSchema,
		Title:        tool.Title,
	}
}

func resourceToMCP(resource domain.ResourceDefinition) mcp.Resource {
	return mcp.Resource{
		Meta:        metaToMCP(resource.Meta),
		Annotations: annotationsToMCP(resource.Annotations),
		Description: resource.Description,
		MIMEType:    resource.MIMEType,
		Name:        resource.Name,
		Size:        resource.Size,
		Title:       resource.Title,
		URI:         resource.URI,
	}
}

func promptToMCP(prompt domain.PromptDefinition) mcp.Prompt {
	return mcp.Prompt{
		Meta:        metaToMCP(prompt.Meta),
		Arguments:   promptArgumentsToMCP(prompt.Arguments),
		Description: prompt.Description,
		Name:        prompt.Name,
		Title:       prompt.Title,
	}
}

func metaFromMCP(meta mcp.Meta) domain.Meta {
	if meta == nil {
		return nil
	}
	cloned := domain.CloneJSONValue(map[string]any(meta))
	if typed, ok := cloned.(map[string]any); ok {
		return domain.Meta(typed)
	}
	return nil
}

func metaToMCP(meta domain.Meta) mcp.Meta {
	if meta == nil {
		return nil
	}
	cloned := domain.CloneJSONValue(map[string]any(meta))
	if typed, ok := cloned.(map[string]any); ok {
		return mcp.Meta(typed)
	}
	return nil
}

func annotationsFromMCP(ann *mcp.Annotations) *domain.Annotations {
	if ann == nil {
		return nil
	}
	out := domain.Annotations{
		Audience:     make([]domain.Role, 0, len(ann.Audience)),
		LastModified: ann.LastModified,
		Priority:     ann.Priority,
	}
	for _, role := range ann.Audience {
		out.Audience = append(out.Audience, domain.Role(role))
	}
	return &out
}

func annotationsToMCP(ann *domain.Annotations) *mcp.Annotations {
	if ann == nil {
		return nil
	}
	out := mcp.Annotations{
		Audience:     make([]mcp.Role, 0, len(ann.Audience)),
		LastModified: ann.LastModified,
		Priority:     ann.Priority,
	}
	for _, role := range ann.Audience {
		out.Audience = append(out.Audience, mcp.Role(role))
	}
	return &out
}

func toolAnnotationsFromMCP(ann *mcp.ToolAnnotations) *domain.ToolAnnotations {
	if ann == nil {
		return nil
	}
	out := domain.ToolAnnotations{
		IdempotentHint: ann.IdempotentHint,
		ReadOnlyHint:   ann.ReadOnlyHint,
		Title:          ann.Title,
	}
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

func toolAnnotationsToMCP(ann *domain.ToolAnnotations) *mcp.ToolAnnotations {
	if ann == nil {
		return nil
	}
	out := mcp.ToolAnnotations{
		IdempotentHint: ann.IdempotentHint,
		ReadOnlyHint:   ann.ReadOnlyHint,
		Title:          ann.Title,
	}
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

func promptArgumentsFromMCP(args []*mcp.PromptArgument) []domain.PromptArgument {
	if len(args) == 0 {
		return nil
	}
	out := make([]domain.PromptArgument, 0, len(args))
	for _, arg := range args {
		if arg == nil {
			continue
		}
		out = append(out, domain.PromptArgument{
			Name:        arg.Name,
			Title:       arg.Title,
			Description: arg.Description,
			Required:    arg.Required,
		})
	}
	return out
}

func promptArgumentsToMCP(args []domain.PromptArgument) []*mcp.PromptArgument {
	if len(args) == 0 {
		return nil
	}
	out := make([]*mcp.PromptArgument, 0, len(args))
	for _, arg := range args {
		argCopy := arg
		out = append(out, &mcp.PromptArgument{
			Name:        argCopy.Name,
			Title:       argCopy.Title,
			Description: argCopy.Description,
			Required:    argCopy.Required,
		})
	}
	return out
}
