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
	ErrUsernameRequired = errors.New("username is required")
	ErrUserNotFound     = errors.New("user not found")
)

type UserService struct {
	repository userRepository
}

type userRepository interface {
	Create(user *models.User) error
	FindByID(id string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	FindAll() ([]models.User, error)
}

func NewUserService(repository userRepository) *UserService {
	return &UserService{repository: repository}
}

func (service *UserService) ListUsers() ([]models.User, error) {
	return service.repository.FindAll()
}

func (service *UserService) GetOrCreateUser(username string) (*models.User, error) {
	username = normalizeUsername(username)
	if username == "" {
		return nil, ErrUsernameRequired
	}

	existing, err := service.repository.FindByUsername(username)
	if err == nil {
		return existing, nil
	}
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return nil, err
	}

	now := time.Now().UTC()
	user := &models.User{
		ID:        uuid.NewString(),
		Username:  username,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := service.repository.Create(user); err != nil {
		found, findErr := service.repository.FindByUsername(username)
		if findErr == nil {
			return found, nil
		}

		return nil, err
	}

	return user, nil
}

func (service *UserService) ResolveUserID(userID string) (string, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		userID = models.DefaultUserID
	}

	if _, err := service.repository.FindByID(userID); err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return "", ErrUserNotFound
		}

		return "", err
	}

	return userID, nil
}

func normalizeUsername(username string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(username)), " ")
}
