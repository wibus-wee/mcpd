package router

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"mcpd/internal/domain"
)

func TestBasicRouter_RouteSuccess(t *testing.T) {
	respPayload := json.RawMessage(`{"ok":true}`)
	sched := &fakeScheduler{
		instance: &domain.Instance{
			ID:   "inst1",
			Conn: &fakeConn{resp: respPayload},
		},
	}
	r := NewBasicRouter(sched)

	resp, err := r.Route(context.Background(), "svc", "rk", json.RawMessage(`{"ping":true}`))
	require.NoError(t, err)
	require.JSONEq(t, string(respPayload), string(resp))
	require.True(t, sched.released)
}

func TestBasicRouter_AcquireError(t *testing.T) {
	sched := &fakeScheduler{acquireErr: errors.New("busy")}
	r := NewBasicRouter(sched)

	_, err := r.Route(context.Background(), "svc", "", json.RawMessage(`{}`))
	require.Error(t, err)
}

func TestBasicRouter_NoConn(t *testing.T) {
	sched := &fakeScheduler{
		instance: &domain.Instance{ID: "x"},
	}
	r := NewBasicRouter(sched)

	_, err := r.Route(context.Background(), "svc", "", json.RawMessage(`{}`))
	require.Error(t, err)
}

type fakeScheduler struct {
	instance   *domain.Instance
	acquireErr error
	released   bool
}

func (f *fakeScheduler) Acquire(ctx context.Context, serverType, routingKey string) (*domain.Instance, error) {
	return f.instance, f.acquireErr
}

func (f *fakeScheduler) Release(ctx context.Context, instance *domain.Instance) error {
	f.released = true
	return nil
}

type fakeConn struct {
	req  json.RawMessage
	resp json.RawMessage
	err  error
}

func (f *fakeConn) Send(ctx context.Context, msg json.RawMessage) error {
	f.req = msg
	return f.err
}

func (f *fakeConn) Recv(ctx context.Context) (json.RawMessage, error) {
	return f.resp, f.err
}

func (f *fakeConn) Close() error { return nil }
