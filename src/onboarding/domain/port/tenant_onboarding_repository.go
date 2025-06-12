package port

import (
	"context"
	"onboarding/src/onboarding/domain/entity"
	"pim/src/shared/domain/criteria"
)

// TenantonboardingRepository define los métodos para persistir Tenantonboarding
type TenantonboardingRepository interface {
	Create(ctx context.Context, tenant_onboarding *entity.Tenantonboarding) error
	Update(ctx context.Context, tenant_onboarding *entity.Tenantonboarding) error
	FindByID(ctx context.Context, id string, tenantID string) (*entity.Tenantonboarding, error)
	FindByTenant(ctx context.Context, tenantID string) ([]*entity.Tenantonboarding, error)
	Delete(ctx context.Context, id string, tenantID string) error
	SearchByCriteria(ctx context.Context, crit criteria.Criteria) ([]*entity.Tenantonboarding, error)
	CountByCriteria(ctx context.Context, crit criteria.Criteria) (int, error)
}
