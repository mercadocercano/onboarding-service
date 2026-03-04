package response

import (
	"time"
)

// RegisterUserResponse representa la respuesta del registro de usuario
type RegisterUserResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ProcessID string `json:"process_id"`
	TenantID  string `json:"tenant_id"`
	UserID    string `json:"user_id"`
	NextStep  int    `json:"next_step"`

	// Autenticación (para poder acceder al backoffice inmediatamente después)
	AccessToken  string `json:"access_token,omitempty"`  // JWT token
	RefreshToken string `json:"refresh_token,omitempty"` // Para renovar sesión (misma estructura que login)

	// Datos del usuario creado
	UserData UserData `json:"user_data"`

	// Información del tenant
	TenantData TenantData `json:"tenant_data"`

	CreatedAt time.Time `json:"created_at"`
}

// UserData contiene información básica del usuario
type UserData struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// TenantData contiene información básica del tenant
type TenantData struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// NewRegisterUserResponse crea una nueva respuesta de registro exitoso
func NewRegisterUserResponse(processID, tenantID, userID, accessToken, refreshToken string, userData UserData, tenantData TenantData) *RegisterUserResponse {
	return &RegisterUserResponse{
		Success:      true,
		Message:      "Usuario registrado exitosamente",
		ProcessID:    processID,
		TenantID:     tenantID,
		UserID:       userID,
		NextStep:     3, // Siguiente paso: verificación de email
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserData:     userData,
		TenantData:   tenantData,
		CreatedAt:    time.Now(),
	}
}

// NewRegisterUserErrorResponse crea una respuesta de error
func NewRegisterUserErrorResponse(message string) *RegisterUserResponse {
	return &RegisterUserResponse{
		Success:   false,
		Message:   message,
		NextStep:  2, // Permanecer en el paso de registro
		CreatedAt: time.Now(),
	}
}
