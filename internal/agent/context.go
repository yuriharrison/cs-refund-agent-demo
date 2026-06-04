package agent

import "context"

type contextKey int

const (
	contextKeySessionID contextKey = iota
	contextKeyForceError
)

func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, contextKeySessionID, sessionID)
}

func SessionIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(contextKeySessionID).(string); ok {
		return v
	}
	return ""
}

func WithForceError(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyForceError, true)
}

func ForceErrorFromContext(ctx context.Context) bool {
	if v, ok := ctx.Value(contextKeyForceError).(bool); ok {
		return v
	}
	return false
}
