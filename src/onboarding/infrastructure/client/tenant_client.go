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
	log.Printf("[TENANT_CLIENT] BootstrapTenantConfig - tenantID: %s", tenantID)

	if tenantID == "" {
		return fmt.Errorf("tenantID cannot be empty")
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	url := c.baseURL + "/api/v1/tenant/bootstrap"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return fmt.Errorf("error creando request HTTP: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Tenant-ID", tenantID)
	// TODO: En producción, agregar Authorization con service JWT

	client := &http.Client{Timeout: c.timeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[TENANT_CLIENT] Bootstrap request failed (best-effort, continuing): %v", err)
		return nil
	}
	defer resp.Body.Close()

	// Consumir body para reutilizar conexión
	io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("[TENANT_CLIENT] Bootstrap returned status %d (best-effort, continuing)", resp.StatusCode)
		return nil
	}

	log.Printf("[TENANT_CLIENT] Bootstrap success for tenant: %s", tenantID)
	return nil
}
