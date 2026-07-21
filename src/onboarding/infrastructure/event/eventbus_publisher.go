package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/mercadocercano/eventbus"

	"onboarding/src/onboarding/domain/port"
)

// EventBusPublisher implementa port.EventPublisher sobre el EventBus PostgreSQL.
// Publica `onboarding.tenant.registered`, consumido por notifications.
type EventBusPublisher struct {
	publish   *eventbus.PublishEventUseCase
	namespace string
}

// NewEventBusPublisher crea el publisher. namespace vacío → 'mc'.
func NewEventBusPublisher(publish *eventbus.PublishEventUseCase, namespace string) *EventBusPublisher {
	if namespace == "" {
		namespace = "mc"
	}
	return &EventBusPublisher{publish: publish, namespace: namespace}
}

// PublishTenantRegistered arma el payload del welcome email y lo publica. El productor
// manda parámetros del template (no HTML); el consumer (notification) los reenvía al
// template. Paridad 1:1 con el data map del antiguo NotificationClient.SendWelcomeEmail.
func (p *EventBusPublisher) PublishTenantRegistered(ctx context.Context, ev port.TenantRegisteredEvent) error {
	name := ev.Name
	if name == "" {
		name = "Usuario"
	}
	company := ev.Company
	if company == "" {
		company = "Tu Empresa"
	}

	data := map[string]interface{}{
		"name":          name,
		"company":       company,
		"business_type": businessTypeDisplay(ev.BusinessType),
		"welcome_link":  "https://backoffice.mercadocercano.com/dashboard",
		"dashboard_url": "https://backoffice.mercadocercano.com/dashboard",
		"support_url":   "https://mercadocercano.com/soporte",
		"contact_email": "soporte@mercadocercano.com",
		"current_year":  time.Now().Year(),
	}

	payload, err := json.Marshal(map[string]interface{}{
		"namespace": p.namespace,
		"tenant_id": ev.TenantID,
		"user_id":   ev.UserID,
		"type":      "email",
		"action":    "WELCOME",
		"recipient": ev.Recipient,
		"data":      data,
	})
	if err != nil {
		return err
	}

	return p.publish.Execute(ctx, ev.TenantID, "tenant", "onboarding.tenant.registered", payload, "onboarding-service")
}

// businessTypeDisplay convierte el business_type a etiqueta legible (espejo del helper del
// antiguo NotificationClient, para que el welcome email diga "Comercio Minorista" y no "retail").
func businessTypeDisplay(businessType string) string {
	m := map[string]string{
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
	if display, ok := m[businessType]; ok {
		return display
	}
	return businessType
}
