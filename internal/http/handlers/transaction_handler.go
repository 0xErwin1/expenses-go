package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/iperez/new-expenses-go/internal/domain/models"
	"github.com/iperez/new-expenses-go/internal/http/middleware"
	"github.com/iperez/new-expenses-go/internal/service"
	"github.com/iperez/new-expenses-go/pkg/apperror"
	"github.com/iperez/new-expenses-go/pkg/response"
)

// TransactionHandler exposes the transaction endpoints.
type TransactionHandler struct {
	transactions *service.TransactionService
}

func NewTransactionHandler(transactions *service.TransactionService) *TransactionHandler {
	return &TransactionHandler{transactions: transactions}
}

func (h *TransactionHandler) Register(router fiber.Router) {
	router.Get("/", middleware.RequireAuth(), h.List)
	router.Post("/", middleware.RequireAuth(), h.Create)
	router.Get("/balance", middleware.RequireAuth(), h.Balance)
	router.Get("/month-by-years", middleware.RequireAuth(), h.MonthsByYear)
	router.Get("/total-saving", middleware.RequireAuth(), h.TotalSavings)
	router.Get("/:transactionId", middleware.RequireAuth(), h.Get)
	router.Delete("/:transactionId", middleware.RequireAuth(), h.Delete)
}

type transactionEnvelope struct {
	service.CreateTransactionInput
	Transactions []service.CreateTransactionInput `json:"transactions"`
}

func (h *TransactionHandler) Create(c *fiber.Ctx) error {
	var payload transactionEnvelope
	if err := c.BodyParser(&payload); err != nil {
		if ute, ok := err.(*json.UnmarshalTypeError); ok {
			msg := fmt.Sprintf("field '%s' expects %s", ute.Field, ute.Type.String())
			return apperror.New(apperror.ServerParamsMissing, msg)
		}
		return apperror.New(apperror.ServerParamsMissing, err.Error())
	}

	inputs := payload.Transactions
	if len(inputs) == 0 {
		inputs = []service.CreateTransactionInput{payload.CreateTransactionInput}
	}

	for idx := range inputs {
		sanitizeTransactionInput(&inputs[idx])
	}

	transactions, err := h.transactions.Create(c.UserContext(), middleware.UserID(c), inputs)
	if err != nil {
		return err
	}

	responses := make([]transactionResponse, 0, len(transactions))
	for idx := range transactions {
		responses = append(responses, newTransactionResponse(&transactions[idx]))
	}

	if len(responses) == 1 {
		return c.Status(fiber.StatusCreated).JSON(response.Success(responses[0]))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success(responses))
}

func (h *TransactionHandler) List(c *fiber.Ctx) error {
	filters, err := parseFilters(c)
	if err != nil {
		return err
	}

	transactions, err := h.transactions.List(c.UserContext(), middleware.UserID(c), filters)
	if err != nil {
		return err
	}

	responses := make([]transactionResponse, 0, len(transactions))
	for idx := range transactions {
		responses = append(responses, newTransactionResponse(&transactions[idx]))
	}

	return c.JSON(response.Success(responses))
}

func (h *TransactionHandler) Get(c *fiber.Ctx) error {
	transaction, err := h.transactions.GetByID(c.UserContext(), middleware.UserID(c), c.Params("transactionId"))
	if err != nil {
		return err
	}

	return c.JSON(response.Success(newTransactionResponse(transaction)))
}

func (h *TransactionHandler) Delete(c *fiber.Ctx) error {
	if err := h.transactions.Delete(c.UserContext(), middleware.UserID(c), c.Params("transactionId")); err != nil {
		return err
	}

	return c.JSON(response.Success(nil))
}

func (h *TransactionHandler) Balance(c *fiber.Ctx) error {
	filters, err := parseFilters(c)
	if err != nil {
		return err
	}

	balances, err := h.transactions.Balances(c.UserContext(), middleware.UserID(c), filters)
	if err != nil {
		return err
	}

	return c.JSON(response.Success(balances))
}

func (h *TransactionHandler) MonthsByYear(c *fiber.Ctx) error {
	result, err := h.transactions.MonthsAndYears(c.UserContext(), middleware.UserID(c))
	if err != nil {
		return err
	}

	formatted := make(map[string][]string)
	for year, months := range result {
		key := strconv.Itoa(year)
		formatted[key] = make([]string, 0, len(months))
		for _, month := range months {
			formatted[key] = append(formatted[key], string(month))
		}
	}

	return c.JSON(response.Success(formatted))
}

func (h *TransactionHandler) TotalSavings(c *fiber.Ctx) error {
	total, err := h.transactions.TotalSavings(c.UserContext(), middleware.UserID(c))
	if err != nil {
		return err
	}

	return c.JSON(response.Success(fiber.Map{"totalSavings": total}))
}

func parseFilters(c *fiber.Ctx) (service.TransactionFilters, error) {
	filters := service.TransactionFilters{}
	validTypes := service.ValidTransactionTypes()
	validMonths := service.ValidMonths()

	if typeParam := c.Query("type"); typeParam != "" {
		normalized := models.TransactionType(strings.ToUpper(typeParam))
		if _, ok := validTypes[normalized]; !ok {
			return filters, apperror.New(apperror.ServerParamsMissing, "Invalid transaction type")
		}
		filters.Type = &normalized
	}

	if monthParam := c.Query("month"); monthParam != "" {
		normalized := models.Month(strings.ToUpper(monthParam))
		if _, ok := validMonths[normalized]; !ok {
			return filters, apperror.New(apperror.ServerParamsMissing, "Invalid month")
		}
		filters.Month = &normalized
	}

	if dayParam := c.Query("day"); dayParam != "" {
		dayValue, err := strconv.Atoi(dayParam)
		if err != nil || dayValue <= 0 || dayValue > 31 {
			return filters, apperror.New(apperror.ServerParamsMissing, "Day must be between 1 and 31")
		}
		filters.Day = &dayValue
	}

	if yearParam := c.Query("year"); yearParam != "" {
		yearValue, err := strconv.Atoi(yearParam)
		if err != nil || yearValue < 2000 {
			return filters, apperror.New(apperror.ServerParamsMissing, "Year must be >= 2000")
		}
		filters.Year = &yearValue
	}

	return filters, nil
}

func sanitizeTransactionInput(input *service.CreateTransactionInput) {
	input.Note = strings.TrimSpace(input.Note)

	if input.CategoryID != nil && strings.TrimSpace(*input.CategoryID) == "" {
		input.CategoryID = nil
	}

	if input.Category != nil {
		input.Category.Name = strings.TrimSpace(input.Category.Name)
		input.Category.Note = strings.TrimSpace(input.Category.Note)
		input.Category.Type = models.TransactionType(strings.ToUpper(string(input.Category.Type)))
	}
}

type transactionResponse struct {
	TransactionID string            `json:"transactionId"`
	Type          string            `json:"type"`
	Amount        float64           `json:"amount"`
	Currency      string            `json:"currency"`
	Note          string            `json:"note"`
	Day           *int              `json:"day"`
	Month         string            `json:"month"`
	Year          int               `json:"year"`
	ExchangeRate  *float64          `json:"exchangeRate"`
	CategoryID    *string           `json:"categoryId"`
	Category      *categoryResponse `json:"category"`
}

func newTransactionResponse(transaction *models.Transaction) transactionResponse {
	var category *categoryResponse
	if transaction.Category != nil {
		cat := newCategoryResponse(transaction.Category)
		category = &cat
	}

	return transactionResponse{
		TransactionID: transaction.TransactionID,
		Type:          string(transaction.Type),
		Amount:        transaction.Amount,
		Currency:      string(transaction.Currency),
		Note:          transaction.Note,
		Day:           transaction.Day,
		Month:         string(transaction.Month),
		Year:          transaction.Year,
		ExchangeRate:  transaction.ExchangeRate,
		CategoryID:    transaction.CategoryID,
		Category:      category,
	}
}
