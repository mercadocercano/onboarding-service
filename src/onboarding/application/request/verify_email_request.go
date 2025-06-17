package request

import (
	"fmt"
	"strings"
)

// VerifyEmailRequest representa la solicitud para verificar email
type VerifyEmailRequest struct {
	ProcessID        string `json:"process_id" binding:"required"`        // ID del proceso de onboarding
	VerificationCode string `json:"verification_code" binding:"required"` // Código de verificación de 6 dígitos
}

// Validate valida la solicitud de verificación de email
func (r *VerifyEmailRequest) Validate() error {
	if strings.TrimSpace(r.ProcessID) == "" {
		return fmt.Errorf("process_id es requerido")
	}

	if strings.TrimSpace(r.VerificationCode) == "" {
		return fmt.Errorf("verification_code es requerido")
	}

	// Limpiar el código de verificación
	r.VerificationCode = strings.TrimSpace(r.VerificationCode)
	r.VerificationCode = strings.ReplaceAll(r.VerificationCode, " ", "")
	r.VerificationCode = strings.ReplaceAll(r.VerificationCode, "-", "")

	// Validar formato del código (6 dígitos)
	if len(r.VerificationCode) != 6 {
		return fmt.Errorf("el código de verificación debe tener 6 dígitos")
	}

	// Validar que solo contenga dígitos
	for _, char := range r.VerificationCode {
		if char < '0' || char > '9' {
			return fmt.Errorf("el código de verificación solo puede contener dígitos")
		}
	}

	return nil
}

// Sanitize limpia y normaliza los datos de entrada
func (r *VerifyEmailRequest) Sanitize() {
	r.ProcessID = strings.TrimSpace(r.ProcessID)
	r.VerificationCode = strings.TrimSpace(r.VerificationCode)
	r.VerificationCode = strings.ReplaceAll(r.VerificationCode, " ", "")
	r.VerificationCode = strings.ReplaceAll(r.VerificationCode, "-", "")
}
