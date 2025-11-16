package repository

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	db := sqlx.NewDb(mockDB, "postgres")
	return db, mock
}

func TestUserRepository_Create(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)

	user := &model.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	t.Run("successful creation", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "created_at", "updated_at"}).
			AddRow(user.ID, user.Email, user.CreatedAt, user.UpdatedAt)

		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(user.ID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt).
			WillReturnRows(rows)

		err := repo.Create(user)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("duplicate email error", func(t *testing.T) {
		mock.ExpectQuery(`INSERT INTO users`).
			WithArgs(user.ID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt).
			WillReturnError(ErrUserAlreadyExists)

		err := repo.Create(user)
		assert.Error(t, err)
		assert.Equal(t, ErrUserAlreadyExists, err)
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	userID := uuid.New()

	t.Run("user found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, "test@example.com", "hashedpassword", time.Now(), time.Now())

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE id`).
			WithArgs(userID).
			WillReturnRows(rows)

		user, err := repo.GetByID(userID)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, "test@example.com", user.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM users WHERE id`).
			WithArgs(userID).
			WillReturnError(ErrUserNotFound)

		user, err := repo.GetByID(userID)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrUserNotFound, err)
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, mock := setupMockDB(t)
	defer db.Close()

	repo := NewUserRepository(db)
	email := "test@example.com"
	userID := uuid.New()

	t.Run("user found", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at", "updated_at"}).
			AddRow(userID, email, "hashedpassword", time.Now(), time.Now())

		mock.ExpectQuery(`SELECT (.+) FROM users WHERE email`).
			WithArgs(email).
			WillReturnRows(rows)

		user, err := repo.GetByEmail(email)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT (.+) FROM users WHERE email`).
			WithArgs(email).
			WillReturnError(ErrUserNotFound)

		user, err := repo.GetByEmail(email)
		assert.Error(t, err)
		assert.Nil(t, user)
	})
}
