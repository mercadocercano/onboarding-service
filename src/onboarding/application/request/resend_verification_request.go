package request

import (
	"fmt"
	"strings"
)

// ResendVerificationRequest representa la solicitud para reenviar email de verificación
type ResendVerificationRequest struct {
	ProcessID string `json:"process_id" binding:"required"` // ID del proceso de onboarding
	Email     string `json:"email" binding:"required"`      // Email del usuario (para seguridad)
}

// Validate valida la solicitud de reenvío de verificación
func (r *ResendVerificationRequest) Validate() error {
	if strings.TrimSpace(r.ProcessID) == "" {
		return fmt.Errorf("process_id es requerido")
	}

	if strings.TrimSpace(r.Email) == "" {
		return fmt.Errorf("email es requerido")
	}

	// Validar formato básico de email
	if !strings.Contains(r.Email, "@") || !strings.Contains(r.Email, ".") {
		return fmt.Errorf("formato de email inválido")
	}

	return nil
}

// Sanitize limpia y normaliza los datos de entrada
func (r *ResendVerificationRequest) Sanitize() {
	r.ProcessID = strings.TrimSpace(r.ProcessID)
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
}
