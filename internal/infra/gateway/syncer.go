package gateway

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"mcpv/internal/infra/retry"
	"mcpv/internal/infra/rpc"
	controlv1 "mcpv/pkg/api/control/v1"
)

type snapshotStream[S any] interface {
	Recv() (S, error)
}

type syncMessages struct {
	listFailed       string
	watchFailed      string
	watchInterrupted string
}

func (g *Gateway) syncTools(ctx context.Context) {
	runSnapshotSync(ctx, g,
		syncMessages{
			listFailed:       "rpc list tools failed",
			watchFailed:      "rpc watch tools failed",
			watchInterrupted: "rpc tool watch interrupted",
		},
		g.listToolsSnapshot,
		g.watchToolsSnapshot,
		g.applyToolsSnapshot,
		g.markToolsReady,
	)
}

func (g *Gateway) syncResources(ctx context.Context) {
	runSnapshotSync(ctx, g,
		syncMessages{
			listFailed:       "rpc list resources failed",
			watchFailed:      "rpc watch resources failed",
			watchInterrupted: "rpc resource watch interrupted",
		},
		g.listResourcesSnapshot,
		g.watchResourcesSnapshot,
		g.applyResourcesSnapshot,
		nil,
	)
}

func (g *Gateway) syncPrompts(ctx context.Context) {
	runSnapshotSync(ctx, g,
		syncMessages{
			listFailed:       "rpc list prompts failed",
			watchFailed:      "rpc watch prompts failed",
			watchInterrupted: "rpc prompt watch interrupted",
		},
		g.listPromptsSnapshot,
		g.watchPromptsSnapshot,
		g.applyPromptsSnapshot,
		nil,
	)
}

func runSnapshotSync[S any](
	ctx context.Context,
	g *Gateway,
	messages syncMessages,
	listFn func(context.Context, *rpc.Client) (S, error),
	watchFn func(context.Context, *rpc.Client, string) (snapshotStream[S], error),
	applyFn func(S) string,
	readyFn func(),
) {
	backoff := retry.NewBackoff(retry.Policy{
		BaseDelay: time.Second,
		MaxDelay:  30 * time.Second,
	})
	lastETag := ""

	for {
		if ctx.Err() != nil {
			return
		}

		client, err := g.clients.get(ctx)
		if err != nil {
			g.logger.Warn("rpc connect failed", zap.Error(err))
			backoff.Sleep(ctx)
			continue
		}

		snapshot, err := listFn(ctx, client)
		if err != nil {
			if status.Code(err) == codes.FailedPrecondition {
				if regErr := g.registerCaller(ctx); regErr == nil {
					continue
				}
			}
			g.logger.Warn(messages.listFailed, zap.Error(err))
			g.clients.reset()
			backoff.Sleep(ctx)
			continue
		}
		if etag := applyFn(snapshot); etag != "" {
			lastETag = etag
		}
		if readyFn != nil {
			readyFn()
		}

		stream, err := watchFn(ctx, client, lastETag)
		if err != nil {
			if status.Code(err) == codes.FailedPrecondition {
				if regErr := g.registerCaller(ctx); regErr == nil {
					continue
				}
			}
			g.logger.Warn(messages.watchFailed, zap.Error(err))
			g.clients.reset()
			backoff.Sleep(ctx)
			continue
		}

		backoff.Reset()

		for {
			snapshot, err := stream.Recv()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				if status.Code(err) == codes.Canceled {
					return
				}
				g.logger.Warn(messages.watchInterrupted, zap.Error(err))
				g.clients.reset()
				backoff.Sleep(ctx)
				break
			}
			if etag := applyFn(snapshot); etag != "" {
				lastETag = etag
			}
		}
	}
}

func (g *Gateway) listToolsSnapshot(ctx context.Context, client *rpc.Client) (*controlv1.ToolsSnapshot, error) {
	resp, err := client.Control().ListTools(ctx, &controlv1.ListToolsRequest{Caller: g.caller})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, nil
	}
	return resp.GetSnapshot(), nil
}

func (g *Gateway) listResourcesSnapshot(ctx context.Context, client *rpc.Client) (*controlv1.ResourcesSnapshot, error) {
	return g.listAllResources(ctx, client)
}

func (g *Gateway) listPromptsSnapshot(ctx context.Context, client *rpc.Client) (*controlv1.PromptsSnapshot, error) {
	return g.listAllPrompts(ctx, client)
}

func (g *Gateway) watchToolsSnapshot(ctx context.Context, client *rpc.Client, lastETag string) (snapshotStream[*controlv1.ToolsSnapshot], error) {
	return client.Control().WatchTools(ctx, &controlv1.WatchToolsRequest{
		Caller:   g.caller,
		LastEtag: lastETag,
	})
}

func (g *Gateway) watchResourcesSnapshot(ctx context.Context, client *rpc.Client, lastETag string) (snapshotStream[*controlv1.ResourcesSnapshot], error) {
	return client.Control().WatchResources(ctx, &controlv1.WatchResourcesRequest{
		Caller:   g.caller,
		LastEtag: lastETag,
	})
}

func (g *Gateway) watchPromptsSnapshot(ctx context.Context, client *rpc.Client, lastETag string) (snapshotStream[*controlv1.PromptsSnapshot], error) {
	return client.Control().WatchPrompts(ctx, &controlv1.WatchPromptsRequest{
		Caller:   g.caller,
		LastEtag: lastETag,
	})
}

func (g *Gateway) applyToolsSnapshot(snapshot *controlv1.ToolsSnapshot) string {
	if snapshot == nil {
		return ""
	}
	g.registry.ApplySnapshot(snapshot)
	return snapshot.GetEtag()
}

func (g *Gateway) applyResourcesSnapshot(snapshot *controlv1.ResourcesSnapshot) string {
	if snapshot == nil {
		return ""
	}
	g.resources.ApplySnapshot(snapshot)
	return snapshot.GetEtag()
}

func (g *Gateway) applyPromptsSnapshot(snapshot *controlv1.PromptsSnapshot) string {
	if snapshot == nil {
		return ""
	}
	g.prompts.ApplySnapshot(snapshot)
	return snapshot.GetEtag()
}
