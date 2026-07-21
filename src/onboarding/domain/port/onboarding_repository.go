package port

import (
	"context"

	"onboarding/src/onboarding/domain/entity"

	"github.com/google/uuid"
)

// OnboardingRepository define las operaciones de persistencia para el onboarding.
//
// RLS (E28, RULE-09/RULE-10): toda operación sobre `onboarding_processes` y
// `verification_codes` corre bajo la policy `tenant_isolation` — el tenant sale de la
// entidad (creada server-side vía IAM) o, para el lookup pre-auth por process_id, del
// resolver two-step D1 en la implementación. `ctx` se hila desde el request de Gin;
// nunca context.Background() en un path de request.
type OnboardingRepository interface {
	// Process operations
	SaveProcess(ctx context.Context, process *entity.OnboardingProcess) error
	GetProcessByID(ctx context.Context, id uuid.UUID) (*entity.OnboardingProcess, error)
	UpdateProcess(ctx context.Context, process *entity.OnboardingProcess) error

	// Step definitions operations (catálogo global sin RLS, seed en migración 002 —
	// legible pre-auth por diseño)
	GetStepDefinitions(ctx context.Context) ([]*entity.StepDefinition, error)
	GetStepDefinitionByNumber(ctx context.Context, stepNumber int) (*entity.StepDefinition, error)

	// Verification code operations — el tenant viene de code.TenantID (escrituras) o
	// explícito en la firma (lectura: el caller ya cargó el proceso y tiene el tenant)
	SaveVerificationCode(ctx context.Context, code *entity.VerificationCode) error
	GetVerificationCodeByProcessID(ctx context.Context, tenantID, processID uuid.UUID) (*entity.VerificationCode, error)
	UpdateVerificationCode(ctx context.Context, code *entity.VerificationCode) error
}
