package response

import "time"

// ResendVerificationResponse representa la respuesta del reenvío de email de verificación
type ResendVerificationResponse struct {
	Success        bool      `json:"success"`
	Message        string    `json:"message"`
	ProcessID      string    `json:"process_id"`
	EmailSent      bool      `json:"email_sent"`
	NewCodeSent    bool      `json:"new_code_sent"`
	ExpiryTime     string    `json:"expiry_time"`
	NextResendTime int       `json:"next_resend_time_seconds"` // Tiempo en segundos hasta poder reenviar
	SentAt         time.Time `json:"sent_at"`
	Error          string    `json:"error,omitempty"`
}

// NewResendVerificationSuccessResponse crea una respuesta exitosa
func NewResendVerificationSuccessResponse(processID string) *ResendVerificationResponse {
	return &ResendVerificationResponse{
		Success:        true,
		Message:        "Email de verificación reenviado exitosamente",
		ProcessID:      processID,
		EmailSent:      true,
		NewCodeSent:    true,
		ExpiryTime:     "15 minutos",
		NextResendTime: 60, // 60 segundos hasta poder reenviar
		SentAt:         time.Now(),
	}
}

// NewResendVerificationErrorResponse crea una respuesta de error
func NewResendVerificationErrorResponse(message string, err error) *ResendVerificationResponse {
	response := &ResendVerificationResponse{
		Success:   false,
		Message:   message,
		EmailSent: false,
	}

	if err != nil {
		response.Error = err.Error()
	}

	return response
}

// NewResendVerificationThrottleResponse crea una respuesta para cuando se supera el límite de reenvíos
func NewResendVerificationThrottleResponse(processID string, waitSeconds int) *ResendVerificationResponse {
	return &ResendVerificationResponse{
		Success:        false,
		Message:        "Debes esperar antes de solicitar otro código",
		ProcessID:      processID,
		EmailSent:      false,
		NextResendTime: waitSeconds,
		Error:          "RATE_LIMITED",
	}
}
