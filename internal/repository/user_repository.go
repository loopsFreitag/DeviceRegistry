package repository

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/loopsFreitag/DeviceRegistry/internal/model"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

type UserRepository interface {
	Create(user *model.User) error
	GetByID(id uuid.UUID) (*model.User, error)
	GetByEmail(email string) (*model.User, error)
	Update(user *model.User) error
	Delete(id uuid.UUID) error
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *model.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, email, created_at, updated_at
	`

	err := r.db.QueryRowx(query, user.ID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt).
		StructScan(user)
	if err != nil {
		// Check for unique constraint violation
		if err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"` {
			return ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

func (r *userRepository) GetByID(id uuid.UUID) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, email, password_hash, created_at, updated_at FROM users WHERE id = $1`

	err := r.db.Get(user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *userRepository) GetByEmail(email string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, email, password_hash, created_at, updated_at FROM users WHERE email = $1`

	err := r.db.Get(user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *userRepository) Update(user *model.User) error {
	query := `
		UPDATE users 
		SET email = $2, password_hash = $3, updated_at = $4
		WHERE id = $1
		RETURNING id, email, created_at, updated_at
	`

	err := r.db.QueryRowx(query, user.ID, user.Email, user.PasswordHash, user.UpdatedAt).
		StructScan(user)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrUserNotFound
		}
		return err
	}

	return nil
}

func (r *userRepository) Delete(id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}
