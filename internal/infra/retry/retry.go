package retry

import "context"

func Retry(ctx context.Context, policy Policy, fn func(context.Context) error) error {
	if ctx == nil {
		return fn(context.Background())
	}
	p := policy.normalized()
	if p.MaxRetries == 0 {
		return fn(ctx)
	}
	attempts := 0
	backoff := NewBackoff(p)
	for {
		err := fn(ctx)
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if p.MaxRetries >= 0 && attempts >= p.MaxRetries {
			return err
		}
		attempts++
		if !backoff.Sleep(ctx) {
			return ctx.Err()
		}
	}
}

func Loop(ctx context.Context, policy Policy, fn func(context.Context) error) error {
	if ctx == nil {
		ctx = context.Background()
	}
	p := policy.normalized()
	backoff := NewBackoff(p)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err := fn(ctx)
		if err == nil {
			backoff.Reset()
			continue
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !backoff.Sleep(ctx) {
			return ctx.Err()
		}
	}
}
