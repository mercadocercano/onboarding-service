package request

import (
	"errors"
	"strings"
)

// CompleteOnboardingRequest representa la solicitud para completar el onboarding
type CompleteOnboardingRequest struct {
	ProcessID string `json:"process_id" binding:"required"`
}

// Validate valida la solicitud de completar onboarding
func (r *CompleteOnboardingRequest) Validate() error {
	// Validar process_id
	processID := strings.TrimSpace(r.ProcessID)
	if processID == "" {
		return errors.New("el ID del proceso es requerido")
	}
	r.ProcessID = processID

	return nil
}

// GetCleanProcessID retorna el process ID limpio
func (r *CompleteOnboardingRequest) GetCleanProcessID() string {
	return strings.TrimSpace(r.ProcessID)
}
