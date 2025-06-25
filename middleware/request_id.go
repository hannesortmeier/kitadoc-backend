package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type contextKey string

const (
	requestIDKey contextKey = "requestID"
)

// RequestIDMiddleware generates a unique request ID and adds it to the request context.
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestID := uuid.New().String()
		ctx := context.WithValue(request.Context(), requestIDKey, requestID)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from the context.
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// GetLoggerWithReqID returns a logrus entry with the request ID field.
func GetLoggerWithReqID(ctx context.Context) *logrus.Entry {
	if requestID := GetRequestID(ctx); requestID != "" {
		return logrus.WithField("request_id", requestID)
	}
	return logrus.NewEntry(logrus.StandardLogger())
}
