package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"yt-off/backend/internal/models"
)

var ErrDownloadNotFound = errors.New("download not found")

type DownloadRepository struct {
	db *sql.DB
}

func NewDownloadRepository(db *sql.DB) *DownloadRepository {
	return &DownloadRepository{db: db}
}

func (repository *DownloadRepository) Create(task *models.DownloadTask) error {
	_, err := repository.db.Exec(
		`INSERT INTO downloads (
			id, user_id, url, format_id, status, progress, speed, eta, filename, file_size, extension, container_id, error, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		task.ID,
		downloadUserID(task.UserID),
		task.URL,
		task.FormatID,
		task.Status,
		task.Progress,
		nullableString(task.Speed),
		nullableString(task.ETA),
		nullableString(task.FileName),
		task.FileSize,
		nullableString(task.Extension),
		nullableString(task.ContainerID),
		nullableString(task.Error),
		formatTime(task.CreatedAt),
		formatTime(task.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("create download: %w", err)
	}

	return nil
}

func (repository *DownloadRepository) FindByID(id string) (*models.DownloadTask, error) {
	row := repository.db.QueryRow(
		`SELECT d.id, d.user_id, u.username, d.url, d.format_id, d.status, d.progress, d.speed, d.eta, d.filename, d.file_size, d.extension, d.container_id, d.error, d.created_at, d.updated_at
		FROM downloads d
		LEFT JOIN users u ON u.id = d.user_id
		WHERE d.id = ?`,
		id,
	)

	task, err := scanDownload(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDownloadNotFound
		}

		return nil, fmt.Errorf("find download by id: %w", err)
	}

	return task, nil
}

func (repository *DownloadRepository) FindAll() ([]models.DownloadTask, error) {
	rows, err := repository.db.Query(
		`SELECT d.id, d.user_id, u.username, d.url, d.format_id, d.status, d.progress, d.speed, d.eta, d.filename, d.file_size, d.extension, d.container_id, d.error, d.created_at, d.updated_at
		FROM downloads d
		LEFT JOIN users u ON u.id = d.user_id
		ORDER BY d.created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list downloads: %w", err)
	}
	defer rows.Close()

	return scanDownloads(rows)
}

func (repository *DownloadRepository) FindByUserID(userID string) ([]models.DownloadTask, error) {
	rows, err := repository.db.Query(
		`SELECT d.id, d.user_id, u.username, d.url, d.format_id, d.status, d.progress, d.speed, d.eta, d.filename, d.file_size, d.extension, d.container_id, d.error, d.created_at, d.updated_at
		FROM downloads d
		LEFT JOIN users u ON u.id = d.user_id
		WHERE d.user_id = ?
		ORDER BY d.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list downloads: %w", err)
	}
	defer rows.Close()

	return scanDownloads(rows)
}

func scanDownloads(rows *sql.Rows) ([]models.DownloadTask, error) {
	downloads := make([]models.DownloadTask, 0)
	for rows.Next() {
		task, err := scanDownload(rows)
		if err != nil {
			return nil, fmt.Errorf("scan download: %w", err)
		}

		downloads = append(downloads, *task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate downloads: %w", err)
	}

	return downloads, nil
}

func (repository *DownloadRepository) Update(task *models.DownloadTask) error {
	result, err := repository.db.Exec(
		`UPDATE downloads
		SET user_id = ?,
			url = ?,
			format_id = ?,
			status = ?,
			progress = ?,
			speed = ?,
			eta = ?,
			filename = ?,
			file_size = ?,
			extension = ?,
			container_id = ?,
			error = ?,
			created_at = ?,
			updated_at = ?
		WHERE id = ?`,
		downloadUserID(task.UserID),
		task.URL,
		task.FormatID,
		task.Status,
		task.Progress,
		nullableString(task.Speed),
		nullableString(task.ETA),
		nullableString(task.FileName),
		task.FileSize,
		nullableString(task.Extension),
		nullableString(task.ContainerID),
		nullableString(task.Error),
		formatTime(task.CreatedAt),
		formatTime(task.UpdatedAt),
		task.ID,
	)
	if err != nil {
		return fmt.Errorf("update download: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read update result: %w", err)
	}
	if rowsAffected == 0 {
		return ErrDownloadNotFound
	}

	return nil
}

func (repository *DownloadRepository) MarkFileRemoved(fileName string) error {
	_, err := repository.db.Exec(
		`UPDATE downloads
		SET file_size = 0,
			updated_at = ?
		WHERE filename = ?`,
		formatTime(time.Now().UTC()),
		fileName,
	)
	if err != nil {
		return fmt.Errorf("mark download file removed: %w", err)
	}

	return nil
}

type downloadScanner interface {
	Scan(dest ...any) error
}

func scanDownload(scanner downloadScanner) (*models.DownloadTask, error) {
	var task models.DownloadTask
	var userID sql.NullString
	var ownerUsername sql.NullString
	var speed sql.NullString
	var eta sql.NullString
	var fileName sql.NullString
	var extension sql.NullString
	var containerID sql.NullString
	var downloadError sql.NullString
	var createdAt string
	var updatedAt string

	if err := scanner.Scan(
		&task.ID,
		&userID,
		&ownerUsername,
		&task.URL,
		&task.FormatID,
		&task.Status,
		&task.Progress,
		&speed,
		&eta,
		&fileName,
		&task.FileSize,
		&extension,
		&containerID,
		&downloadError,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, err
	}

	task.UserID = userID.String
	task.OwnerUsername = ownerUsername.String
	task.Speed = speed.String
	task.ETA = eta.String
	task.FileName = fileName.String
	task.Extension = extension.String
	task.ContainerID = containerID.String
	task.Error = downloadError.String

	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	parsedUpdatedAt, err := parseTime(updatedAt)
	if err != nil {
		return nil, err
	}
	task.CreatedAt = parsedCreatedAt
	task.UpdatedAt = parsedUpdatedAt

	return &task, nil
}

func downloadUserID(userID string) string {
	if userID == "" {
		return models.DefaultUserID
	}

	return userID
}

func nullableString(value string) sql.NullString {
	return sql.NullString{
		String: value,
		Valid:  value != "",
	}
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse download timestamp: %w", err)
	}

	return parsed, nil
}
