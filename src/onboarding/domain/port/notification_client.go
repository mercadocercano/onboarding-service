package port

import (
	"context"
)

// NotificationRequest representa una solicitud de notificación
type NotificationRequest struct {
	Type      string                 `json:"type"`
	Action    string                 `json:"action"`
	Recipient string                 `json:"recipient"`
	Data      map[string]interface{} `json:"data"`
	Async     bool                   `json:"async"`
}

// NotificationResponse representa la respuesta del servicio de notificaciones
type NotificationResponse struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
}

// NotificationClient define el contrato para el cliente de notificaciones
type NotificationClient interface {
	SendEmailVerification(ctx context.Context, email, name, verificationCode string) error
	SendWelcomeEmail(ctx context.Context, email, userName, companyName, businessType string) error
	SendNotification(ctx context.Context, req *NotificationRequest) (*NotificationResponse, error)
}

// SendEmailVerificationRequest representa la estructura para envío de email de verificación
type SendEmailVerificationRequest struct {
	Type      string                 `json:"type"`
	Action    string                 `json:"action"`
	Recipient string                 `json:"recipient"`
	Data      map[string]interface{} `json:"data"`
	Async     bool                   `json:"async"`
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
