package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

// RequestLogger logs incoming HTTP requests and their responses.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		
		// Get logger with request ID from context
		logger := GetLoggerWithReqID(request.Context()).WithFields(logrus.Fields{
			"method": request.Method,
			"uri":    request.RequestURI,
			"proto":  request.Proto,
		})

		logger.Info("Incoming request")

		next.ServeHTTP(writer, request)

		logger.WithField("duration", time.Since(start)).Info("Request completed")
	})
}