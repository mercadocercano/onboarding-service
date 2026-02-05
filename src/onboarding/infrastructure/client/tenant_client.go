package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"onboarding/src/onboarding/domain/port"
)

// TenantClient implementa la comunicación HTTP con tenant-service
type TenantClient struct {
	baseURL string
	timeout time.Duration
}

// NewTenantClient crea una nueva instancia del cliente de tenant-service
func NewTenantClient(baseURL string) port.TenantClient {
	return &TenantClient{
		baseURL: baseURL,
		timeout: 10 * time.Second, // Timeout corto para best-effort
	}
}

// BootstrapTenantConfig inicializa la configuración default de un tenant
// Esta operación es best-effort: no falla el onboarding si hay error
func (c *TenantClient) BootstrapTenantConfig(ctx context.Context, tenantID string) error {
	log.Printf("=== TENANT CLIENT: BootstrapTenantConfig START ===")
	log.Printf("Parameters:")
	log.Printf("  - TenantID: %s", tenantID)
	log.Printf("  - Base URL: %s", c.baseURL)

	// Validar parámetros
	if tenantID == "" {
		log.Printf("ERROR: TenantID is empty")
		return fmt.Errorf("tenantID cannot be empty")
	}

	// Crear el contexto con timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Preparar la request HTTP
	url := c.baseURL + "/api/v1/tenant/bootstrap"
	log.Printf("Creating HTTP request:")
	log.Printf("  - Method: POST")
	log.Printf("  - URL: %s", url)
	log.Printf("  - Timeout: %v", c.timeout)

	// Body vacío para POST
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		log.Printf("ERROR: Failed to create HTTP request: %v", err)
		return fmt.Errorf("error creando request HTTP: %w", err)
	}

	// Headers requeridos
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Tenant-ID", tenantID)
	// TODO: En producción, agregar Authorization con service JWT
	// httpReq.Header.Set("Authorization", "Bearer "+serviceToken)

	log.Printf("Headers set:")
	log.Printf("  - Content-Type: application/json")
	log.Printf("  - X-Tenant-ID: %s", tenantID)

	// Ejecutar la request
	log.Printf("Executing HTTP request...")
	client := &http.Client{Timeout: c.timeout}

	startTime := time.Now()
	resp, err := client.Do(httpReq)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("WARNING: HTTP request failed after %v (best-effort, continuing)", duration)
		log.Printf("  - Error type: %T", err)
		log.Printf("  - Error message: %v", err)
		log.Printf("=== TENANT CLIENT: BootstrapTenantConfig END (REQUEST ERROR - IGNORED) ===")
		// Best-effort: no retornamos error para no bloquear onboarding
		return nil
	}
	defer resp.Body.Close()

	log.Printf("HTTP request completed in %v", duration)
	log.Printf("Response received:")
	log.Printf("  - Status Code: %d", resp.StatusCode)
	log.Printf("  - Status: %s", resp.Status)

	// Leer la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("WARNING: Failed to read response body: %v (best-effort, continuing)", err)
		log.Printf("=== TENANT CLIENT: BootstrapTenantConfig END (READ ERROR - IGNORED) ===")
		return nil
	}

	log.Printf("Response body:")
	log.Printf("  - Size: %d bytes", len(body))
	log.Printf("  - Content: %s", string(body))

	// Verificar el status code
	// tenant-service siempre debe responder 200 (idempotente)
	if resp.StatusCode != http.StatusOK {
		log.Printf("WARNING: Non-200 status code received: %d (best-effort, continuing)", resp.StatusCode)
		log.Printf("  - Response Body: %s", string(body))
		log.Printf("=== TENANT CLIENT: BootstrapTenantConfig END (STATUS WARNING - IGNORED) ===")
		// Best-effort: no retornamos error
		return nil
	}

	log.Printf("SUCCESS: Tenant config bootstrapped successfully")
	log.Printf("=== TENANT CLIENT: BootstrapTenantConfig END (SUCCESS) ===")
	return nil
}
