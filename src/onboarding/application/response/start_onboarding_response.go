package response

// StartOnboardingResponse representa la respuesta del inicio de onboarding
type StartOnboardingResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	ProcessID      string `json:"process_id"`
	CurrentStep    int    `json:"current_step"`
	StepsCompleted []int  `json:"steps_completed"`
	StepsPending   []int  `json:"steps_pending"`
	NextStepURL    string `json:"next_step_url"`
	Error          string `json:"error,omitempty"`
}

// NewStartOnboardingSuccessResponse crea una respuesta exitosa
func NewStartOnboardingSuccessResponse(processID string) *StartOnboardingResponse {
	return &StartOnboardingResponse{
		Success:        true,
		Message:        "Proceso de onboarding iniciado exitosamente",
		ProcessID:      processID,
		CurrentStep:    1,
		StepsCompleted: []int{},
		StepsPending:   []int{1, 2, 3, 4, 5},
		NextStepURL:    "/onboarding/bienvenida",
	}
}

// NewStartOnboardingErrorResponse crea una respuesta de error
func NewStartOnboardingErrorResponse(message string, err error) *StartOnboardingResponse {
	response := &StartOnboardingResponse{
		Success: false,
		Message: message,
	}

	if err != nil {
		response.Error = err.Error()
	}

	return response
}
