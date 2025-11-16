package response

import "github.com/iperez/new-expenses-go/pkg/apperror"

// CustomResponse mimics the shape exposed by the original TypeScript service.
type CustomResponse struct {
	Result      bool              `json:"result"`
	Data        interface{}       `json:"data,omitempty"`
	Message     string            `json:"message,omitempty"`
	ShowMessage map[string]string `json:"showMessage,omitempty"`
	ErrorCode   apperror.Code     `json:"errorCode,omitempty"`
}

// Success wraps any payload into a success response.
func Success(data interface{}) CustomResponse {
	return CustomResponse{Result: true, Data: data}
}

// Error builds a formatted error response that includes the public message and code.
func Error(code apperror.Code, data interface{}) CustomResponse {
	definition := apperror.Lookup(code)

	return CustomResponse{
		Result:      false,
		Data:        data,
		Message:     definition.Message,
		ShowMessage: definition.ShowMessage,
		ErrorCode:   code,
	}
}
