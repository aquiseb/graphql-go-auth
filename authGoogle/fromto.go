package authGoogle

import (
	"context"
	"fmt"

	google "google.golang.org/api/oauth2/v2"
)

type key int

const (
	UserKey key = iota
)

func UserToContext(ctx context.Context, user *google.Userinfoplus) context.Context {
	return context.WithValue(ctx, UserKey, user)
}

// UserFromContext returns the user from ctx
func UserFromContext(ctx context.Context) (*google.Userinfoplus, error) {
	user, ok := ctx.Value(UserKey).(*google.Userinfoplus)
	if !ok {
		return nil, fmt.Errorf("google: Context missing Google User")
	}
	return user, nil
}
