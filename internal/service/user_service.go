package service

import (
	"context"
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/iperez/new-expenses-go/internal/domain/models"
	"github.com/iperez/new-expenses-go/pkg/apperror"
)

// UserService exposes the operations related to the users table.
type UserService struct {
	db        *gorm.DB
	validator *validator.Validate
}

// CreateUserInput contains the attributes necessary to create a user.
type CreateUserInput struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Password  string `json:"password" validate:"required,min=6"`
}

// NewUserService builds a UserService backed by the provided database handle.
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db, validator: validator.New()}
}

// Create persists a new user and returns it.
func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*models.User, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, apperror.New(apperror.ServerParamsMissing, formatValidationErrors(err))
	}

	normalizedEmail := strings.ToLower(input.Email)

	exists, err := s.userExists(ctx, normalizedEmail)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, apperror.New(apperror.UserExists, nil)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		UserID:    uuid.NewString(),
		Email:     normalizedEmail,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Password:  string(hashed),
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) userExists(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := s.db.WithContext(ctx).Model(&models.User{}).Where("LOWER(email) = ?", email).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// FindByEmail fetches a user using its email address.
func (s *UserService) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("LOWER(email) = ?", strings.ToLower(email)).Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(apperror.UserNotFound, nil)
		}

		return nil, err
	}

	return &user, nil
}

// GetByID fetches the user that owns the provided identifier.
func (s *UserService) GetByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Take(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(apperror.UserNotFound, nil)
		}

		return nil, err
	}

	return &user, nil
}
