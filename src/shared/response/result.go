package response

import (
	"onboarding/src/shared/middleware"
)

// Result representa una respuesta estándar de use case
type Result struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// SuccessResult crea un resultado exitoso
func SuccessResult(data interface{}, message string) Result {
	return Result{
		Success: true,
		Data:    data,
		Message: message,
	}
}

// ErrorResult crea un resultado con error
func ErrorResult(errorCode string, message string) Result {
	return Result{
		Success: false,
		Error:   errorCode,
		Message: message,
	}
}

// ToMiddlewareError convierte el error del resultado a BusinessError del middleware
func (r Result) ToMiddlewareError() middleware.BusinessError {
	if r.Success {
		return middleware.BusinessError{}
	}

	// Mapear códigos de error específicos
	switch r.Error {
	case "INVALID_VERIFICATION_CODE":
		return middleware.ErrInvalidVerificationCode
	case "RATE_LIMITED":
		return middleware.ErrRateLimited
	case "PROCESS_NOT_FOUND":
		return middleware.ErrProcessNotFound
	case "PROCESS_ALREADY_COMPLETED":
		return middleware.ErrProcessAlreadyCompleted
	case "INVALID_EMAIL":
		return middleware.ErrInvalidEmail
	default:
		return middleware.ErrInternalServer
	}
}
