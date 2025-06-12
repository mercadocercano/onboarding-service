package port

import (
	"context"
	"onboarding/src/onboarding/domain/entity"
	"pim/src/shared/domain/criteria"
)

// OnboardingstepRepository define los métodos para persistir Onboardingstep
type OnboardingstepRepository interface {
	Create(ctx context.Context, onboarding_step *entity.Onboardingstep) error
	Update(ctx context.Context, onboarding_step *entity.Onboardingstep) error
	FindByID(ctx context.Context, id string, tenantID string) (*entity.Onboardingstep, error)
	FindByTenant(ctx context.Context, tenantID string) ([]*entity.Onboardingstep, error)
	Delete(ctx context.Context, id string, tenantID string) error
	SearchByCriteria(ctx context.Context, crit criteria.Criteria) ([]*entity.Onboardingstep, error)
	CountByCriteria(ctx context.Context, crit criteria.Criteria) (int, error)
}
