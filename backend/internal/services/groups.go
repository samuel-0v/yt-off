package services

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"yt-off/backend/internal/models"
	"yt-off/backend/internal/repositories"
)

var (
	ErrDownloadGroupNameRequired = errors.New("group name is required")
	ErrDownloadGroupNotFound     = errors.New("download group not found")
	ErrDownloadGroupItemNotFound = errors.New("download group item not found")
	ErrDownloadGroupForbidden    = errors.New("download group belongs to another user")
)

type DownloadGroupService struct {
	repository downloadGroupRepository
	downloads  downloadLookupRepository
	users      groupUserResolver
}

type downloadGroupRepository interface {
	Create(group *models.DownloadGroup) error
	Update(group *models.DownloadGroup) error
	Delete(id string) error
	FindByID(id string) (*models.DownloadGroup, error)
	FindAll() ([]models.DownloadGroup, error)
	FindByUserID(userID string) ([]models.DownloadGroup, error)
	AddItem(item *models.DownloadGroupItem) (*models.DownloadGroupItem, error)
	RemoveItem(groupID string, itemID string) error
}

type downloadLookupRepository interface {
	FindByID(id string) (*models.DownloadTask, error)
}

type groupUserResolver interface {
	ResolveUserID(userID string) (string, error)
}

func NewDownloadGroupService(repository downloadGroupRepository, downloads downloadLookupRepository, users groupUserResolver) *DownloadGroupService {
	return &DownloadGroupService{
		repository: repository,
		downloads:  downloads,
		users:      users,
	}
}

func (service *DownloadGroupService) ListGroups(scope string, userID string) ([]models.DownloadGroup, error) {
	if isMineScope(scope) {
		resolvedUserID, err := service.users.ResolveUserID(userID)
		if err != nil {
			return nil, err
		}

		return service.repository.FindByUserID(resolvedUserID)
	}

	return service.repository.FindAll()
}

func (service *DownloadGroupService) GetGroup(id string) (*models.DownloadGroup, error) {
	group, err := service.repository.FindByID(strings.TrimSpace(id))
	if err != nil {
		return nil, downloadGroupError(err)
	}

	service.enrichGroupDownloads(group)
	return group, nil
}

func (service *DownloadGroupService) CreateGroup(userID string, name string, description string) (*models.DownloadGroup, error) {
	resolvedUserID, err := service.users.ResolveUserID(userID)
	if err != nil {
		return nil, err
	}

	name = normalizeGroupName(name)
	if name == "" {
		return nil, ErrDownloadGroupNameRequired
	}

	now := time.Now().UTC()
	group := &models.DownloadGroup{
		ID:          uuid.NewString(),
		UserID:      resolvedUserID,
		Name:        name,
		Description: strings.TrimSpace(description),
		Items:       make([]models.DownloadGroupItem, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := service.repository.Create(group); err != nil {
		return nil, err
	}

	return service.GetGroup(group.ID)
}

func (service *DownloadGroupService) UpdateGroup(id string, actorUserID string, name string, description string) (*models.DownloadGroup, error) {
	group, err := service.GetGroup(id)
	if err != nil {
		return nil, err
	}
	if err := service.authorizeGroupOwner(group, actorUserID); err != nil {
		return nil, err
	}

	name = normalizeGroupName(name)
	if name == "" {
		return nil, ErrDownloadGroupNameRequired
	}

	group.Name = name
	group.Description = strings.TrimSpace(description)
	group.UpdatedAt = time.Now().UTC()
	if err := service.repository.Update(group); err != nil {
		return nil, downloadGroupError(err)
	}

	return service.GetGroup(group.ID)
}

func (service *DownloadGroupService) DeleteGroup(id string, actorUserID string) error {
	group, err := service.GetGroup(id)
	if err != nil {
		return err
	}
	if err := service.authorizeGroupOwner(group, actorUserID); err != nil {
		return err
	}

	if err := service.repository.Delete(group.ID); err != nil {
		return downloadGroupError(err)
	}

	return nil
}

func (service *DownloadGroupService) AddItem(groupID string, actorUserID string, downloadID string) (*models.DownloadGroupItem, error) {
	group, err := service.GetGroup(groupID)
	if err != nil {
		return nil, err
	}
	if err := service.authorizeGroupOwner(group, actorUserID); err != nil {
		return nil, err
	}

	download, err := service.downloads.FindByID(strings.TrimSpace(downloadID))
	if err != nil {
		if errors.Is(err, repositories.ErrDownloadNotFound) {
			return nil, ErrDownloadNotFound
		}

		return nil, err
	}

	now := time.Now().UTC()
	item, err := service.repository.AddItem(&models.DownloadGroupItem{
		ID:         uuid.NewString(),
		GroupID:    group.ID,
		DownloadID: download.ID,
		CreatedAt:  now,
	})
	if err != nil {
		return nil, err
	}
	item.Download = cloneDownloadTask(download)

	group.UpdatedAt = now
	if err := service.repository.Update(group); err != nil {
		return nil, downloadGroupError(err)
	}

	return item, nil
}

func (service *DownloadGroupService) RemoveItem(groupID string, itemID string, actorUserID string) error {
	group, err := service.GetGroup(groupID)
	if err != nil {
		return err
	}
	if err := service.authorizeGroupOwner(group, actorUserID); err != nil {
		return err
	}

	if err := service.repository.RemoveItem(group.ID, strings.TrimSpace(itemID)); err != nil {
		return downloadGroupError(err)
	}

	group.UpdatedAt = time.Now().UTC()
	if err := service.repository.Update(group); err != nil {
		return downloadGroupError(err)
	}

	return nil
}

func (service *DownloadGroupService) enrichGroupDownloads(group *models.DownloadGroup) {
	for index := range group.Items {
		download, err := service.downloads.FindByID(group.Items[index].DownloadID)
		if err != nil {
			continue
		}

		group.Items[index].Download = cloneDownloadTask(download)
	}
}

func (service *DownloadGroupService) authorizeGroupOwner(group *models.DownloadGroup, actorUserID string) error {
	resolvedUserID, err := service.users.ResolveUserID(actorUserID)
	if err != nil {
		return err
	}
	if group.UserID != resolvedUserID {
		return ErrDownloadGroupForbidden
	}

	return nil
}

func downloadGroupError(err error) error {
	switch {
	case errors.Is(err, repositories.ErrDownloadGroupNotFound):
		return ErrDownloadGroupNotFound
	case errors.Is(err, repositories.ErrDownloadGroupItemNotFound):
		return ErrDownloadGroupItemNotFound
	default:
		return err
	}
}

func isMineScope(scope string) bool {
	return strings.EqualFold(strings.TrimSpace(scope), "mine")
}

func normalizeGroupName(name string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(name)), " ")
}
