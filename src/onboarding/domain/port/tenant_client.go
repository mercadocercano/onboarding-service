package port

import (
	"context"
)

// TenantClient define el contrato para el cliente del tenant-service
// Este puerto permite la integración con tenant-service sin acoplar dominios
type TenantClient interface {
	// BootstrapTenantConfig inicializa la configuración default de un tenant
	// Esta operación es idempotente y best-effort (no debe bloquear el onboarding)
	BootstrapTenantConfig(ctx context.Context, tenantID string) error
}
