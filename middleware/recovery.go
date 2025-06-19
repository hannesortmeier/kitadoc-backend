package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

// Recovery middleware recovers from panics and logs the stack trace.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(writer, "Internal Server Error", http.StatusInternalServerError)
				GetLoggerWithReqID(request.Context()).WithFields(logrus.Fields{
					"panic": err,
					"stack": string(debug.Stack()),
				}).Error("Recovered from panic")
			}
		}()
		next.ServeHTTP(writer, request)
	})
}