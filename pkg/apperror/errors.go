package apperror

import (
	"fmt"
	"net/http"
)

// Code represents the API error code exposed in HTTP responses.
type Code int

const (
	// Auth errors.
	AuthBadAuth   Code = 1002
	AuthNeedLogin Code = 1005
	// Server errors.
	ServerTooFewParams  Code = 2001
	ServerParamsMissing Code = 2002
	ServerNotFound      Code = 2003
	// User errors.
	UserNotFound Code = 3001
	UserExists   Code = 3004
	// Transaction errors.
	TransactionNotFound             Code = 4001
	TransactionCategoryTypeMismatch Code = 4002
	// Category errors.
	CategoryNotFound        Code = 5001
	CategoryHasTransactions Code = 5002
)

// Definition holds the metadata returned alongside an error response.
type Definition struct {
	Message     string            `json:"message"`
	ShowMessage map[string]string `json:"showMessage"`
	HTTPStatus  int               `json:"HTTPStatusCode"`
}

var registry = map[Code]Definition{
	AuthBadAuth: {
		Message: "Bad auth",
		ShowMessage: map[string]string{
			"EN": "Incorrect email/password",
			"ES": "Email o contraseña incorrectos",
		},
		HTTPStatus: http.StatusBadRequest,
	},
	AuthNeedLogin: {
		Message: "You need to be logged in",
		ShowMessage: map[string]string{
			"EN": "You need to be logged in",
			"ES": "Debe estar conectado",
		},
		HTTPStatus: http.StatusUnauthorized,
	},
	ServerTooFewParams: {
		Message: "Too few parameters",
		ShowMessage: map[string]string{
			"EN": "Too few parameters",
			"ES": "Faltan parametros",
		},
		HTTPStatus: http.StatusBadRequest,
	},
	ServerParamsMissing: {
		Message: "Some body parameters are missing or are incorrect",
		ShowMessage: map[string]string{
			"EN": "Some body parameters are missing or are incorrect",
			"ES": "Faltan o son incorrectos algunos parametros de la solicitud",
		},
		HTTPStatus: http.StatusBadRequest,
	},
	ServerNotFound: {
		Message: "Not found",
		ShowMessage: map[string]string{
			"EN": "Not found",
			"ES": "Recurso no encontrado",
		},
		HTTPStatus: http.StatusNotFound,
	},
	UserNotFound: {
		Message: "User does not exist",
		ShowMessage: map[string]string{
			"EN": "User does not exist",
			"ES": "El usuario no existe",
		},
		HTTPStatus: http.StatusNotFound,
	},
	UserExists: {
		Message: "User already exists",
		ShowMessage: map[string]string{
			"EN": "User already exists",
			"ES": "El usuario ya existe",
		},
		HTTPStatus: http.StatusConflict,
	},
	TransactionNotFound: {
		Message: "Transaction not exist",
		ShowMessage: map[string]string{
			"EN": "Transaction not exist",
			"ES": "La transacción no existe",
		},
		HTTPStatus: http.StatusNotFound,
	},
	TransactionCategoryTypeMismatch: {
		Message: "Transaction and category are not of the same type.",
		ShowMessage: map[string]string{
			"EN": "Transaction and category are not of the same type.",
			"ES": "La transacción y la categoría no son del mismo tipo.",
		},
		HTTPStatus: http.StatusConflict,
	},
	CategoryNotFound: {
		Message: "Category not exist",
		ShowMessage: map[string]string{
			"EN": "Category not exist",
			"ES": "La categoria no existe",
		},
		HTTPStatus: http.StatusNotFound,
	},
	CategoryHasTransactions: {
		Message: "Cannot delete a category with transactions",
		ShowMessage: map[string]string{
			"EN": "Cannot delete a category with transactions, try with the query ?deleteTransactions=true to delete all transactions.",
			"ES": "No se puede eliminar una categoría con transacciones, pruebe con la query ?deleteTransactions=true para eliminar todas las transacciones",
		},
		HTTPStatus: http.StatusBadRequest,
	},
}

// AppError implements the Go error interface with custom metadata.
type AppError struct {
	Code Code
	Data interface{}
}

func (e AppError) Error() string {
	definition, ok := registry[e.Code]
	if !ok {
		return fmt.Sprintf("unknown error code %d", e.Code)
	}

	return definition.Message
}

// New creates a new AppError with the provided code and optional payload.
func New(code Code, data interface{}) AppError {
	return AppError{
		Code: code,
		Data: data,
	}
}

// Lookup returns the metadata associated with the provided code.
func Lookup(code Code) Definition {
	if def, ok := registry[code]; ok {
		return def
	}

	return Definition{
		Message:     "Unexpected error",
		ShowMessage: map[string]string{"EN": "Unexpected error", "ES": "Error inesperado"},
		HTTPStatus:  http.StatusInternalServerError,
	}
}
