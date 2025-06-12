package entity

import (
	"fmt"
	"time"
	"github.com/google/uuid"
)

// Tenantonboarding representa la entidad tenant_onboarding
type Tenantonboarding struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewTenantonboarding crea una nueva instancia de Tenantonboarding
func NewTenantonboarding(tenantID, name string) (*Tenantonboarding, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id es requerido")
	}
	if name == "" {
		return nil, fmt.Errorf("name es requerido")
	}
	
	now := time.Now()
	return &Tenantonboarding{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Name:      name,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Update actualiza los campos de la entidad
func (e *Tenantonboarding) Update() {
	e.UpdatedAt = time.Now()
}
