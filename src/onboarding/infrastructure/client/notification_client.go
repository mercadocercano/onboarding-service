package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"onboarding/src/onboarding/domain/port"
)

// NotificationHTTPClient implementa NotificationClient usando HTTP
type NotificationHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewNotificationClient crea una nueva instancia del cliente de notificaciones
func NewNotificationClient() port.NotificationClient {
	baseURL := getNotificationEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8282")

	return &NotificationHTTPClient{
		baseURL: baseURL + "/api/v1",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendEmailVerification envía un email de verificación con código
func (c *NotificationHTTPClient) SendEmailVerification(ctx context.Context, email, name, verificationCode string) error {
	log.Printf("Sending email verification to: %s, name: %s, code: %s", email, name, verificationCode)

	// Preparar datos para el template
	data := map[string]interface{}{
		"name":          name,
		"token":         verificationCode,
		"company":       "TiendaVecina",
		"expiry_time":   "15 minutos",
		"support_email": "soporte@tiendavecina.com",
	}

	// Preparar request
	notificationReq := &port.SendEmailVerificationRequest{
		Type:      "email",
		Action:    "EMAIL_VERIFICATION",
		Recipient: email,
		Data:      data,
		Async:     false, // Síncrono para garantizar envío inmediato
	}

	return c.sendNotification(ctx, notificationReq)
}

// SendWelcomeEmail envía un email de bienvenida
func (c *NotificationHTTPClient) SendWelcomeEmail(ctx context.Context, email, name, companyName string) error {
	log.Printf("Sending welcome email to: %s, name: %s, company: %s", email, name, companyName)

	// Preparar datos para el template
	data := map[string]interface{}{
		"name":          name,
		"company":       companyName,
		"welcome_link":  "https://app.tiendavecina.com/dashboard",
		"support_email": "soporte@tiendavecina.com",
	}

	// Preparar request
	notificationReq := &port.SendEmailVerificationRequest{
		Type:      "email",
		Action:    "WELCOME",
		Recipient: email,
		Data:      data,
		Async:     true, // Asíncrono para welcome email
	}

	return c.sendNotification(ctx, notificationReq)
}

// sendNotification envía la notificación al servicio de notificaciones
func (c *NotificationHTTPClient) sendNotification(ctx context.Context, req *port.SendEmailVerificationRequest) error {
	url := fmt.Sprintf("%s/notifications", c.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshaling notification request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error making HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Si es exitoso (200), parsear como NotificationResponse
	if resp.StatusCode == http.StatusOK {
		var notificationResp port.NotificationResponse
		if err := json.NewDecoder(resp.Body).Decode(&notificationResp); err != nil {
			return fmt.Errorf("error decoding notification response: %w", err)
		}
		log.Printf("Email sent successfully. ID: %s, Status: %s", notificationResp.ID, notificationResp.Status)
		return nil
	}

	// Para errores, leer el body como string y devolver error descriptivo
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("notification service error (status: %d): failed to read response", resp.StatusCode)
	}

	return fmt.Errorf("notification service error (status: %d): %s", resp.StatusCode, string(bodyBytes))
}

// generateVerificationCode genera un código de verificación de 6 dígitos
func GenerateVerificationCode() string {
	rand.Seed(time.Now().UnixNano())
	code := rand.Intn(900000) + 100000 // Genera número entre 100000 y 999999
	return fmt.Sprintf("%06d", code)
}

// getNotificationEnv obtiene una variable de entorno con valor por defecto
func getNotificationEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
