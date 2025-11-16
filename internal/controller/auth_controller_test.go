package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
	"github.com/loopsFreitag/DeviceRegistry/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService now implements service.AuthServiceInterface
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) CreateUser(email, password string) (*model.User, error) {
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

func (m *MockAuthService) GetUserByID(userID uuid.UUID) (*model.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockAuthService) UpdateUser(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockAuthService) DeleteUser(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func TestAuthController_Register(t *testing.T) {
	mockService := new(MockAuthService)
	controller := NewAuthController(mockService)

	t.Run("successful registration", func(t *testing.T) {
		reqBody := RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		expectedUser := &model.User{
			ID:        uuid.New(),
			Email:     reqBody.Email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("CreateUser", reqBody.Email, reqBody.Password).
			Return(expectedUser, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		controller.Register(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, reqBody.Email, response.User.Email)
		assert.Equal(t, "User created successfully", response.Message)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		controller.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "error", response.Status)
		assert.Equal(t, "Invalid request body", response.Message)
	})

	t.Run("missing email", func(t *testing.T) {
		reqBody := RegisterRequest{
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		controller.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Email and password are required", response.Message)
	})

	t.Run("missing password", func(t *testing.T) {
		reqBody := RegisterRequest{
			Email: "test@example.com",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		controller.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Email and password are required", response.Message)
	})

	t.Run("user already exists", func(t *testing.T) {
		reqBody := RegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		mockService.On("CreateUser", reqBody.Email, reqBody.Password).
			Return(nil, service.ErrUserAlreadyExists).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		controller.Register(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "User already exists", response.Message)

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		reqBody := RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		mockService.On("CreateUser", reqBody.Email, reqBody.Password).
			Return(nil, assert.AnError).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		controller.Register(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to create user", response.Message)

		mockService.AssertExpectations(t)
	})
}

func TestAuthController_Login(t *testing.T) {
	mockService := new(MockAuthService)
	controller := NewAuthController(mockService)

	t.Run("successful login", func(t *testing.T) {
		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		expectedUser := &model.User{
			ID:    uuid.New(),
			Email: reqBody.Email,
		}

		mockService.On("Login", reqBody.Email, reqBody.Password).
			Return(expectedUser, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		controller.Login(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Check cookie was set
		cookies := w.Result().Cookies()
		assert.NotEmpty(t, cookies)

		var foundSessionCookie bool
		for _, cookie := range cookies {
			if cookie.Name == SessionCookieName {
				foundSessionCookie = true
				assert.NotEmpty(t, cookie.Value)
				assert.True(t, cookie.HttpOnly)
				assert.Equal(t, "/", cookie.Path)
				assert.False(t, cookie.Secure) // As per your config
			}
		}
		assert.True(t, foundSessionCookie, "Session cookie should be set")

		var response AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, reqBody.Email, response.User.Email)
		assert.Equal(t, "Login successful", response.Message)

		mockService.AssertExpectations(t)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		controller.Login(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid request body", response.Message)
	})

	t.Run("invalid credentials", func(t *testing.T) {
		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(reqBody)

		mockService.On("Login", reqBody.Email, reqBody.Password).
			Return(nil, service.ErrInvalidCredentials).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		controller.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid credentials", response.Message)

		mockService.AssertExpectations(t)
	})

	t.Run("service error", func(t *testing.T) {
		reqBody := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(reqBody)

		mockService.On("Login", reqBody.Email, reqBody.Password).
			Return(nil, assert.AnError).Once()

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		controller.Login(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to login", response.Message)

		mockService.AssertExpectations(t)
	})
}
