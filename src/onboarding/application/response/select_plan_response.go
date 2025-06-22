package response

import (
	"time"
)

// SelectPlanResponse representa la respuesta de la selección de plan
type SelectPlanResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ProcessID string `json:"process_id"`

	// Datos del plan seleccionado
	PlanSelected   string `json:"plan_selected"`
	PlanConfigured bool   `json:"plan_configured"`

	// Estado actualizado del proceso
	CurrentStep    int    `json:"current_step"`
	StepsCompleted []int  `json:"steps_completed"`
	StepsPending   []int  `json:"steps_pending"`
	NextStepURL    string `json:"next_step_url"`

	CreatedAt time.Time `json:"created_at"`
}

// NewSelectPlanResponse crea una nueva respuesta de selección exitosa
func NewSelectPlanResponse(
	processID, planSelected string,
	currentStep int,
	stepsCompleted []int,
	stepsPending []int,
	nextStepURL string,
) *SelectPlanResponse {
	return &SelectPlanResponse{
		Success:        true,
		Message:        "Plan seleccionado exitosamente",
		ProcessID:      processID,
		PlanSelected:   planSelected,
		PlanConfigured: true,
		CurrentStep:    currentStep,
		StepsCompleted: stepsCompleted,
		StepsPending:   stepsPending,
		NextStepURL:    nextStepURL,
		CreatedAt:      time.Now(),
	}
}

// NewSelectPlanErrorResponse crea una respuesta de error
func NewSelectPlanErrorResponse(message string) *SelectPlanResponse {
	return &SelectPlanResponse{
		Success:   false,
		Message:   message,
		CreatedAt: time.Now(),
	}
}
