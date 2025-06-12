package port

import (
	"context"
	"onboarding/src/onboarding/domain/entity"
	"pim/src/shared/domain/criteria"
)

// BusinesstypeRepository define los métodos para persistir Businesstype
type BusinesstypeRepository interface {
	Create(ctx context.Context, business_type *entity.Businesstype) error
	Update(ctx context.Context, business_type *entity.Businesstype) error
	FindByID(ctx context.Context, id string, tenantID string) (*entity.Businesstype, error)
	FindByTenant(ctx context.Context, tenantID string) ([]*entity.Businesstype, error)
	Delete(ctx context.Context, id string, tenantID string) error
	SearchByCriteria(ctx context.Context, crit criteria.Criteria) ([]*entity.Businesstype, error)
	CountByCriteria(ctx context.Context, crit criteria.Criteria) (int, error)
}
