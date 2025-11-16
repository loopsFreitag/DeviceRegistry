// internal/middleware/auth_middleware_test.go
package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of AuthServiceInterface
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GetUserByID(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthService) CreateUser(email, password string) (*model.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthService) Register(email, password string) (*model.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthService) Login(email, password string) (*model.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

// Test handler to verify middleware functionality
func testHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r.Context())
		if user != nil {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "authenticated",
				"email":   user.Email,
			})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	t.Run("successful authentication", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		middleware := NewAuthMiddleware(mockAuthService)

		// Create a test user
		userID := uuid.New()
		user := &model.User{
			ID:           userID,
			Email:        "test@example.com",
			PasswordHash: "hashedpassword",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Create a session using the SessionStore
		sessionStore := model.GetSessionStore()
		sessionID := sessionStore.Create(userID, user.Email, 24*time.Hour)

		// Mock the GetUserByID call
		mockAuthService.On("GetUserByID", userID).Return(user, nil)

		// Create request with session cookie
		req := httptest.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: sessionID,
		})

		// Create response recorder
		w := httptest.NewRecorder()

		// Create handler with middleware
		handler := middleware.RequireAuth(testHandler())
		handler.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "authenticated", response["message"])
		assert.Equal(t, "test@example.com", response["email"])

		mockAuthService.AssertExpectations(t)

		// Clean up
		sessionStore.Delete(sessionID)
	})

	t.Run("no session cookie", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		middleware := NewAuthMiddleware(mockAuthService)

		// Create request without session cookie
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()

		// Create handler with middleware
		handler := middleware.RequireAuth(testHandler())
		handler.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response["status"])
		assert.Equal(t, "Unauthorized", response["message"])
	})

	t.Run("invalid session cookie", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		middleware := NewAuthMiddleware(mockAuthService)

		// Create request with invalid session cookie
		req := httptest.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: "invalid-session-id",
		})
		w := httptest.NewRecorder()

		// Create handler with middleware
		handler := middleware.RequireAuth(testHandler())
		handler.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response["status"])
		assert.Equal(t, "Invalid or expired session", response["message"])
	})

	t.Run("expired session", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		middleware := NewAuthMiddleware(mockAuthService)

		userID := uuid.New()

		// Create an expired session (negative duration)
		sessionStore := model.GetSessionStore()
		sessionID := sessionStore.Create(userID, "test@example.com", -1*time.Hour)

		// Create request with expired session cookie
		req := httptest.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: sessionID,
		})
		w := httptest.NewRecorder()

		// Create handler with middleware
		handler := middleware.RequireAuth(testHandler())
		handler.ServeHTTP(w, req)

		// Assertions - expired sessions should be treated as invalid
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response["status"])

		// Clean up
		sessionStore.Delete(sessionID)
	})

	t.Run("user not found in database", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		middleware := NewAuthMiddleware(mockAuthService)

		userID := uuid.New()

		// Create a valid session
		sessionStore := model.GetSessionStore()
		sessionID := sessionStore.Create(userID, "test@example.com", 24*time.Hour)

		// Mock GetUserByID to return error (user not found)
		mockAuthService.On("GetUserByID", userID).Return(nil, errors.New("user not found"))

		// Create request with session cookie
		req := httptest.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: sessionID,
		})
		w := httptest.NewRecorder()

		// Create handler with middleware
		handler := middleware.RequireAuth(testHandler())
		handler.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response["status"])
		assert.Equal(t, "User not found", response["message"])

		mockAuthService.AssertExpectations(t)

		// Clean up
		sessionStore.Delete(sessionID)
	})

	t.Run("user context is properly set", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		middleware := NewAuthMiddleware(mockAuthService)

		userID := uuid.New()
		user := &model.User{
			ID:           userID,
			Email:        "context@example.com",
			PasswordHash: "hashedpassword",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// Create a session
		sessionStore := model.GetSessionStore()
		sessionID := sessionStore.Create(userID, user.Email, 24*time.Hour)

		mockAuthService.On("GetUserByID", userID).Return(user, nil)

		// Create request
		req := httptest.NewRequest("GET", "/protected", nil)
		req.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: sessionID,
		})
		w := httptest.NewRecorder()

		// Custom handler to verify context
		verifyContextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contextUser := GetUserFromContext(r.Context())
			assert.NotNil(t, contextUser)
			assert.Equal(t, userID, contextUser.ID)
			assert.Equal(t, "context@example.com", contextUser.Email)
			w.WriteHeader(http.StatusOK)
		})

		handler := middleware.RequireAuth(verifyContextHandler)
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockAuthService.AssertExpectations(t)

		// Clean up
		sessionStore.Delete(sessionID)
	})
}

func TestGetUserFromContext(t *testing.T) {
	t.Run("user exists in context", func(t *testing.T) {
		userID := uuid.New()
		user := &model.User{
			ID:           userID,
			Email:        "test@example.com",
			PasswordHash: "hashedpassword",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		req := httptest.NewRequest("GET", "/test", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, UserContextKey, user)

		retrievedUser := GetUserFromContext(ctx)

		assert.NotNil(t, retrievedUser)
		assert.Equal(t, userID, retrievedUser.ID)
		assert.Equal(t, "test@example.com", retrievedUser.Email)
	})

	t.Run("user does not exist in context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := req.Context()

		retrievedUser := GetUserFromContext(ctx)

		assert.Nil(t, retrievedUser)
	})

	t.Run("wrong type in context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		ctx := req.Context()
		ctx = context.WithValue(ctx, UserContextKey, "not a user")

		retrievedUser := GetUserFromContext(ctx)

		assert.Nil(t, retrievedUser)
	})
}

func TestRespondWithError(t *testing.T) {
	t.Run("formats error response correctly", func(t *testing.T) {
		w := httptest.NewRecorder()

		respondWithError(w, http.StatusBadRequest, "Test error message")

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "error", response["status"])
		assert.Equal(t, "Test error message", response["message"])
	})

	t.Run("handles different status codes", func(t *testing.T) {
		testCases := []struct {
			statusCode int
			message    string
		}{
			{http.StatusUnauthorized, "Unauthorized"},
			{http.StatusForbidden, "Forbidden"},
			{http.StatusInternalServerError, "Internal Server Error"},
		}

		for _, tc := range testCases {
			w := httptest.NewRecorder()
			respondWithError(w, tc.statusCode, tc.message)

			assert.Equal(t, tc.statusCode, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			assert.Equal(t, "error", response["status"])
			assert.Equal(t, tc.message, response["message"])
		}
	})
}
