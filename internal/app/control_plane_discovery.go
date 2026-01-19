package app

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"

	"mcpd/internal/domain"
	"mcpd/internal/infra/hashutil"
)

type discoveryService struct {
	state    *controlPlaneState
	registry *callerRegistry

	mu               sync.Mutex
	toolWatchers     map[string][]*callerWatcher[domain.ToolSnapshot]
	resourceWatchers map[string][]*callerWatcher[domain.ResourceSnapshot]
	promptWatchers   map[string][]*callerWatcher[domain.PromptSnapshot]
}

const snapshotPageSize = 200

func newDiscoveryService(state *controlPlaneState, registry *callerRegistry) *discoveryService {
	return &discoveryService{
		state:            state,
		registry:         registry,
		toolWatchers:     make(map[string][]*callerWatcher[domain.ToolSnapshot]),
		resourceWatchers: make(map[string][]*callerWatcher[domain.ResourceSnapshot]),
		promptWatchers:   make(map[string][]*callerWatcher[domain.PromptSnapshot]),
	}
}

// indexSubscriber is an interface for types that can be subscribed to for snapshots.
type indexSubscriber[T any] interface {
	Subscribe(ctx context.Context) <-chan T
}

// callerWatcher manages a caller's subscription that can switch between profile indexes.
// Data flows directly from the profile's GenericIndex to the output channel.
type callerWatcher[T any] struct {
	caller   string
	output   chan T
	getIndex func(*profileRuntime) indexSubscriber[T]

	mu        sync.Mutex
	subCtx    context.Context
	subCancel context.CancelFunc
}

// switchProfile cancels the current subscription and subscribes to the new profile's index.
// The new index will immediately send its current snapshot.
func (w *callerWatcher[T]) switchProfile(ctx context.Context, profile *profileRuntime) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Cancel old subscription
	if w.subCancel != nil {
		w.subCancel()
	}

	if profile == nil {
		return
	}

	index := w.getIndex(profile)
	if index == nil {
		return
	}

	// Create new subscription context
	w.subCtx, w.subCancel = context.WithCancel(ctx)

	// Subscribe directly to the profile's index
	// GenericIndex.Subscribe() immediately sends current snapshot
	indexCh := index.Subscribe(w.subCtx)

	// Forward from index channel to output channel
	go func() {
		for {
			select {
			case <-w.subCtx.Done():
				return
			case snapshot, ok := <-indexCh:
				if !ok {
					return
				}
				select {
				case w.output <- snapshot:
				default:
					// Drop if output is full (non-blocking)
				}
			}
		}
	}()
}

// StartProfileChangeListener watches profile changes for caller subscriptions.
func (d *discoveryService) StartProfileChangeListener(ctx context.Context) {
	changes := d.registry.WatchProfileChanges(ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-changes:
				if !ok {
					return
				}
				d.handleProfileChange(ctx, event)
			}
		}
	}()
}

func (d *discoveryService) handleProfileChange(ctx context.Context, event profileChangeEvent) {
	newProfile, _ := d.state.Profile(event.NewProfile)

	d.mu.Lock()
	defer d.mu.Unlock()

	// Switch all tool watchers for this caller
	for _, w := range d.toolWatchers[event.Caller] {
		w.switchProfile(ctx, newProfile)
	}

	// Switch all resource watchers for this caller
	for _, w := range d.resourceWatchers[event.Caller] {
		w.switchProfile(ctx, newProfile)
	}

	// Switch all prompt watchers for this caller
	for _, w := range d.promptWatchers[event.Caller] {
		w.switchProfile(ctx, newProfile)
	}
}

// ListTools lists tools for a caller.
func (d *discoveryService) ListTools(ctx context.Context, caller string) (domain.ToolSnapshot, error) {
	profile, err := d.registry.resolveProfile(caller)
	if err != nil {
		return domain.ToolSnapshot{}, err
	}
	if profile.tools == nil {
		return domain.ToolSnapshot{}, nil
	}
	return profile.tools.Snapshot(), nil
}

// ListToolsAllProfiles lists tools across all profiles.
func (d *discoveryService) ListToolsAllProfiles(ctx context.Context) (domain.ToolSnapshot, error) {
	profileNames := d.registry.activeProfileNames()
	if len(profileNames) == 0 {
		return domain.ToolSnapshot{}, nil
	}

	merged := make([]domain.ToolDefinition, 0)
	seen := make(map[string]struct{})

	for _, name := range profileNames {
		runtime, ok := d.state.Profile(name)
		if !ok || runtime.tools == nil {
			continue
		}
		snapshot := runtime.tools.Snapshot()
		for _, tool := range snapshot.Tools {
			key := tool.SpecKey
			if key == "" {
				key = tool.ServerName
			}
			if key == "" {
				key = tool.Name
			}
			dedupeKey := key + "\x00" + tool.Name
			if _, ok := seen[dedupeKey]; ok {
				continue
			}
			seen[dedupeKey] = struct{}{}
			merged = append(merged, tool)
		}
	}

	if len(merged) == 0 {
		return domain.ToolSnapshot{}, nil
	}

	sort.Slice(merged, func(i, j int) bool {
		if merged[i].SpecKey != merged[j].SpecKey {
			return merged[i].SpecKey < merged[j].SpecKey
		}
		if merged[i].Name != merged[j].Name {
			return merged[i].Name < merged[j].Name
		}
		return merged[i].ServerName < merged[j].ServerName
	})

	return domain.ToolSnapshot{
		ETag:  hashutil.ToolETag(d.state.logger, merged),
		Tools: merged,
	}, nil
}

// ListToolCatalog returns the full tool catalog snapshot.
func (d *discoveryService) ListToolCatalog(ctx context.Context) (domain.ToolCatalogSnapshot, error) {
	profiles := d.state.Profiles()
	if len(profiles) == 0 {
		return domain.ToolCatalogSnapshot{}, nil
	}

	liveTools := make([]domain.ToolDefinition, 0)
	activeProfiles := d.registry.activeProfileNames()
	for _, name := range activeProfiles {
		runtime, ok := d.state.Profile(name)
		if !ok || runtime.tools == nil {
			continue
		}
		snapshot := runtime.tools.Snapshot()
		liveTools = append(liveTools, snapshot.Tools...)
	}

	cachedTools := make([]domain.ToolDefinition, 0)
	for _, runtime := range profiles {
		if runtime.tools == nil {
			continue
		}
		snapshot := runtime.tools.CachedSnapshot()
		if len(snapshot.Tools) == 0 {
			continue
		}
		cachedTools = append(cachedTools, snapshot.Tools...)
	}

	cachedAt := make(map[string]time.Time)
	if len(cachedTools) > 0 {
		cache := d.metadataCache()
		if cache != nil {
			for _, tool := range cachedTools {
				specKey := tool.SpecKey
				if specKey == "" {
					continue
				}
				if _, ok := cachedAt[specKey]; ok {
					continue
				}
				if ts, ok := cache.GetCachedAt(specKey); ok {
					cachedAt[specKey] = ts
				}
			}
		}
	}

	return buildToolCatalogSnapshot(d.state.logger, liveTools, cachedTools, cachedAt), nil
}

// WatchTools streams tool snapshots for a caller.
func (d *discoveryService) WatchTools(ctx context.Context, caller string) (<-chan domain.ToolSnapshot, error) {
	profile, err := d.registry.resolveProfile(caller)
	if err != nil {
		return closedToolSnapshotChannel(), err
	}

	watcher := &callerWatcher[domain.ToolSnapshot]{
		caller: caller,
		output: make(chan domain.ToolSnapshot, 1),
		getIndex: func(p *profileRuntime) indexSubscriber[domain.ToolSnapshot] {
			if p == nil || p.tools == nil {
				return nil
			}
			return p.tools
		},
	}

	// Initial subscription to current profile
	watcher.switchProfile(ctx, profile)

	// Register watcher for profile change notifications
	d.mu.Lock()
	d.toolWatchers[caller] = append(d.toolWatchers[caller], watcher)
	d.mu.Unlock()

	// Cleanup on context done
	go func() {
		<-ctx.Done()
		watcher.mu.Lock()
		if watcher.subCancel != nil {
			watcher.subCancel()
		}
		watcher.mu.Unlock()

		d.mu.Lock()
		watchers := d.toolWatchers[caller]
		for i, w := range watchers {
			if w == watcher {
				d.toolWatchers[caller] = append(watchers[:i], watchers[i+1:]...)
				break
			}
		}
		if len(d.toolWatchers[caller]) == 0 {
			delete(d.toolWatchers, caller)
		}
		d.mu.Unlock()
	}()

	return watcher.output, nil
}

// CallTool executes a tool on behalf of a caller.
func (d *discoveryService) CallTool(ctx context.Context, caller, name string, args json.RawMessage, routingKey string) (json.RawMessage, error) {
	profile, err := d.registry.resolveProfile(caller)
	if err != nil {
		return nil, err
	}
	ctx = domain.WithRouteContext(ctx, domain.RouteContext{Caller: caller, Profile: profile.name})
	ctx = domain.WithStartCause(ctx, domain.StartCause{
		Reason:   domain.StartCauseToolCall,
		Caller:   caller,
		ToolName: name,
		Profile:  profile.name,
	})
	if profile.tools == nil {
		return nil, domain.ErrToolNotFound
	}
	return profile.tools.CallTool(ctx, name, args, routingKey)
}

// CallToolAllProfiles executes a tool across all profiles.
func (d *discoveryService) CallToolAllProfiles(ctx context.Context, name string, args json.RawMessage, routingKey, specKey string) (json.RawMessage, error) {
	ctx = domain.WithStartCause(ctx, domain.StartCause{
		Reason:   domain.StartCauseToolCall,
		ToolName: name,
	})
	profileNames := d.registry.activeProfileNames()
	for _, profileName := range profileNames {
		runtime, ok := d.state.Profile(profileName)
		if !ok || runtime.tools == nil {
			continue
		}
		if specKey != "" && !d.registry.profileContainsSpecKey(runtime, specKey) {
			continue
		}
		result, err := runtime.tools.CallTool(ctx, name, args, routingKey)
		if err == nil {
			return result, nil
		}
		if errors.Is(err, domain.ErrToolNotFound) {
			continue
		}
		return nil, err
	}
	return nil, domain.ErrToolNotFound
}

// ListResources lists resources for a caller.
func (d *discoveryService) ListResources(ctx context.Context, caller string, cursor string) (domain.ResourcePage, error) {
	profile, err := d.registry.resolveProfile(caller)
	if err != nil {
		return domain.ResourcePage{}, err
	}
	if profile.resources == nil {
		return domain.ResourcePage{Snapshot: domain.ResourceSnapshot{}}, nil
	}
	snapshot := profile.resources.Snapshot()
	return paginateResources(snapshot, cursor)
}

// ListResourcesAllProfiles lists resources across all profiles.
func (d *discoveryService) ListResourcesAllProfiles(ctx context.Context, cursor string) (domain.ResourcePage, error) {
	profileNames := d.registry.activeProfileNames()
	if len(profileNames) == 0 {
		return domain.ResourcePage{Snapshot: domain.ResourceSnapshot{}}, nil
	}

	merged := make([]domain.ResourceDefinition, 0)
	seen := make(map[string]struct{})

	for _, profileName := range profileNames {
		runtime, ok := d.state.Profile(profileName)
		if !ok || runtime.resources == nil {
			continue
		}
		snapshot := runtime.resources.Snapshot()
		for _, resource := range snapshot.Resources {
			if resource.URI == "" {
				continue
			}
			if _, ok := seen[resource.URI]; ok {
				continue
			}
			seen[resource.URI] = struct{}{}
			merged = append(merged, resource)
		}
	}

	if len(merged) == 0 {
		return domain.ResourcePage{Snapshot: domain.ResourceSnapshot{}}, nil
	}

	sort.Slice(merged, func(i, j int) bool { return merged[i].URI < merged[j].URI })
	snapshot := domain.ResourceSnapshot{
		ETag:      hashutil.ResourceETag(d.state.logger, merged),
		Resources: merged,
	}
	return paginateResources(snapshot, cursor)
}

// WatchResources streams resource snapshots for a caller.
func (d *discoveryService) WatchResources(ctx context.Context, caller string) (<-chan domain.ResourceSnapshot, error) {
	profile, err := d.registry.resolveProfile(caller)
	if err != nil {
		return closedResourceSnapshotChannel(), err
	}

	watcher := &callerWatcher[domain.ResourceSnapshot]{
		caller: caller,
		output: make(chan domain.ResourceSnapshot, 1),
		getIndex: func(p *profileRuntime) indexSubscriber[domain.ResourceSnapshot] {
			if p == nil || p.resources == nil {
				return nil
			}
			return p.resources
		},
	}

	// Initial subscription to current profile
	watcher.switchProfile(ctx, profile)

	// Register watcher for profile change notifications
	d.mu.Lock()
	d.resourceWatchers[caller] = append(d.resourceWatchers[caller], watcher)
	d.mu.Unlock()

	// Cleanup on context done
	go func() {
		<-ctx.Done()
		watcher.mu.Lock()
		if watcher.subCancel != nil {
			watcher.subCancel()
		}
		watcher.mu.Unlock()

		d.mu.Lock()
		watchers := d.resourceWatchers[caller]
		for i, w := range watchers {
			if w == watcher {
				d.resourceWatchers[caller] = append(watchers[:i], watchers[i+1:]...)
				break
			}
		}
		if len(d.resourceWatchers[caller]) == 0 {
			delete(d.resourceWatchers, caller)
		}
		d.mu.Unlock()
	}()

	return watcher.output, nil
}

// ReadResource reads a resource on behalf of a caller.
func (d *discoveryService) ReadResource(ctx context.Context, caller, uri string) (json.RawMessage, error) {
	profile, err := d.registry.resolveProfile(caller)
	if err != nil {
		return nil, err
	}
	ctx = domain.WithRouteContext(ctx, domain.RouteContext{Caller: caller, Profile: profile.name})
	if profile.resources == nil {
		return nil, domain.ErrResourceNotFound
	}
	return profile.resources.ReadResource(ctx, uri)
}

// ReadResourceAllProfiles reads a resource across all profiles.
func (d *discoveryService) ReadResourceAllProfiles(ctx context.Context, uri, specKey string) (json.RawMessage, error) {
	profileNames := d.registry.activeProfileNames()
	for _, profileName := range profileNames {
		runtime, ok := d.state.Profile(profileName)
		if !ok || runtime.resources == nil {
			continue
		}
		if specKey != "" && !d.registry.profileContainsSpecKey(runtime, specKey) {
			continue
		}
		result, err := runtime.resources.ReadResource(ctx, uri)
		if err == nil {
			return result, nil
		}
		if errors.Is(err, domain.ErrResourceNotFound) {
			continue
		}
		return nil, err
	}
	return nil, domain.ErrResourceNotFound
}

// ListPrompts lists prompts for a caller.
func (d *discoveryService) ListPrompts(ctx context.Context, caller string, cursor string) (domain.PromptPage, error) {
	profile, err := d.registry.resolveProfile(caller)
	if err != nil {
		return domain.PromptPage{}, err
	}
	if profile.prompts == nil {
		return domain.PromptPage{Snapshot: domain.PromptSnapshot{}}, nil
	}
	snapshot := profile.prompts.Snapshot()
	return paginatePrompts(snapshot, cursor)
}

// ListPromptsAllProfiles lists prompts across all profiles.
func (d *discoveryService) ListPromptsAllProfiles(ctx context.Context, cursor string) (domain.PromptPage, error) {
	profileNames := d.registry.activeProfileNames()
	if len(profileNames) == 0 {
		return domain.PromptPage{Snapshot: domain.PromptSnapshot{}}, nil
	}

	merged := make([]domain.PromptDefinition, 0)
	seen := make(map[string]struct{})

	for _, profileName := range profileNames {
		runtime, ok := d.state.Profile(profileName)
		if !ok || runtime.prompts == nil {
			continue
		}
		snapshot := runtime.prompts.Snapshot()
		for _, prompt := range snapshot.Prompts {
			if prompt.Name == "" {
				continue
			}
			if _, ok := seen[prompt.Name]; ok {
				continue
			}
			seen[prompt.Name] = struct{}{}
			merged = append(merged, prompt)
		}
	}

	if len(merged) == 0 {
		return domain.PromptPage{Snapshot: domain.PromptSnapshot{}}, nil
	}

	sort.Slice(merged, func(i, j int) bool { return merged[i].Name < merged[j].Name })
	snapshot := domain.PromptSnapshot{
		ETag:    hashutil.PromptETag(d.state.logger, merged),
		Prompts: merged,
	}
	return paginatePrompts(snapshot, cursor)
}

// WatchPrompts streams prompt snapshots for a caller.
func (d *discoveryService) WatchPrompts(ctx context.Context, caller string) (<-chan domain.PromptSnapshot, error) {
	profile, err := d.registry.resolveProfile(caller)
	if err != nil {
		return closedPromptSnapshotChannel(), err
	}

	watcher := &callerWatcher[domain.PromptSnapshot]{
		caller: caller,
		output: make(chan domain.PromptSnapshot, 1),
		getIndex: func(p *profileRuntime) indexSubscriber[domain.PromptSnapshot] {
			if p == nil || p.prompts == nil {
				return nil
			}
			return p.prompts
		},
	}

	// Initial subscription to current profile
	watcher.switchProfile(ctx, profile)

	// Register watcher for profile change notifications
	d.mu.Lock()
	d.promptWatchers[caller] = append(d.promptWatchers[caller], watcher)
	d.mu.Unlock()

	// Cleanup on context done
	go func() {
		<-ctx.Done()
		watcher.mu.Lock()
		if watcher.subCancel != nil {
			watcher.subCancel()
		}
		watcher.mu.Unlock()

		d.mu.Lock()
		watchers := d.promptWatchers[caller]
		for i, w := range watchers {
			if w == watcher {
				d.promptWatchers[caller] = append(watchers[:i], watchers[i+1:]...)
				break
			}
		}
		if len(d.promptWatchers[caller]) == 0 {
			delete(d.promptWatchers, caller)
		}
		d.mu.Unlock()
	}()

	return watcher.output, nil
}

// GetPrompt resolves a prompt for a caller.
func (d *discoveryService) GetPrompt(ctx context.Context, caller, name string, args json.RawMessage) (json.RawMessage, error) {
	profile, err := d.registry.resolveProfile(caller)
	if err != nil {
		return nil, err
	}
	ctx = domain.WithRouteContext(ctx, domain.RouteContext{Caller: caller, Profile: profile.name})
	if profile.prompts == nil {
		return nil, domain.ErrPromptNotFound
	}
	return profile.prompts.GetPrompt(ctx, name, args)
}

// GetPromptAllProfiles resolves a prompt across all profiles.
func (d *discoveryService) GetPromptAllProfiles(ctx context.Context, name string, args json.RawMessage, specKey string) (json.RawMessage, error) {
	profileNames := d.registry.activeProfileNames()
	for _, profileName := range profileNames {
		runtime, ok := d.state.Profile(profileName)
		if !ok || runtime.prompts == nil {
			continue
		}
		if specKey != "" && !d.registry.profileContainsSpecKey(runtime, specKey) {
			continue
		}
		result, err := runtime.prompts.GetPrompt(ctx, name, args)
		if err == nil {
			return result, nil
		}
		if errors.Is(err, domain.ErrPromptNotFound) {
			continue
		}
		return nil, err
	}
	return nil, domain.ErrPromptNotFound
}

// GetToolSnapshotForCaller returns the tool snapshot for a caller.
func (d *discoveryService) GetToolSnapshotForCaller(caller string) (domain.ToolSnapshot, error) {
	profile, err := d.registry.resolveProfile(caller)
	if err != nil {
		return domain.ToolSnapshot{}, err
	}
	if profile.tools == nil {
		return domain.ToolSnapshot{}, nil
	}
	return profile.tools.Snapshot(), nil
}

func paginateResources(snapshot domain.ResourceSnapshot, cursor string) (domain.ResourcePage, error) {
	resources := snapshot.Resources
	start := 0
	if cursor != "" {
		start = indexAfterResourceCursor(resources, cursor)
		if start < 0 {
			return domain.ResourcePage{}, domain.ErrInvalidCursor
		}
	}

	end := start + snapshotPageSize
	if end > len(resources) {
		end = len(resources)
	}
	nextCursor := ""
	if end < len(resources) {
		nextCursor = resources[end-1].URI
	}
	page := domain.ResourceSnapshot{
		ETag:      snapshot.ETag,
		Resources: append([]domain.ResourceDefinition(nil), resources[start:end]...),
	}
	return domain.ResourcePage{Snapshot: page, NextCursor: nextCursor}, nil
}

func paginatePrompts(snapshot domain.PromptSnapshot, cursor string) (domain.PromptPage, error) {
	prompts := snapshot.Prompts
	start := 0
	if cursor != "" {
		start = indexAfterPromptCursor(prompts, cursor)
		if start < 0 {
			return domain.PromptPage{}, domain.ErrInvalidCursor
		}
	}

	end := start + snapshotPageSize
	if end > len(prompts) {
		end = len(prompts)
	}
	nextCursor := ""
	if end < len(prompts) {
		nextCursor = prompts[end-1].Name
	}
	page := domain.PromptSnapshot{
		ETag:    snapshot.ETag,
		Prompts: append([]domain.PromptDefinition(nil), prompts[start:end]...),
	}
	return domain.PromptPage{Snapshot: page, NextCursor: nextCursor}, nil
}

func indexAfterResourceCursor(resources []domain.ResourceDefinition, cursor string) int {
	idx := sort.Search(len(resources), func(i int) bool {
		return resources[i].URI >= cursor
	})
	if idx >= len(resources) || resources[idx].URI != cursor {
		return -1
	}
	return idx + 1
}

func indexAfterPromptCursor(prompts []domain.PromptDefinition, cursor string) int {
	idx := sort.Search(len(prompts), func(i int) bool {
		return prompts[i].Name >= cursor
	})
	if idx >= len(prompts) || prompts[idx].Name != cursor {
		return -1
	}
	return idx + 1
}

func (d *discoveryService) metadataCache() *domain.MetadataCache {
	if d == nil || d.state == nil || d.state.bootstrapManager == nil {
		return nil
	}
	return d.state.bootstrapManager.GetCache()
}

func buildToolCatalogSnapshot(logger *zap.Logger, liveTools []domain.ToolDefinition, cachedTools []domain.ToolDefinition, cachedAt map[string]time.Time) domain.ToolCatalogSnapshot {
	entriesByKey := make(map[string]domain.ToolCatalogEntry)

	for _, tool := range liveTools {
		if tool.Name == "" {
			continue
		}
		key := toolDedupKey(tool)
		if key == "" {
			continue
		}
		entriesByKey[key] = domain.ToolCatalogEntry{
			Definition: tool,
			Source:     domain.ToolSourceLive,
		}
	}

	for _, tool := range cachedTools {
		if tool.Name == "" {
			continue
		}
		key := toolDedupKey(tool)
		if key == "" {
			continue
		}
		if _, ok := entriesByKey[key]; ok {
			continue
		}
		entry := domain.ToolCatalogEntry{
			Definition: tool,
			Source:     domain.ToolSourceCache,
		}
		if tool.SpecKey != "" && cachedAt != nil {
			if ts, ok := cachedAt[tool.SpecKey]; ok {
				entry.CachedAt = ts
			}
		}
		entriesByKey[key] = entry
	}

	if len(entriesByKey) == 0 {
		return domain.ToolCatalogSnapshot{}
	}

	entries := make([]domain.ToolCatalogEntry, 0, len(entriesByKey))
	for _, entry := range entriesByKey {
		entries = append(entries, entry)
	}

	sort.Slice(entries, func(i, j int) bool {
		left := entries[i].Definition
		right := entries[j].Definition
		if left.SpecKey != right.SpecKey {
			return left.SpecKey < right.SpecKey
		}
		if left.Name != right.Name {
			return left.Name < right.Name
		}
		return left.ServerName < right.ServerName
	})

	tools := make([]domain.ToolDefinition, 0, len(entries))
	for _, entry := range entries {
		tools = append(tools, entry.Definition)
	}

	return domain.ToolCatalogSnapshot{
		ETag:  hashutil.ToolETag(logger, tools),
		Tools: entries,
	}
}

func toolDedupKey(tool domain.ToolDefinition) string {
	key := tool.SpecKey
	if key == "" {
		key = tool.ServerName
	}
	if key == "" {
		key = tool.Name
	}
	if key == "" || tool.Name == "" {
		return ""
	}
	return key + "\x00" + tool.Name
}

func closedToolSnapshotChannel() chan domain.ToolSnapshot {
	ch := make(chan domain.ToolSnapshot)
	close(ch)
	return ch
}

func closedResourceSnapshotChannel() chan domain.ResourceSnapshot {
	ch := make(chan domain.ResourceSnapshot)
	close(ch)
	return ch
}

func closedPromptSnapshotChannel() chan domain.PromptSnapshot {
	ch := make(chan domain.PromptSnapshot)
	close(ch)
	return ch
}
