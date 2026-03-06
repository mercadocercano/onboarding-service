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
	"time"

	"onboarding/src/onboarding/domain/port"
)

type NotificationClient struct {
	baseURL string
	timeout time.Duration
}

// NewNotificationClient crea una nueva instancia del cliente de notificaciones
func NewNotificationClient(baseURL string) port.NotificationClient {
	return &NotificationClient{
		baseURL: baseURL,
		timeout: 30 * time.Second,
	}
}

// GenerateVerificationCode genera un código de verificación de 6 dígitos
func GenerateVerificationCode() string {
	// Modo de testing: usar código fijo para tests automatizados
	// Variable de entorno: TESTING_MODE=true o BYPASS_EMAIL_VERIFICATION=true
	testingMode := getEnv("TESTING_MODE", "false")
	bypassVerification := getEnv("BYPASS_EMAIL_VERIFICATION", "false")
	
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
	log.Printf("Sending email verification to: %s, name: %s, code: %s", email, name, verificationCode)

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
	log.Printf("=== NOTIFICATION CLIENT: SendWelcomeEmail START ===")
	log.Printf("Parameters received:")
	log.Printf("  - Email: %s", email)
	log.Printf("  - User Name: %s", userName)
	log.Printf("  - Company Name: %s", companyName)
	log.Printf("  - Business Type: %s", businessType)
	log.Printf("  - Base URL: %s", c.baseURL)
	log.Printf("  - Timeout: %v", c.timeout)

	// Validar parámetros
	if email == "" {
		log.Printf("ERROR: Email parameter is empty")
		return fmt.Errorf("email cannot be empty")
	}
	if userName == "" {
		log.Printf("WARNING: User name is empty, using default")
		userName = "Usuario"
	}
	if companyName == "" {
		log.Printf("WARNING: Company name is empty, using default")
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

	log.Printf("Template data prepared:")
	log.Printf("  - name: %s", data["name"])
	log.Printf("  - company: %s", data["company"])
	log.Printf("  - business_type: %s", data["business_type"])
	log.Printf("  - welcome_link: %s", data["welcome_link"])
	log.Printf("  - current_year: %v", data["current_year"])

	request := &port.NotificationRequest{
		Type:      "email",
		Action:    "WELCOME",
		Recipient: email,
		Data:      data,
		Async:     false, // Enviar de forma síncrona para garantizar entrega inmediata
	}

	log.Printf("Notification request created:")
	log.Printf("  - Type: %s", request.Type)
	log.Printf("  - Action: %s", request.Action)
	log.Printf("  - Recipient: %s", request.Recipient)
	log.Printf("  - Async: %t", request.Async)

	log.Printf("Calling SendNotification...")
	response, err := c.SendNotification(ctx, request)
	if err != nil {
		log.Printf("ERROR: SendNotification failed")
		log.Printf("  - Error type: %T", err)
		log.Printf("  - Error message: %v", err)
		log.Printf("=== NOTIFICATION CLIENT: SendWelcomeEmail END (ERROR) ===")
		return fmt.Errorf("error enviando correo de bienvenida: %w", err)
	}

	log.Printf("SUCCESS: SendNotification completed")
	if response != nil {
		log.Printf("Response received:")
		log.Printf("  - Success: %t", response.Success)
		log.Printf("  - Message: %s", response.Message)
		log.Printf("  - NotificationID: %s", response.NotificationID)
		log.Printf("  - Status: %s", response.Status)
	} else {
		log.Printf("  - Response is nil")
	}

	log.Printf("=== NOTIFICATION CLIENT: SendWelcomeEmail END (SUCCESS) ===")
	return nil
}

// SendNotification envía una notificación genérica
func (c *NotificationClient) SendNotification(ctx context.Context, req *port.NotificationRequest) (*port.NotificationResponse, error) {
	log.Printf("=== HTTP CLIENT: SendNotification START ===")

	// Preparar el request HTTP
	log.Printf("Serializing request to JSON...")
	jsonData, err := json.Marshal(req)
	if err != nil {
		log.Printf("ERROR: Failed to serialize request to JSON: %v", err)
		return nil, fmt.Errorf("error serializando request: %w", err)
	}

	log.Printf("Request payload: %s", string(jsonData))
	log.Printf("Request size: %d bytes", len(jsonData))

	// Crear el contexto con timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Crear la request HTTP
	url := c.baseURL + "/notifications"
	log.Printf("Creating HTTP request:")
	log.Printf("  - Method: POST")
	log.Printf("  - URL: %s", url)
	log.Printf("  - Timeout: %v", c.timeout)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("ERROR: Failed to create HTTP request: %v", err)
		return nil, fmt.Errorf("error creando request HTTP: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	log.Printf("Headers set: Content-Type=application/json")

	// Ejecutar la request
	log.Printf("Executing HTTP request...")
	client := &http.Client{Timeout: c.timeout}

	startTime := time.Now()
	resp, err := client.Do(httpReq)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("ERROR: HTTP request failed after %v", duration)
		log.Printf("  - Error type: %T", err)
		log.Printf("  - Error message: %v", err)

		log.Printf("=== HTTP CLIENT: SendNotification END (REQUEST ERROR) ===")
		return nil, fmt.Errorf("error ejecutando request HTTP: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("HTTP request completed in %v", duration)
	log.Printf("Response received:")
	log.Printf("  - Status Code: %d", resp.StatusCode)
	log.Printf("  - Status: %s", resp.Status)
	log.Printf("  - Headers: %+v", resp.Header)

	// Leer la respuesta
	log.Printf("Reading response body...")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read response body: %v", err)
		return nil, fmt.Errorf("error leyendo respuesta: %w", err)
	}

	log.Printf("Response body read:")
	log.Printf("  - Size: %d bytes", len(body))
	log.Printf("  - Content: %s", string(body))

	// Verificar el status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("ERROR: Non-success status code received")
		log.Printf("  - Status Code: %d", resp.StatusCode)
		log.Printf("  - Response Body: %s", string(body))
		log.Printf("=== HTTP CLIENT: SendNotification END (STATUS ERROR) ===")
		return nil, fmt.Errorf("notification service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parsear la respuesta
	log.Printf("Parsing response JSON...")
	var response port.NotificationResponse
	if err := json.Unmarshal(body, &response); err != nil {
		log.Printf("WARNING: Failed to parse as NotificationResponse, trying generic parse...")
		log.Printf("  - Parse error: %v", err)

		// Si no podemos parsear como NotificationResponse, intentar como respuesta de éxito genérica
		var genericResponse map[string]interface{}
		if err2 := json.Unmarshal(body, &genericResponse); err2 == nil {
			log.Printf("SUCCESS: Parsed as generic response: %+v", genericResponse)
			response = port.NotificationResponse{
				Success:        true,
				Message:        "Notification sent successfully",
				NotificationID: fmt.Sprintf("%v", genericResponse["id"]),
				Status:         "sent",
			}
		} else {
			log.Printf("ERROR: Failed to parse response as any JSON format")
			log.Printf("  - Generic parse error: %v", err2)
			log.Printf("  - Raw response: %s", string(body))
			log.Printf("=== HTTP CLIENT: SendNotification END (PARSE ERROR) ===")
			return nil, fmt.Errorf("error parseando respuesta JSON: %w", err)
		}
	} else {
		log.Printf("SUCCESS: Parsed as NotificationResponse: %+v", response)
	}

	log.Printf("=== HTTP CLIENT: SendNotification END (SUCCESS) ===")
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
