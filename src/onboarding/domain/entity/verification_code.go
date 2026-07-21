package entity

import (
	"time"

	"github.com/google/uuid"
)

// VerificationCode representa un código de verificación de email.
// TenantID (E28/D3): scope RLS de la fila — derivado siempre del proceso ya cargado,
// nunca de input del cliente.
type VerificationCode struct {
	ID        uuid.UUID `json:"id"`
	ProcessID uuid.UUID `json:"process_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	UserEmail string    `json:"user_email"`
	Code      string    `json:"code"`
	IsUsed    bool      `json:"is_used"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewVerificationCode crea un nuevo código de verificación
func NewVerificationCode(tenantID, processID uuid.UUID, userEmail, code string) *VerificationCode {
	now := time.Now()

	return &VerificationCode{
		ID:        uuid.New(),
		ProcessID: processID,
		TenantID:  tenantID,
		UserEmail: userEmail,
		Code:      code,
		IsUsed:    false,
		ExpiresAt: now.Add(15 * time.Minute), // Expira en 15 minutos
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// IsExpired verifica si el código ha expirado
func (vc *VerificationCode) IsExpired() bool {
	return time.Now().After(vc.ExpiresAt)
}

// IsValid verifica si el código es válido (no usado y no expirado)
func (vc *VerificationCode) IsValid() bool {
	return !vc.IsUsed && !vc.IsExpired()
}

// MarkAsUsed marca el código como usado
func (vc *VerificationCode) MarkAsUsed() {
	vc.IsUsed = true
	vc.UpdatedAt = time.Now()
}
