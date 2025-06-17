package response

import (
	"time"
)

// GetProcessStatusResponse representa la respuesta del estado del proceso
type GetProcessStatusResponse struct {
	Success         bool      `json:"success"`
	Message         string    `json:"message"`
	ProcessID       string    `json:"process_id"`
	Status          string    `json:"status"` // "in_progress" | "completed"
	CurrentStep     int       `json:"current_step"`
	StepsCompleted  []int     `json:"steps_completed"`
	ProgressPercent float64   `json:"progress_percent"`
	RetrievedAt     time.Time `json:"retrieved_at"`
}

// NewGetProcessStatusResponse crea una nueva respuesta exitosa
func NewGetProcessStatusResponse(processID, status string, currentStep int, stepsCompleted []int, progressPercent float64) *GetProcessStatusResponse {
	return &GetProcessStatusResponse{
		Success:         true,
		Message:         "Estado del proceso obtenido exitosamente",
		ProcessID:       processID,
		Status:          status,
		CurrentStep:     currentStep,
		StepsCompleted:  stepsCompleted,
		ProgressPercent: progressPercent,
		RetrievedAt:     time.Now(),
	}
}

// NewGetProcessStatusErrorResponse crea una respuesta de error
func NewGetProcessStatusErrorResponse(message string) *GetProcessStatusResponse {
	return &GetProcessStatusResponse{
		Success:     false,
		Message:     message,
		RetrievedAt: time.Now(),
	}
}
