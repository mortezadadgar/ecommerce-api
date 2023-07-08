package http

import (
	"context"
)

type contextKey int

const userContextKey = contextKey(iota)

func newUserContext(ctx context.Context, userID int) context.Context {
	return context.WithValue(ctx, userContextKey, userID)
}

func userIDFromContext(ctx context.Context) int {
	user, _ := ctx.Value(userContextKey).(int)
	return user
}
