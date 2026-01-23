package models

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserModelInterface interface {
	Insert(name, email, password string) error
	Authenticate(email, password string) (int, error)
	Exists(id int) (bool, error)
	Get(id int) (User, error)
	PasswordUpdate(id int, currentPassword, newPassword string) error
}

type User struct {
	ID             int
	Name           string
	Email          string
	HashedPassword []byte
	Created        time.Time
}

type UserModel struct {
	DB *pgxpool.Pool
}

func (m *UserModel) Insert(name, email, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	stmt := `INSERT INTO users (name, email, hashed_password, created)
	         VALUES ($1, $2, $3, NOW() AT TIME ZONE 'UTC')`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = m.DB.Exec(ctx, stmt, name, email, hashedPassword)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_uc_email" {
			return ErrDuplicateEmail
		}

		return fmt.Errorf("inserting user: %w", err)
	}

	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	stmt := `SELECT id, hashed_password FROM users WHERE email = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInvalidCredentials
		}
		return 0, fmt.Errorf("querying user credentials: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		}
		return 0, fmt.Errorf("comparing password hash: %w", err)
	}

	return id, nil
}

func (m *UserModel) Exists(id int) (bool, error) {
	var exists bool

	stmt := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("checking user existence: %w", err)
	}

	return exists, nil
}

func (m *UserModel) Get(id int) (User, error) {
	var user User

	stmt := `SELECT id, name, email, created FROM users WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRow(ctx, stmt, id).
		Scan(&user.ID, &user.Name, &user.Email, &user.Created)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNoRecord
		}
		return User{}, fmt.Errorf("fetching user: %w", err)
	}

	return user, nil
}

func (m *UserModel) PasswordUpdate(id int, currentPassword, newPassword string) error {
	var currentHashedPassword []byte

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stmt := `SELECT hashed_password FROM users WHERE id = $1`
	err := m.DB.QueryRow(ctx, stmt, id).Scan(&currentHashedPassword)
	if err != nil {
		return fmt.Errorf("fetching current password: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword(currentHashedPassword, []byte(currentPassword)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidCredentials
		}
		return fmt.Errorf("validating current password: %w", err)
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return fmt.Errorf("hashing new password: %w", err)
	}

	stmt = `UPDATE users SET hashed_password = $1 WHERE id = $2`

	_, err = m.DB.Exec(ctx, stmt, newHashedPassword, id)
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	return nil
}
