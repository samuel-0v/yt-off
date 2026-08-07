package repositories

import (
	"database/sql"
	"errors"
	"fmt"

	"yt-off/backend/internal/models"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (repository *UserRepository) Create(user *models.User) error {
	_, err := repository.db.Exec(
		`INSERT INTO users (id, username, created_at, updated_at)
		VALUES (?, ?, ?, ?)`,
		user.ID,
		user.Username,
		formatTime(user.CreatedAt),
		formatTime(user.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (repository *UserRepository) FindByID(id string) (*models.User, error) {
	row := repository.db.QueryRow(
		`SELECT id, username, created_at, updated_at
		FROM users
		WHERE id = ?`,
		id,
	)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return user, nil
}

func (repository *UserRepository) FindByUsername(username string) (*models.User, error) {
	row := repository.db.QueryRow(
		`SELECT id, username, created_at, updated_at
		FROM users
		WHERE lower(username) = lower(?)`,
		username,
	)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("find user by username: %w", err)
	}

	return user, nil
}

func (repository *UserRepository) FindAll() ([]models.User, error) {
	rows, err := repository.db.Query(
		`SELECT id, username, created_at, updated_at
		FROM users
		ORDER BY username COLLATE NOCASE`,
	)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]models.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}

		users = append(users, *user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}

	return users, nil
}

type userScanner interface {
	Scan(dest ...any) error
}

func scanUser(scanner userScanner) (*models.User, error) {
	var user models.User
	var createdAt string
	var updatedAt string

	if err := scanner.Scan(&user.ID, &user.Username, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	parsedUpdatedAt, err := parseTime(updatedAt)
	if err != nil {
		return nil, err
	}
	user.CreatedAt = parsedCreatedAt
	user.UpdatedAt = parsedUpdatedAt

	return &user, nil
}
