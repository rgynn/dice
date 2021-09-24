package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type LoggerContextKey struct{}

func LoggerContext(ctx context.Context, contextlogger *logrus.Entry) context.Context {
	return context.WithValue(ctx, LoggerContextKey{}, contextlogger)
}

func LoggerFromContext(ctx context.Context) (*logrus.Entry, error) {
	contextlogger, ok := ctx.Value(LoggerContextKey{}).(*logrus.Entry)
	if !ok {
		return nil, errors.New("failed to type assert *logrus.Logger from context")
	}
	return contextlogger, nil
}

func ContextLoggerMiddleware(logLevel logrus.Level) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contextlogger := logrus.New().WithFields(logrus.Fields{
				"start":  time.Now().UTC().Format(time.RFC3339),
				"method": r.Method,
				"path":   r.URL.Path,
				"query":  r.URL.Query().Encode(),
			})
			if rid, err := RequestIDFromContext(r.Context()); err == nil {
				contextlogger = contextlogger.WithField("rid", *rid)
			}
			contextlogger.Logger.SetLevel(logLevel)
			h.ServeHTTP(w, r.WithContext(LoggerContext(r.Context(), contextlogger)))
		})
	}
}

func AccessLoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
}
