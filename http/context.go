package http

import (
	"context"

	"github.com/mortezadadgar/ecommerce-api/domain"
)

type contextKey int

const userContextKey = contextKey(iota)

func newUserContext(ctx context.Context, user domain.Users) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func userIDFromContext(ctx context.Context) int {
	user, _ := ctx.Value(userContextKey).(domain.Users)
	return user.ID
}
