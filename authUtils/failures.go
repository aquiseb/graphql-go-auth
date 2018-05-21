package authUtils

import (
	"context"
	"fmt"
	"net/http"
)

type key int

const (
	errorKey key = iota
)

// DefaultFailureHandler :
// - Responds with a 400 status code and message parsed from ctx
var DefaultFailureHandler = http.HandlerFunc(failureHandler)

func failureHandler(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	err := ErrorFromContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// should be unreachable, ErrorFromContext always returns some non-nil error
	http.Error(w, "", http.StatusBadRequest)
}

// WithError :
// -  Returns a copy of ctx that stores the given error value
func WithError(ctx context.Context, err error) context.Context {
	return context.WithValue(ctx, errorKey, err)
}

// ErrorFromContext :
// - Returns the error value from the ctx
// - Or an error that the context was missing an error value
func ErrorFromContext(ctx context.Context) error {
	err, ok := ctx.Value(errorKey).(error)
	if !ok {
		return fmt.Errorf("Context missing error value")
	}
	return err
}
