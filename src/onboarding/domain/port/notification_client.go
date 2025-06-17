package port

import "context"

// NotificationClient define la interfaz para interactuar con el servicio de notificaciones
type NotificationClient interface {
	SendEmailVerification(ctx context.Context, email, name, verificationCode string) error
	SendWelcomeEmail(ctx context.Context, email, name, companyName string) error
}

// SendEmailVerificationRequest representa la estructura para envío de email de verificación
type SendEmailVerificationRequest struct {
	Type      string                 `json:"type"`
	Action    string                 `json:"action"`
	Recipient string                 `json:"recipient"`
	Data      map[string]interface{} `json:"data"`
	Async     bool                   `json:"async"`
}

// NotificationResponse representa la respuesta del servicio de notificaciones
// Este es el formato real que devuelve el servicio (sin success/error wrapper)
type NotificationResponse struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// NotificationErrorResponse representa una respuesta de error del servicio de notificaciones
type NotificationErrorResponse struct {
	Success bool `json:"success"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details string `json:"details,omitempty"`
	} `json:"error,omitempty"`
}
