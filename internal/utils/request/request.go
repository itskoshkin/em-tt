package request

import (
	"context"
)

const HeaderName = "X-Request-ID"

type ctxKey struct{}

func WithContext(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

func FromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(ctxKey{}).(string)
	return id, ok
}
