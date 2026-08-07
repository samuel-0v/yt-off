package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"yt-off/backend/internal/models"
)

var ErrSettingNotFound = errors.New("setting not found")

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (repository *SettingsRepository) Upsert(key string, value string) error {
	_, err := repository.db.Exec(
		`INSERT INTO settings (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at`,
		key,
		value,
		formatTime(time.Now().UTC()),
	)
	if err != nil {
		return fmt.Errorf("upsert setting: %w", err)
	}

	return nil
}

func (repository *SettingsRepository) UpsertMany(values map[string]string) error {
	tx, err := repository.db.Begin()
	if err != nil {
		return fmt.Errorf("begin settings transaction: %w", err)
	}
	defer tx.Rollback()

	statement, err := tx.Prepare(
		`INSERT INTO settings (key, value, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at`,
	)
	if err != nil {
		return fmt.Errorf("prepare settings upsert: %w", err)
	}
	defer statement.Close()

	now := formatTime(time.Now().UTC())
	for key, value := range values {
		if _, err := statement.Exec(key, value, now); err != nil {
			return fmt.Errorf("upsert setting %q: %w", key, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit settings transaction: %w", err)
	}

	return nil
}

func (repository *SettingsRepository) FindByKey(key string) (*models.Setting, error) {
	row := repository.db.QueryRow(
		`SELECT key, value, updated_at
		FROM settings
		WHERE key = ?`,
		key,
	)

	var setting models.Setting
	var updatedAt string
	if err := row.Scan(&setting.Key, &setting.Value, &updatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSettingNotFound
		}

		return nil, fmt.Errorf("find setting by key: %w", err)
	}

	parsedUpdatedAt, err := parseTime(updatedAt)
	if err != nil {
		return nil, err
	}
	setting.UpdatedAt = parsedUpdatedAt

	return &setting, nil
}

func (repository *SettingsRepository) FindAll() (map[string]string, error) {
	rows, err := repository.db.Query(
		`SELECT key, value
		FROM settings`,
	)
	if err != nil {
		return nil, fmt.Errorf("list settings: %w", err)
	}
	defer rows.Close()

	values := make(map[string]string)
	for rows.Next() {
		var key string
		var value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		values[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate settings: %w", err)
	}

	return values, nil
}
