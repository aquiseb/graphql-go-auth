package authCommon

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

type key int

// Anti-collision keys for context
const (
	TokenKey   key = iota
	StateKey   key = iota
	AuthURLKey key = iota
)

// StateToContext adds the state to ctx
func StateToContext(ctx context.Context, state string) context.Context {
	return context.WithValue(ctx, StateKey, state)
}

// AuthURLToContext adds authURL to ctx
func AuthURLToContext(ctx context.Context, authURL string) context.Context {
	return context.WithValue(ctx, AuthURLKey, authURL)
}

// TokenToContext adds authURL to ctx
func TokenToContext(ctx context.Context, token *oauth2.Token) context.Context {
	return context.WithValue(ctx, TokenKey, token)
}

// StateFromContext returns the state value from ctx
func StateFromContext(ctx context.Context) (string, error) {
	state, ok := ctx.Value(StateKey).(string)
	if !ok {
		return "", fmt.Errorf("oauth2: Context missing state value")
	}
	return state, nil
}

// StateAndCodeFromReq returns state and code from req
func StateAndCodeFromReq(req *http.Request) (authCode, state string, err error) {
	err = req.ParseForm()
	if err != nil {
		return "", "", err
	}
	authCode = req.Form.Get("code")
	state = req.Form.Get("state")

	if authCode == "" || state == "" {
		return "", "", errors.New("Oauth2: Request missing code or state")
	}
	return authCode, state, nil
}

// TokenFromContext returns the Token from ctx
func TokenFromContext(ctx context.Context) (*oauth2.Token, error) {
	token, ok := ctx.Value(TokenKey).(*oauth2.Token)
	if !ok {
		return nil, fmt.Errorf("oauth2: Context missing Token")
	}
	return token, nil
}
