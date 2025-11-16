package service

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
	"github.com/loopsFreitag/DeviceRegistry/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = repository.ErrUserAlreadyExists
	ErrUserNotFound       = repository.ErrUserNotFound
)

type AuthService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (s *AuthService) CreateUser(email, password string) (*model.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(hashedPassword),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(email, password string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if err == repository.ErrUserNotFound {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func (s *AuthService) GetUserByID(userID uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(userID)
}

func (s *AuthService) UpdateUser(user *model.User) error {
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(user)
}

func (s *AuthService) DeleteUser(userID uuid.UUID) error {
	return s.userRepo.Delete(userID)
}
