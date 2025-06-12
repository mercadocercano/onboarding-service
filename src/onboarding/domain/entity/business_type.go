package entity

import (
	"fmt"
	"time"
	"github.com/google/uuid"
)

// Businesstype representa la entidad business_type
type Businesstype struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewBusinesstype crea una nueva instancia de Businesstype
func NewBusinesstype(tenantID, name string) (*Businesstype, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id es requerido")
	}
	if name == "" {
		return nil, fmt.Errorf("name es requerido")
	}
	
	now := time.Now()
	return &Businesstype{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Name:      name,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Update actualiza los campos de la entidad
func (e *Businesstype) Update() {
	e.UpdatedAt = time.Now()
}
