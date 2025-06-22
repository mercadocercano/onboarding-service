package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse representa la estructura estándar de respuesta de error
type ErrorResponse struct {
	Success bool  `json:"success"`
	Error   Error `json:"error"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// BusinessError representa errores de negocio que pueden ser manejados
type BusinessError struct {
	Code       string
	Message    string
	Details    string
	HTTPStatus int
}

func (e BusinessError) Error() string {
	return e.Message
}

// Errores de negocio predefinidos
var (
	ErrInvalidRequestFormat = BusinessError{
		Code:       "INVALID_REQUEST_FORMAT",
		Message:    "Datos inválidos en la solicitud",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidEmail = BusinessError{
		Code:       "INVALID_EMAIL",
		Message:    "El formato del email es inválido",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrProcessNotFound = BusinessError{
		Code:       "PROCESS_NOT_FOUND",
		Message:    "El proceso de onboarding no fue encontrado",
		HTTPStatus: http.StatusNotFound,
	}

	ErrInvalidVerificationCode = BusinessError{
		Code:       "INVALID_VERIFICATION_CODE",
		Message:    "El código de verificación es inválido",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrRateLimited = BusinessError{
		Code:       "RATE_LIMITED",
		Message:    "Demasiadas solicitudes, intenta más tarde",
		HTTPStatus: http.StatusTooManyRequests,
	}

	ErrProcessAlreadyCompleted = BusinessError{
		Code:       "PROCESS_ALREADY_COMPLETED",
		Message:    "El proceso ya está completado",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInternalServer = BusinessError{
		Code:       "INTERNAL_SERVER_ERROR",
		Message:    "Error interno del servidor",
		HTTPStatus: http.StatusInternalServerError,
	}
)

// ErrorHandlerMiddleware maneja todos los errores de forma centralizada
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Ejecutar el siguiente handler
		ctx.Next()

		// Si hay errores, manejarlos
		if len(ctx.Errors) > 0 {
			err := ctx.Errors.Last().Err
			handleError(ctx, err)
		}
	}
}

// handleError maneja diferentes tipos de errores
func handleError(ctx *gin.Context, err error) {
	switch e := err.(type) {
	case BusinessError:
		// Error de negocio - respuesta controlada
		log.Printf("Business error occurred: code=%s, message=%s, path=%s",
			e.Code, e.Message, ctx.Request.URL.Path)

		ctx.JSON(e.HTTPStatus, ErrorResponse{
			Success: false,
			Error: Error{
				Code:    e.Code,
				Message: e.Message,
				Details: e.Details,
			},
		})

	default:
		// Error no controlado - error interno
		log.Printf("Unhandled error occurred: %v, path=%s, method=%s",
			err, ctx.Request.URL.Path, ctx.Request.Method)

		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Success: false,
			Error: Error{
				Code:    "INTERNAL_SERVER_ERROR",
				Message: "Error interno del servidor",
				Details: err.Error(),
			},
		})
	}
}

// AbortWithBusinessError es un helper para abortar con error de negocio
func AbortWithBusinessError(ctx *gin.Context, err BusinessError) {
	ctx.Error(err)
	ctx.Abort()
}

// AbortWithError es un helper para abortar con error genérico
func AbortWithError(ctx *gin.Context, err error) {
	ctx.Error(err)
	ctx.Abort()
}

// ShouldBindJSONWithError helper para binding con manejo de errores
func ShouldBindJSONWithError(ctx *gin.Context, obj interface{}) error {
	if err := ctx.ShouldBindJSON(obj); err != nil {
		AbortWithBusinessError(ctx, ErrInvalidRequestFormat)
		return err
	}
	return nil
}
