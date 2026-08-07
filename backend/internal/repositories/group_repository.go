package repositories

import (
	"database/sql"
	"errors"
	"fmt"

	"yt-off/backend/internal/models"
)

var (
	ErrDownloadGroupNotFound     = errors.New("download group not found")
	ErrDownloadGroupItemNotFound = errors.New("download group item not found")
)

type DownloadGroupRepository struct {
	db *sql.DB
}

func NewDownloadGroupRepository(db *sql.DB) *DownloadGroupRepository {
	return &DownloadGroupRepository{db: db}
}

func (repository *DownloadGroupRepository) Create(group *models.DownloadGroup) error {
	_, err := repository.db.Exec(
		`INSERT INTO download_groups (id, user_id, name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		group.ID,
		group.UserID,
		group.Name,
		nullableString(group.Description),
		formatTime(group.CreatedAt),
		formatTime(group.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("create download group: %w", err)
	}

	return nil
}

func (repository *DownloadGroupRepository) Update(group *models.DownloadGroup) error {
	result, err := repository.db.Exec(
		`UPDATE download_groups
		SET name = ?,
			description = ?,
			updated_at = ?
		WHERE id = ?`,
		group.Name,
		nullableString(group.Description),
		formatTime(group.UpdatedAt),
		group.ID,
	)
	if err != nil {
		return fmt.Errorf("update download group: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read download group update result: %w", err)
	}
	if rowsAffected == 0 {
		return ErrDownloadGroupNotFound
	}

	return nil
}

func (repository *DownloadGroupRepository) Delete(id string) error {
	tx, err := repository.db.Begin()
	if err != nil {
		return fmt.Errorf("begin delete download group: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM download_group_items WHERE group_id = ?`, id); err != nil {
		return fmt.Errorf("delete download group items: %w", err)
	}
	result, err := tx.Exec(`DELETE FROM download_groups WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete download group: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read download group delete result: %w", err)
	}
	if rowsAffected == 0 {
		return ErrDownloadGroupNotFound
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit delete download group: %w", err)
	}

	return nil
}

func (repository *DownloadGroupRepository) FindByID(id string) (*models.DownloadGroup, error) {
	row := repository.db.QueryRow(
		`SELECT g.id, g.user_id, u.username, g.name, g.description, COUNT(i.id), g.created_at, g.updated_at
		FROM download_groups g
		LEFT JOIN users u ON u.id = g.user_id
		LEFT JOIN download_group_items i ON i.group_id = g.id
		WHERE g.id = ?
		GROUP BY g.id, g.user_id, u.username, g.name, g.description, g.created_at, g.updated_at`,
		id,
	)

	group, err := scanDownloadGroup(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDownloadGroupNotFound
		}

		return nil, fmt.Errorf("find download group by id: %w", err)
	}

	items, err := repository.FindItems(id)
	if err != nil {
		return nil, err
	}
	group.Items = items
	group.ItemCount = len(items)

	return group, nil
}

func (repository *DownloadGroupRepository) FindAll() ([]models.DownloadGroup, error) {
	return repository.findGroups(
		`SELECT g.id, g.user_id, u.username, g.name, g.description, COUNT(i.id), g.created_at, g.updated_at
		FROM download_groups g
		LEFT JOIN users u ON u.id = g.user_id
		LEFT JOIN download_group_items i ON i.group_id = g.id
		GROUP BY g.id, g.user_id, u.username, g.name, g.description, g.created_at, g.updated_at
		ORDER BY g.updated_at DESC`,
	)
}

func (repository *DownloadGroupRepository) FindByUserID(userID string) ([]models.DownloadGroup, error) {
	return repository.findGroups(
		`SELECT g.id, g.user_id, u.username, g.name, g.description, COUNT(i.id), g.created_at, g.updated_at
		FROM download_groups g
		LEFT JOIN users u ON u.id = g.user_id
		LEFT JOIN download_group_items i ON i.group_id = g.id
		WHERE g.user_id = ?
		GROUP BY g.id, g.user_id, u.username, g.name, g.description, g.created_at, g.updated_at
		ORDER BY g.updated_at DESC`,
		userID,
	)
}

func (repository *DownloadGroupRepository) FindItems(groupID string) ([]models.DownloadGroupItem, error) {
	rows, err := repository.db.Query(
		`SELECT id, group_id, download_id, position, created_at
		FROM download_group_items
		WHERE group_id = ?
		ORDER BY position, created_at`,
		groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("list download group items: %w", err)
	}
	defer rows.Close()

	items := make([]models.DownloadGroupItem, 0)
	for rows.Next() {
		item, err := scanDownloadGroupItem(rows)
		if err != nil {
			return nil, fmt.Errorf("scan download group item: %w", err)
		}

		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate download group items: %w", err)
	}

	return items, nil
}

func (repository *DownloadGroupRepository) AddItem(item *models.DownloadGroupItem) (*models.DownloadGroupItem, error) {
	position, err := repository.nextItemPosition(item.GroupID)
	if err != nil {
		return nil, err
	}
	item.Position = position

	_, err = repository.db.Exec(
		`INSERT OR IGNORE INTO download_group_items (id, group_id, download_id, position, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		item.ID,
		item.GroupID,
		item.DownloadID,
		item.Position,
		formatTime(item.CreatedAt),
	)
	if err != nil {
		return nil, fmt.Errorf("add download group item: %w", err)
	}

	return repository.FindItemByDownloadID(item.GroupID, item.DownloadID)
}

func (repository *DownloadGroupRepository) FindItemByDownloadID(groupID string, downloadID string) (*models.DownloadGroupItem, error) {
	row := repository.db.QueryRow(
		`SELECT id, group_id, download_id, position, created_at
		FROM download_group_items
		WHERE group_id = ? AND download_id = ?`,
		groupID,
		downloadID,
	)

	item, err := scanDownloadGroupItem(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDownloadGroupItemNotFound
		}

		return nil, fmt.Errorf("find download group item by download: %w", err)
	}

	return item, nil
}

func (repository *DownloadGroupRepository) RemoveItem(groupID string, itemID string) error {
	result, err := repository.db.Exec(
		`DELETE FROM download_group_items
		WHERE group_id = ? AND id = ?`,
		groupID,
		itemID,
	)
	if err != nil {
		return fmt.Errorf("remove download group item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read download group item delete result: %w", err)
	}
	if rowsAffected == 0 {
		return ErrDownloadGroupItemNotFound
	}

	return nil
}

func (repository *DownloadGroupRepository) findGroups(query string, args ...any) ([]models.DownloadGroup, error) {
	rows, err := repository.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list download groups: %w", err)
	}
	defer rows.Close()

	groups := make([]models.DownloadGroup, 0)
	for rows.Next() {
		group, err := scanDownloadGroup(rows)
		if err != nil {
			return nil, fmt.Errorf("scan download group: %w", err)
		}

		groups = append(groups, *group)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate download groups: %w", err)
	}

	return groups, nil
}

func (repository *DownloadGroupRepository) nextItemPosition(groupID string) (int, error) {
	row := repository.db.QueryRow(
		`SELECT COALESCE(MAX(position) + 1, 0)
		FROM download_group_items
		WHERE group_id = ?`,
		groupID,
	)

	var position int
	if err := row.Scan(&position); err != nil {
		return 0, fmt.Errorf("read next download group item position: %w", err)
	}

	return position, nil
}

type downloadGroupScanner interface {
	Scan(dest ...any) error
}

func scanDownloadGroup(scanner downloadGroupScanner) (*models.DownloadGroup, error) {
	var group models.DownloadGroup
	var ownerUsername sql.NullString
	var description sql.NullString
	var createdAt string
	var updatedAt string

	if err := scanner.Scan(
		&group.ID,
		&group.UserID,
		&ownerUsername,
		&group.Name,
		&description,
		&group.ItemCount,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, err
	}

	group.OwnerUsername = ownerUsername.String
	group.Description = description.String

	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	parsedUpdatedAt, err := parseTime(updatedAt)
	if err != nil {
		return nil, err
	}
	group.CreatedAt = parsedCreatedAt
	group.UpdatedAt = parsedUpdatedAt

	return &group, nil
}

func scanDownloadGroupItem(scanner downloadGroupScanner) (*models.DownloadGroupItem, error) {
	var item models.DownloadGroupItem
	var createdAt string

	if err := scanner.Scan(&item.ID, &item.GroupID, &item.DownloadID, &item.Position, &createdAt); err != nil {
		return nil, err
	}

	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return nil, err
	}
	item.CreatedAt = parsedCreatedAt

	return &item, nil
}
