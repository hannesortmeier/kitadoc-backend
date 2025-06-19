package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"kitadoc-backend/config"
	"kitadoc-backend/data"
	"kitadoc-backend/models"
	"kitadoc-backend/services"

	"github.com/golang-jwt/jwt/v5"
)

type contextKeyUser string

const (
	ContextKeyUser contextKeyUser = "user"
)

// Claims defines the structure of our JWT claims.
type Claims struct {
	UserID int      `json:"user_id"`
	Role   data.Role `json:"role"`
	jwt.RegisteredClaims
}

// Authenticate middleware validates JWT tokens and injects user context.
func Authenticate(userService services.UserService, cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			logger := GetLoggerWithReqID(request.Context())
			authHeader := request.Header.Get("Authorization")
			if authHeader == "" {
				logger.Warn("Unauthorized: Missing Authorization header")
				http.Error(writer, "Unauthorized", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				logger.Warn("Unauthorized: Invalid Authorization header format")
				http.Error(writer, "Invalid Authorization header format", http.StatusUnauthorized)
				return
			}

			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					logger.WithField("signing_method", token.Method).Warn("Unexpected signing method for JWT")
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(cfg.Server.JWTSecret), nil
			})

			if err != nil || !token.Valid {
				logger.WithError(err).Warn("Invalid or expired token")
				http.Error(writer, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			// Fetch user from database to ensure they still exist and are active
			user, err := userService.GetUserByID(logger, request.Context(), claims.UserID)
			if err != nil {
				logger.WithError(err).WithField("user_id", claims.UserID).Warn("User not found or inactive during authentication")
				http.Error(writer, "User not found or inactive", http.StatusUnauthorized)
				return
			}

			// Inject user into context
			ctx := context.WithValue(request.Context(), ContextKeyUser, user)
			next.ServeHTTP(writer, request.WithContext(ctx))
		})
	}
}

// Authorize middleware checks if the authenticated user has the required role.
func Authorize(requiredRole data.Role) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			user, ok := request.Context().Value(ContextKeyUser).(*models.User)
			if !ok {
				GetLoggerWithReqID(request.Context()).Error("Forbidden: User context not found in Authorize middleware")
				http.Error(writer, "Forbidden: User context not found", http.StatusForbidden)
				return
			}

			if user.Role != string(requiredRole) && user.Role != string(data.RoleAdmin) { // Admin can do anything
				http.Error(writer, "Forbidden: Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(writer, request)
		})
	}
}