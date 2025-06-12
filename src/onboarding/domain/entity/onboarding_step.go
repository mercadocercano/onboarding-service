package entity

import (
	"fmt"
	"time"
	"github.com/google/uuid"
)

// Onboardingstep representa la entidad onboarding_step
type Onboardingstep struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewOnboardingstep crea una nueva instancia de Onboardingstep
func NewOnboardingstep(tenantID, name string) (*Onboardingstep, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id es requerido")
	}
	if name == "" {
		return nil, fmt.Errorf("name es requerido")
	}
	
	now := time.Now()
	return &Onboardingstep{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Name:      name,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Update actualiza los campos de la entidad
func (e *Onboardingstep) Update() {
	e.UpdatedAt = time.Now()
}
