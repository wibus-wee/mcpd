package gateway

import (
	"context"
	"errors"

	"mcpv/internal/infra/rpc"
	controlv1 "mcpv/pkg/api/control/v1"
)

func (g *Gateway) listAllResources(ctx context.Context, client *rpc.Client) (*controlv1.ResourcesSnapshot, error) {
	cursor := ""
	var combined []*controlv1.ResourceDefinition
	etag := ""
	etagSet := false

	for {
		resp, err := client.Control().ListResources(ctx, &controlv1.ListResourcesRequest{
			Caller: g.caller,
			Cursor: cursor,
		})
		if err != nil {
			return nil, err
		}
		if resp != nil && resp.GetSnapshot() != nil {
			pageETag := resp.GetSnapshot().GetEtag()
			if !etagSet {
				etag = pageETag
				etagSet = true
			} else if pageETag != etag {
				return nil, errors.New("resource snapshot changed during pagination")
			}
			combined = append(combined, resp.GetSnapshot().GetResources()...)
		}
		if resp == nil || resp.GetNextCursor() == "" {
			break
		}
		cursor = resp.GetNextCursor()
	}

	return &controlv1.ResourcesSnapshot{
		Etag:      etag,
		Resources: combined,
	}, nil
}

func (g *Gateway) listAllPrompts(ctx context.Context, client *rpc.Client) (*controlv1.PromptsSnapshot, error) {
	cursor := ""
	var combined []*controlv1.PromptDefinition
	etag := ""
	etagSet := false

	for {
		resp, err := client.Control().ListPrompts(ctx, &controlv1.ListPromptsRequest{
			Caller: g.caller,
			Cursor: cursor,
		})
		if err != nil {
			return nil, err
		}
		if resp != nil && resp.GetSnapshot() != nil {
			pageETag := resp.GetSnapshot().GetEtag()
			if !etagSet {
				etag = pageETag
				etagSet = true
			} else if pageETag != etag {
				return nil, errors.New("prompt snapshot changed during pagination")
			}
			combined = append(combined, resp.GetSnapshot().GetPrompts()...)
		}
		if resp == nil || resp.GetNextCursor() == "" {
			break
		}
		cursor = resp.GetNextCursor()
	}

	return &controlv1.PromptsSnapshot{
		Etag:    etag,
		Prompts: combined,
	}, nil
}
