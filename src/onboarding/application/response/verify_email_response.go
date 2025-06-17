package response

// VerifyEmailResponse representa la respuesta de la verificación de email
type VerifyEmailResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	ProcessID      string `json:"process_id"`
	Verified       bool   `json:"verified"`
	CurrentStep    int    `json:"current_step"`
	StepsCompleted []int  `json:"steps_completed"`
	StepsPending   []int  `json:"steps_pending"`
	NextStepURL    string `json:"next_step_url"`
	Error          string `json:"error,omitempty"`
}

// NewVerifyEmailSuccessResponse crea una respuesta exitosa
func NewVerifyEmailSuccessResponse(processID string) *VerifyEmailResponse {
	return &VerifyEmailResponse{
		Success:        true,
		Message:        "Email verificado exitosamente",
		ProcessID:      processID,
		Verified:       true,
		CurrentStep:    4,
		StepsCompleted: []int{1, 2, 3},
		StepsPending:   []int{4, 5},
		NextStepURL:    "/onboarding/configurar-tienda",
	}
}

// NewVerifyEmailErrorResponse crea una respuesta de error
func NewVerifyEmailErrorResponse(message string, err error) *VerifyEmailResponse {
	response := &VerifyEmailResponse{
		Success:  false,
		Message:  message,
		Verified: false,
	}

	if err != nil {
		response.Error = err.Error()
	}

	return response
}

// NewVerifyEmailInvalidCodeResponse crea una respuesta para código inválido
func NewVerifyEmailInvalidCodeResponse(processID string) *VerifyEmailResponse {
	return &VerifyEmailResponse{
		Success:   false,
		Message:   "Código de verificación inválido",
		ProcessID: processID,
		Verified:  false,
		Error:     "INVALID_VERIFICATION_CODE",
	}
}
