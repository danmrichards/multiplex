package host

import (
	"context"
)

// hostKey is the key for host values in Contexts. It is unexported; clients
// use host.NewContext and host.FromContext instead of using this key directly.
const hostKey key = "host"

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key string

// NewContext returns a new Context that carries value h.
func NewContext(ctx context.Context, h string) context.Context {
	return context.WithValue(ctx, hostKey, h)
}

// FromContext returns the host value stored in ctx, if any.
func FromContext(ctx context.Context) (string, bool) {
	u, ok := ctx.Value(hostKey).(string)
	return u, ok
}
