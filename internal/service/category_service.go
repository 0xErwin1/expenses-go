package service

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/iperez/new-expenses-go/internal/domain/models"
	"github.com/iperez/new-expenses-go/pkg/apperror"
)

// CategoryService contains the logic necessary to manage categories.
type CategoryService struct {
	db        *gorm.DB
	validator *validator.Validate
}

// CreateCategoryInput contains the payload required to create a category.
type CreateCategoryInput struct {
	Type models.TransactionType `json:"type" validate:"required,oneof=INCOME EXPENSE SAVING INSTALLMENTS"`
	Name string                 `json:"name" validate:"required"`
	Note string                 `json:"note"`
}

// UpdateCategoryPayload is used when the transaction service needs to create a category on the fly.
type UpdateCategoryPayload struct {
	Type models.TransactionType `json:"type" validate:"omitempty,oneof=INCOME EXPENSE SAVING INSTALLMENTS"`
	Name string                 `json:"name" validate:"required"`
	Note string                 `json:"note"`
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	return &CategoryService{db: db, validator: validator.New()}
}

func (s *CategoryService) Create(ctx context.Context, userID string, input CreateCategoryInput) (*models.Category, error) {
	return s.createWithDB(ctx, nil, userID, input)
}

func (s *CategoryService) GetByID(ctx context.Context, userID, categoryID string) (*models.Category, error) {
	return s.getByIDWithDB(ctx, nil, userID, categoryID)
}

func (s *CategoryService) List(ctx context.Context, userID string) ([]models.Category, error) {
	var categories []models.Category
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *CategoryService) Delete(ctx context.Context, userID, categoryID string, deleteTransactions bool) error {
	tx := s.db.WithContext(ctx).Begin()

	var category models.Category
	if err := tx.Where("category_id = ? AND user_id = ?", categoryID, userID).Take(&category).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.New(apperror.CategoryNotFound, nil)
		}

		return err
	}

	var txCount int64
	if err := tx.Model(&models.Transaction{}).Where("category_id = ? AND user_id = ?", categoryID, userID).Count(&txCount).Error; err != nil {
		tx.Rollback()
		return err
	}

	if txCount > 0 && !deleteTransactions {
		tx.Rollback()
		return apperror.New(apperror.CategoryHasTransactions, nil)
	}

	if deleteTransactions {
		if err := tx.Where("category_id = ? AND user_id = ?", categoryID, userID).Delete(&models.Transaction{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		if err := tx.Model(&models.Transaction{}).
			Where("category_id = ? AND user_id = ?", categoryID, userID).
			Update("category_id", nil).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Delete(&category).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// EnsureAndCreate validates or creates a category while creating a transaction.
func (s *CategoryService) EnsureAndCreate(ctx context.Context, db *gorm.DB, userID string, categoryID *string, payload *UpdateCategoryPayload, transactionType models.TransactionType) (*models.Category, error) {
	if categoryID != nil {
		category, err := s.getByIDWithDB(ctx, db, userID, *categoryID)
		if err != nil {
			return nil, err
		}

		if category.Type != transactionType {
			return nil, apperror.New(apperror.TransactionCategoryTypeMismatch, nil)
		}

		return category, nil
	}

	if payload == nil {
		return nil, apperror.New(apperror.ServerTooFewParams, nil)
	}

	if payload.Type != "" && payload.Type != transactionType {
		return nil, apperror.New(apperror.TransactionCategoryTypeMismatch, nil)
	}

	input := CreateCategoryInput{
		Type: transactionType,
		Name: payload.Name,
		Note: payload.Note,
	}

	if payload.Type != "" {
		input.Type = payload.Type
	}

	return s.createWithDB(ctx, db, userID, input)
}

func (s *CategoryService) createWithDB(ctx context.Context, db *gorm.DB, userID string, input CreateCategoryInput) (*models.Category, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, apperror.New(apperror.ServerParamsMissing, err)
	}

	category := &models.Category{
		CategoryID: uuid.NewString(),
		Type:       input.Type,
		Name:       input.Name,
		Note:       input.Note,
		UserID:     userID,
	}

	exec := s.db
	if db != nil {
		exec = db
	}

	if err := exec.WithContext(ctx).Create(category).Error; err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) getByIDWithDB(ctx context.Context, db *gorm.DB, userID, categoryID string) (*models.Category, error) {
	exec := s.db
	if db != nil {
		exec = db
	}

	var category models.Category
	if err := exec.WithContext(ctx).Where("category_id = ? AND user_id = ?", categoryID, userID).Take(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(apperror.CategoryNotFound, nil)
		}

		return nil, err
	}

	return &category, nil
}
