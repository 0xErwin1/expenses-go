package service

import "github.com/go-playground/validator/v10"

// formatValidationErrors converts validator errors into a client-friendly slice of field/message entries.
func formatValidationErrors(err error) []map[string]string {
	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return []map[string]string{{"field": "", "msg": err.Error()}}
	}

	issues := make([]map[string]string, 0, len(verrs))
	for _, fe := range verrs {
		issues = append(issues, map[string]string{
			"field": fe.Field(),
			"msg":   fe.Error(),
		})
	}

	return issues
}
