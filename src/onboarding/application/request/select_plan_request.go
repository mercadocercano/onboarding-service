package request

import (
	"errors"
	"strings"
)

// SelectPlanRequest representa la solicitud para seleccionar un plan
type SelectPlanRequest struct {
	ProcessID    string `json:"process_id" binding:"required"`
	SelectedPlan string `json:"selected_plan" binding:"required"`
}

// Validate valida la solicitud de selección de plan
func (r *SelectPlanRequest) Validate() error {
	// Validar process_id
	processID := strings.TrimSpace(r.ProcessID)
	if processID == "" {
		return errors.New("el ID del proceso es requerido")
	}
	r.ProcessID = processID

	// Validar selected_plan
	selectedPlan := strings.TrimSpace(r.SelectedPlan)
	if selectedPlan == "" {
		return errors.New("el plan seleccionado es requerido")
	}

	// Validar que el plan sea válido
	validPlans := map[string]bool{
		"free":       true,
		"basic":      true,
		"premium":    true,
		"enterprise": true,
	}

	if !validPlans[selectedPlan] {
		return errors.New("plan seleccionado no válido. Debe ser: free, basic, premium o enterprise")
	}

	r.SelectedPlan = selectedPlan
	return nil
}

// GetCleanPlan retorna el plan limpio
func (r *SelectPlanRequest) GetCleanPlan() string {
	return strings.TrimSpace(r.SelectedPlan)
}
