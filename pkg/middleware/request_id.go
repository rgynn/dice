package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/rgynn/dice/pkg/helper"
)

type RequestIDContextKey struct{}

func RequestIDContext(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, RequestIDContextKey{}, rid)
}

func RequestIDFromContext(ctx context.Context) (*string, error) {
	rid, ok := ctx.Value(RequestIDContextKey{}).(string)
	if !ok {
		return nil, errors.New("failed to type assert request id (string) from context")
	}
	return &rid, nil
}

func RequestIDMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqid := r.Header.Get("X-Request-ID")
		if reqid == "" {
			reqid = helper.RandomString(20)
		}
		h.ServeHTTP(w, r.WithContext(RequestIDContext(r.Context(), reqid)))
	})
}
