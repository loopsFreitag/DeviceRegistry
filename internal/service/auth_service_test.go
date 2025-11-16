package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
	"github.com/loopsFreitag/DeviceRegistry/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Mock UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*model.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func TestAuthService_CreateUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo)

	t.Run("successful user creation", func(t *testing.T) {
		email := "test@example.com"
		password := "password123"

		mockRepo.On("Create", mock.AnythingOfType("*model.User")).Return(nil).Once()

		user, err := service.CreateUser(email, password)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.NotEmpty(t, user.ID)
		assert.NotEmpty(t, user.PasswordHash)

		// Verify password was hashed
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		assert.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user already exists", func(t *testing.T) {
		email := "existing@example.com"
		password := "password123"

		mockRepo.On("Create", mock.AnythingOfType("*model.User")).
			Return(repository.ErrUserAlreadyExists).Once()

		user, err := service.CreateUser(email, password)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrUserAlreadyExists, err)

		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo)

	email := "test@example.com"
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	t.Run("successful login", func(t *testing.T) {
		existingUser := &model.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: string(hashedPassword),
		}

		mockRepo.On("GetByEmail", email).Return(existingUser, nil).Once()

		user, err := service.Login(email, password)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		existingUser := &model.User{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: string(hashedPassword),
		}

		mockRepo.On("GetByEmail", email).Return(existingUser, nil).Once()

		user, err := service.Login(email, "wrongpassword")

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrInvalidCredentials, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.On("GetByEmail", email).Return(nil, repository.ErrUserNotFound).Once()

		user, err := service.Login(email, password)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrInvalidCredentials, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.On("GetByEmail", email).Return(nil, errors.New("db error")).Once()

		user, err := service.Login(email, password)

		assert.Error(t, err)
		assert.Nil(t, user)

		mockRepo.AssertExpectations(t)
	})
}

func TestAuthService_GetUserByID(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo)

	userID := uuid.New()

	t.Run("user found", func(t *testing.T) {
		expectedUser := &model.User{
			ID:    userID,
			Email: "test@example.com",
		}

		mockRepo.On("GetByID", userID).Return(expectedUser, nil).Once()

		user, err := service.GetUserByID(userID)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)

		mockRepo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.On("GetByID", userID).Return(nil, repository.ErrUserNotFound).Once()

		user, err := service.GetUserByID(userID)

		assert.Error(t, err)
		assert.Nil(t, user)

		mockRepo.AssertExpectations(t)
	})
}
