package middleware

import (
	"context"
	"net/http"

	"github.com/loopsFreitag/DeviceRegistry/internal/model"
	"github.com/loopsFreitag/DeviceRegistry/internal/service"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthMiddleware struct {
	authService *service.AuthService
	sessions    *model.SessionStore
}

func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		sessions:    model.GetSessionStore(),
	}
}

func (am *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		session, exists := am.sessions.Get(cookie.Value)
		if !exists {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired session")
			return
		}

		// Get user from database
		user, err := am.authService.GetUserByID(session.UserID)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "User not found")
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Helper function to get user from context
func GetUserFromContext(ctx context.Context) *model.User {
	user, ok := ctx.Value(UserContextKey).(*model.User)
	if !ok {
		return nil
	}
	return user
}

// Local helper for errors
func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(`{"status":"error","message":"` + message + `"}`))
}
