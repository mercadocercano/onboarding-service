package port

import "context"

// TenantRegisteredEvent son los datos para el evento de bienvenida (welcome email)
// publicado al completar el onboarding.
type TenantRegisteredEvent struct {
	TenantID     string
	UserID       string
	Recipient    string // email destino
	Name         string
	Company      string
	BusinessType string
}

// EventPublisher publica eventos de dominio de onboarding al EventBus. Reemplaza la
// llamada HTTP sincrónica a notifications para el welcome email (Plan F1,
// ingestión event-driven). Es opcional: si no está cableado, el use case cae al HTTP.
type EventPublisher interface {
	PublishTenantRegistered(ctx context.Context, ev TenantRegisteredEvent) error
}
