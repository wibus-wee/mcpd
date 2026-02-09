package core

import "context"

// BootstrapWaiter waits for bootstrap completion using the provided context.
type BootstrapWaiter func(ctx context.Context) error
