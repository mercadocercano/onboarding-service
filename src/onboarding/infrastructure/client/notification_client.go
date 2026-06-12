package client

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/hornosg/go-shared/infrastructure/env"

	"onboarding/src/onboarding/domain/port"
)

type NotificationClient struct {
	baseURL string
	timeout time.Duration
	apiKey  string
}

// NewNotificationClient crea una nueva instancia del cliente de notificaciones con autenticación S2S
func NewNotificationClient(baseURL string) port.NotificationClient {
	apiKey := os.Getenv("S2S_API_KEY")

	return &NotificationClient{
		baseURL: baseURL,
		timeout: 30 * time.Second,
		apiKey:  apiKey,
	}
}

// GenerateVerificationCode genera un código de verificación de 6 dígitos
func GenerateVerificationCode() string {
	// Modo de testing: usar código fijo para tests automatizados
	// Variable de entorno: TESTING_MODE=true o BYPASS_EMAIL_VERIFICATION=true
	testingMode := env.Get("TESTING_MODE", "false")
	bypassVerification := env.Get("BYPASS_EMAIL_VERIFICATION", "false")

	if testingMode == "true" || bypassVerification == "true" {
		log.Printf("⚠️ TESTING MODE: Using fixed verification code 123456")
		return "123456"
	}

	// Generar un número aleatorio de 6 dígitos (100000-999999)
	min := int64(100000)
	max := int64(999999)

	// Usar crypto/rand para seguridad
	n, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		// Fallback a código por defecto si hay error
		log.Printf("⚠️ Error generating random code, using fallback: 123456")
		return "123456"
	}

	code := min + n.Int64()
	return fmt.Sprintf("%06d", code)
}

// SendEmailVerification envía un email de verificación con código
func (c *NotificationClient) SendEmailVerification(ctx context.Context, email, name, verificationCode string) error {
	log.Printf("Sending email verification")

	// Preparar datos para el template
	data := map[string]interface{}{
		"name":          name,
		"token":         verificationCode,
		"company":       "MercadoCercano",
		"expiry_time":   "15 minutos",
		"support_email": "soporte@mercadocercano.com",
	}

	// Preparar request
	notificationReq := &port.NotificationRequest{
		Type:      "email",
		Action:    "EMAIL_VERIFICATION",
		Recipient: email,
		Data:      data,
		Async:     false, // Síncrono para garantizar envío inmediato
	}

	_, err := c.SendNotification(ctx, notificationReq)
	return err
}

// SendWelcomeEmail envía un correo de bienvenida después de completar el onboarding
func (c *NotificationClient) SendWelcomeEmail(ctx context.Context, email, userName, companyName, businessType string) error {
	log.Printf("[NOTIFICATION] SendWelcomeEmail - businessType: %s", businessType)

	// Validar parámetros
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}
	if userName == "" {
		userName = "Usuario"
	}
	if companyName == "" {
		companyName = "Tu Empresa"
	}

	// Preparar los datos para la plantilla (nombres que coinciden con el template HTML)
	data := map[string]interface{}{
		"name":          userName,
		"company":       companyName,
		"business_type": getBusinessTypeDisplay(businessType),
		"welcome_link":  "https://backoffice.mercadocercano.com/dashboard",
		"dashboard_url": "https://backoffice.mercadocercano.com/dashboard",
		"support_url":   "https://mercadocercano.com/soporte",
		"contact_email": "soporte@mercadocercano.com",
		"current_year":  time.Now().Year(),
	}

	request := &port.NotificationRequest{
		Type:      "email",
		Action:    "WELCOME",
		Recipient: email,
		Data:      data,
		Async:     false, // Enviar de forma síncrona para garantizar entrega inmediata
	}

	response, err := c.SendNotification(ctx, request)
	if err != nil {
		log.Printf("[NOTIFICATION] SendWelcomeEmail failed: %v", err)
		return fmt.Errorf("error enviando correo de bienvenida: %w", err)
	}

	if response != nil {
		log.Printf("[NOTIFICATION] SendWelcomeEmail success - notificationID: %s", response.NotificationID)
	}
	return nil
}

// SendNotification envía una notificación genérica
func (c *NotificationClient) SendNotification(ctx context.Context, req *port.NotificationRequest) (*port.NotificationResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error serializando request: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	url := c.baseURL + "/notifications"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creando request HTTP: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	// Autenticación S2S via API Key
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	client := &http.Client{Timeout: c.timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[NOTIFICATION] HTTP request failed: %v", err)
		return nil, fmt.Errorf("error ejecutando request HTTP: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("[NOTIFICATION] Service returned status %d", resp.StatusCode)
		return nil, fmt.Errorf("notification service returned status %d: %s", resp.StatusCode, string(body))
	}

	var response port.NotificationResponse
	if err := json.Unmarshal(body, &response); err != nil {
		var genericResponse map[string]interface{}
		if err2 := json.Unmarshal(body, &genericResponse); err2 == nil {
			response = port.NotificationResponse{
				Success:        true,
				Message:        "Notification sent successfully",
				NotificationID: fmt.Sprintf("%v", genericResponse["id"]),
				Status:         "sent",
			}
		} else {
			return nil, fmt.Errorf("error parseando respuesta JSON: %w", err)
		}
	}

	return &response, nil
}

// getBusinessTypeDisplay convierte el business_type a un formato legible
func getBusinessTypeDisplay(businessType string) string {
	businessTypeMap := map[string]string{
		"retail":              "Comercio Minorista",
		"home-construction":   "Hogar y Construcción",
		"fashion":             "Moda y Vestimenta",
		"electronics":         "Electrónicos",
		"automotive":          "Automotriz",
		"food-beverage":       "Alimentos y Bebidas",
		"health-pharmacy":     "Salud y Farmacia",
		"sports-fitness":      "Deportes y Fitness",
		"beauty-cosmetics":    "Belleza y Cosméticos",
		"toys-games":          "Juguetes y Juegos",
		"pet-supplies":        "Mascotas",
		"office-supplies":     "Oficina y Papelería",
		"jewelry-accessories": "Joyería y Accesorios",
		"polirubro":           "Polirubro",
	}

	if display, exists := businessTypeMap[businessType]; exists {
		return display
	}
	return businessType
}
