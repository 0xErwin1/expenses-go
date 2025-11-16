package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/iperez/new-expenses-go/internal/domain/models"
	"github.com/iperez/new-expenses-go/pkg/apperror"
	"github.com/iperez/new-expenses-go/pkg/day"
)

// TransactionService contains the business logic for managing transactions.
type TransactionService struct {
	db         *gorm.DB
	categories *CategoryService
}

// CreateTransactionInput is the payload accepted when creating a transaction.
type CreateTransactionInput struct {
	Type         models.TransactionType `json:"type"`
	Amount       float64                `json:"amount"`
	Currency     models.Currency        `json:"currency"`
	Note         string                 `json:"note"`
	Day          *int                   `json:"day"`
	Month        models.Month           `json:"month"`
	Year         int                    `json:"year"`
	ExchangeRate *float64               `json:"exchangeRate"`
	CategoryID   *string                `json:"categoryId"`
	Category     *UpdateCategoryPayload `json:"category"`
}

// TransactionFilters encapsulates the optional parameters supported by list/balance endpoints.
type TransactionFilters struct {
	Type  *models.TransactionType
	Day   *int
	Month *models.Month
	Year  *int
}

// BalanceSummary represents the total amount per currency.
type BalanceSummary struct {
	Total float64 `json:"total"`
	UYU   float64 `json:"uyu"`
	USD   float64 `json:"usd"`
	EUR   float64 `json:"eur"`
}

// TransactionBalances splits the totals per transaction type.
type TransactionBalances struct {
	Expenses BalanceSummary `json:"expenses"`
	Incomes  BalanceSummary `json:"incomes"`
	Savings  BalanceSummary `json:"savings"`
}

func NewTransactionService(db *gorm.DB, categories *CategoryService) *TransactionService {
	return &TransactionService{db: db, categories: categories}
}

func (s *TransactionService) Create(ctx context.Context, userID string, payloads []CreateTransactionInput) ([]models.Transaction, error) {
	if len(payloads) == 0 {
		return nil, apperror.New(apperror.ServerParamsMissing, []string{"transactions are required"})
	}

	created := make([]models.Transaction, 0, len(payloads))

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for index := range payloads {
			if err := s.validateTransaction(&payloads[index], index); err != nil {
				return err
			}

			category, err := s.categories.EnsureAndCreate(ctx, tx, userID, payloads[index].CategoryID, payloads[index].Category, payloads[index].Type)
			if err != nil {
				return err
			}

			transaction := models.Transaction{
				TransactionID: uuid.NewString(),
				Type:          payloads[index].Type,
				Amount:        payloads[index].Amount,
				Currency:      payloads[index].Currency,
				Note:          payloads[index].Note,
				Day:           payloads[index].Day,
				Month:         payloads[index].Month,
				Year:          payloads[index].Year,
				ExchangeRate:  payloads[index].ExchangeRate,
				UserID:        userID,
			}

			if category != nil {
				transaction.CategoryID = &category.CategoryID
				transaction.Category = category
			}

			if err := tx.Create(&transaction).Error; err != nil {
				return err
			}

			created = append(created, transaction)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return created, nil
}

func (s *TransactionService) validateTransaction(payload *CreateTransactionInput, index int) error {
	errorsList := make([]map[string]string, 0)

	payload.Type = models.TransactionType(strings.ToUpper(string(payload.Type)))
	payload.Currency = models.Currency(strings.ToUpper(string(payload.Currency)))
	payload.Month = models.Month(strings.ToUpper(string(payload.Month)))

	if _, ok := validTransactionTypes[payload.Type]; !ok {
		errorsList = append(errorsList, validationIssue(index, "type", fmt.Sprintf("Allowed values: %s", strings.Join(transactionKeys(), ", "))))
	}

	if payload.Amount <= 0 {
		errorsList = append(errorsList, validationIssue(index, "amount", "Amount must be greater than zero"))
	}

	if _, ok := validCurrencies[payload.Currency]; !ok {
		errorsList = append(errorsList, validationIssue(index, "currency", "Unsupported currency"))
	}

	if payload.Month == "" {
		errorsList = append(errorsList, validationIssue(index, "month", "Month is required"))
	} else if _, ok := validMonths[payload.Month]; !ok {
		errorsList = append(errorsList, validationIssue(index, "month", "Invalid month"))
	}

	if payload.Year < 2000 {
		errorsList = append(errorsList, validationIssue(index, "year", "Year must be >= 2000"))
	}

	if payload.Day != nil {
		if *payload.Day <= 0 || *payload.Day > day.MaxDays(string(payload.Month), payload.Year) {
			errorsList = append(errorsList, validationIssue(index, "day", "Day is out of range for the provided month"))
		}
	}

	if (payload.Currency == models.CurrencyUSD || payload.Currency == models.CurrencyEUR) && payload.ExchangeRate == nil {
		errorsList = append(errorsList, validationIssue(index, "exchangeRate", "Exchange rate is required for USD/EUR"))
	}

	if payload.CategoryID == nil && payload.Category == nil {
		errorsList = append(errorsList, validationIssue(index, "category", "Either categoryId or category must be provided"))
	}

	if payload.CategoryID != nil && payload.Category != nil {
		errorsList = append(errorsList, validationIssue(index, "category", "Provide only categoryId or category"))
	}

	if len(errorsList) > 0 {
		return apperror.New(apperror.ServerParamsMissing, errorsList)
	}

	return nil
}

func validationIssue(index int, field, message string) map[string]string {
	return map[string]string{
		"field": fmt.Sprintf("transactions[%d].%s", index, field),
		"msg":   message,
	}
}

var validTransactionTypes = map[models.TransactionType]struct{}{
	models.TransactionIncome:      {},
	models.TransactionExpense:     {},
	models.TransactionSaving:      {},
	models.TransactionInstallment: {},
}

var validCurrencies = map[models.Currency]struct{}{
	models.CurrencyUSD: {},
	models.CurrencyUYU: {},
	models.CurrencyEUR: {},
}

var validMonths = map[models.Month]struct{}{
	models.MonthJanuary:   {},
	models.MonthFebruary:  {},
	models.MonthMarch:     {},
	models.MonthApril:     {},
	models.MonthMay:       {},
	models.MonthJune:      {},
	models.MonthJuly:      {},
	models.MonthAugust:    {},
	models.MonthSeptember: {},
	models.MonthOctober:   {},
	models.MonthNovember:  {},
	models.MonthDecember:  {},
}

func transactionKeys() []string {
	keys := make([]string, 0, len(validTransactionTypes))
	for key := range validTransactionTypes {
		keys = append(keys, string(key))
	}

	return keys
}

// ValidTransactionTypes exposes the allowed transaction types (used by handlers for request validation).
func ValidTransactionTypes() map[models.TransactionType]struct{} {
	return validTransactionTypes
}

// ValidMonths exposes the valid month values.
func ValidMonths() map[models.Month]struct{} {
	return validMonths
}

// List returns every transaction that matches the provided filters.
func (s *TransactionService) List(ctx context.Context, userID string, filters TransactionFilters) ([]models.Transaction, error) {
	query := s.db.WithContext(ctx).Where("user_id = ?", userID)

	if filters.Type != nil {
		query = query.Where("type = ?", *filters.Type)
	}

	if filters.Day != nil {
		query = query.Where("day = ?", *filters.Day)
	}

	if filters.Month != nil {
		query = query.Where("month = ?", *filters.Month)
	}

	if filters.Year != nil {
		query = query.Where("year = ?", *filters.Year)
	}

	var transactions []models.Transaction
	if err := query.Preload("Category").Order("created_at DESC").Find(&transactions).Error; err != nil {
		return nil, err
	}

	return transactions, nil
}

func (s *TransactionService) GetByID(ctx context.Context, userID, transactionID string) (*models.Transaction, error) {
	var transaction models.Transaction
	if err := s.db.WithContext(ctx).
		Where("transaction_id = ? AND user_id = ?", transactionID, userID).
		Preload("Category").
		Take(&transaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(apperror.TransactionNotFound, nil)
		}

		return nil, err
	}

	return &transaction, nil
}

func (s *TransactionService) Delete(ctx context.Context, userID, transactionID string) error {
	result := s.db.WithContext(ctx).
		Where("transaction_id = ? AND user_id = ?", transactionID, userID).
		Delete(&models.Transaction{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return apperror.New(apperror.TransactionNotFound, nil)
	}

	return nil
}

func (s *TransactionService) Balances(ctx context.Context, userID string, filters TransactionFilters) (TransactionBalances, error) {
	transactions, err := s.List(ctx, userID, filters)
	if err != nil {
		return TransactionBalances{}, err
	}

	summary := TransactionBalances{}

	for _, trx := range transactions {
		switch trx.Type {
		case models.TransactionExpense, models.TransactionInstallment:
			summary.Expenses = addToSummary(summary.Expenses, trx)
		case models.TransactionIncome:
			summary.Incomes = addToSummary(summary.Incomes, trx)
		case models.TransactionSaving:
			summary.Savings = addToSummary(summary.Savings, trx)
		}
	}

	return summary, nil
}

func addToSummary(summary BalanceSummary, trx models.Transaction) BalanceSummary {
	switch trx.Currency {
	case models.CurrencyUYU:
		summary.UYU += trx.Amount
		summary.Total += trx.Amount
	case models.CurrencyUSD:
		summary.USD += trx.Amount
		if trx.ExchangeRate != nil {
			summary.Total += trx.Amount * *trx.ExchangeRate
		}
	case models.CurrencyEUR:
		summary.EUR += trx.Amount
		if trx.ExchangeRate != nil {
			summary.Total += trx.Amount * *trx.ExchangeRate
		}
	}

	summary.Total = round(summary.Total)
	summary.UYU = round(summary.UYU)
	summary.USD = round(summary.USD)
	summary.EUR = round(summary.EUR)

	return summary
}

func round(value float64) float64 {
	return math.Round(value*100) / 100
}

func (s *TransactionService) MonthsAndYears(ctx context.Context, userID string) (map[int][]models.Month, error) {
	type monthYear struct {
		Month models.Month
		Year  int
	}

	var rows []monthYear
	if err := s.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Select("month, year").
		Where("user_id = ?", userID).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[int][]models.Month)
	for _, row := range rows {
		months := result[row.Year]
		if !containsMonth(months, row.Month) {
			months = append(months, row.Month)
		}
		result[row.Year] = months
	}

	for year := range result {
		result[year] = sortMonths(result[year])
	}

	return result, nil
}

func containsMonth(list []models.Month, month models.Month) bool {
	for _, value := range list {
		if value == month {
			return true
		}
	}

	return false
}

func sortMonths(months []models.Month) []models.Month {
	ordered := append([]models.Month(nil), months...)
	order := []models.Month{
		models.MonthJanuary,
		models.MonthFebruary,
		models.MonthMarch,
		models.MonthApril,
		models.MonthMay,
		models.MonthJune,
		models.MonthJuly,
		models.MonthAugust,
		models.MonthSeptember,
		models.MonthOctober,
		models.MonthNovember,
		models.MonthDecember,
	}

	sortIndex := make(map[models.Month]int)
	for idx, month := range order {
		sortIndex[month] = idx
	}

	sort.Slice(ordered, func(i, j int) bool {
		return sortIndex[ordered[i]] < sortIndex[ordered[j]]
	})

	return ordered
}

func (s *TransactionService) TotalSavings(ctx context.Context, userID string) (float64, error) {
	var total float64
	if err := s.db.WithContext(ctx).
		Model(&models.Transaction{}).
		Select("COALESCE(SUM(amount),0)").
		Where("user_id = ? AND type = ?", userID, models.TransactionSaving).
		Scan(&total).Error; err != nil {
		return 0, err
	}

	return round(total), nil
}
