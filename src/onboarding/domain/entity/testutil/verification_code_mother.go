package testutil

import (
	"time"

	"onboarding/src/onboarding/domain/entity"

	"github.com/google/uuid"
)

// VerificationCodeMother construye instancias de VerificationCode para tests.
type VerificationCodeMother struct {
	code *entity.VerificationCode
}

// NewVerificationCodeMother crea un builder con valores por defecto válidos.
func NewVerificationCodeMother() *VerificationCodeMother {
	now := time.Now()
	return &VerificationCodeMother{
		code: &entity.VerificationCode{
			ID:        uuid.New(),
			ProcessID: uuid.New(),
			TenantID:  uuid.New(),
			UserEmail: "test@example.com",
			Code:      "123456",
			IsUsed:    false,
			ExpiresAt: now.Add(15 * time.Minute),
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}

func (m *VerificationCodeMother) WithID(id uuid.UUID) *VerificationCodeMother {
	m.code.ID = id
	return m
}

func (m *VerificationCodeMother) WithProcessID(id uuid.UUID) *VerificationCodeMother {
	m.code.ProcessID = id
	return m
}

func (m *VerificationCodeMother) WithTenantID(id uuid.UUID) *VerificationCodeMother {
	m.code.TenantID = id
	return m
}

func (m *VerificationCodeMother) WithEmail(email string) *VerificationCodeMother {
	m.code.UserEmail = email
	return m
}

func (m *VerificationCodeMother) WithCode(code string) *VerificationCodeMother {
	m.code.Code = code
	return m
}

func (m *VerificationCodeMother) WithUsed(used bool) *VerificationCodeMother {
	m.code.IsUsed = used
	return m
}

func (m *VerificationCodeMother) WithExpired() *VerificationCodeMother {
	m.code.ExpiresAt = time.Now().Add(-1 * time.Minute)
	return m
}

func (m *VerificationCodeMother) WithExpiresAt(t time.Time) *VerificationCodeMother {
	m.code.ExpiresAt = t
	return m
}

func (m *VerificationCodeMother) WithCreatedAt(t time.Time) *VerificationCodeMother {
	m.code.CreatedAt = t
	return m
}

func (m *VerificationCodeMother) Build() *entity.VerificationCode {
	return m.code
}
