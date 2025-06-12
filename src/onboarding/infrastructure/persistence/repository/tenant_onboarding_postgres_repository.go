package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"onboarding/src/onboarding/domain/entity"
	"onboarding/src/onboarding/domain/exception"
	"pim/src/shared/domain/criteria"
	sharedCriteria "pim/src/shared/infrastructure/criteria"
)

// TenantonboardingPostgresRepository implementa el repositorio usando PostgreSQL
type TenantonboardingPostgresRepository struct {
	db        *sql.DB
	converter *sharedCriteria.SQLCriteriaConverter
}

// NewTenantonboardingPostgresRepository crea una nueva instancia del repositorio
func NewTenantonboardingPostgresRepository(db *sql.DB) *TenantonboardingPostgresRepository {
	return &TenantonboardingPostgresRepository{
		db:        db,
		converter: sharedCriteria.NewSQLCriteriaConverter(),
	}
}

// Create crea un nuevo tenant_onboarding
func (r *TenantonboardingPostgresRepository) Create(ctx context.Context, tenant_onboarding *entity.Tenantonboarding) error {
	query := `
		INSERT INTO tenant_onboardings (
			id, tenant_id, name, active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		tenant_onboarding.ID,
		tenant_onboarding.TenantID,
		tenant_onboarding.Name,
		tenant_onboarding.Active,
		tenant_onboarding.CreatedAt,
		tenant_onboarding.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error creando tenant_onboarding: %v", err)
		return fmt.Errorf("%w: %v", exception.ErrTenantonboardingCreateFailed, err)
	}

	return nil
}

// Update actualiza un tenant_onboarding existente
func (r *TenantonboardingPostgresRepository) Update(ctx context.Context, tenant_onboarding *entity.Tenantonboarding) error {
	query := `
		UPDATE tenant_onboardings SET
			name = $3,
			active = $4,
			updated_at = $5
		WHERE id = $1 AND tenant_id = $2
	`

	result, err := r.db.ExecContext(ctx, query,
		tenant_onboarding.ID,
		tenant_onboarding.TenantID,
		tenant_onboarding.Name,
		tenant_onboarding.Active,
		tenant_onboarding.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error actualizando tenant_onboarding: %v", err)
		return fmt.Errorf("%w: %v", exception.ErrTenantonboardingUpdateFailed, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return exception.ErrTenantonboardingNotFound
	}

	return nil
}

// FindByID busca un tenant_onboarding por su ID
func (r *TenantonboardingPostgresRepository) FindByID(ctx context.Context, id string, tenantID string) (*entity.Tenantonboarding, error) {
	query := `
		SELECT id, tenant_id, name, active, created_at, updated_at
		FROM tenant_onboardings 
		WHERE id = $1 AND tenant_id = $2
	`

	row := r.db.QueryRowContext(ctx, query, id, tenantID)
	return r.scanTenantonboarding(row)
}

// FindByTenant obtiene todos los tenant_onboardings de un tenant
func (r *TenantonboardingPostgresRepository) FindByTenant(ctx context.Context, tenantID string) ([]*entity.Tenantonboarding, error) {
	query := `
		SELECT id, tenant_id, name, active, created_at, updated_at
		FROM tenant_onboardings 
		WHERE tenant_id = $1 AND active = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTenantonboardings(rows)
}

// Delete elimina un tenant_onboarding
func (r *TenantonboardingPostgresRepository) Delete(ctx context.Context, id string, tenantID string) error {
	query := `DELETE FROM tenant_onboardings WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		log.Printf("Error eliminando tenant_onboarding: %v", err)
		return fmt.Errorf("%w: %v", exception.ErrTenantonboardingDeleteFailed, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return exception.ErrTenantonboardingNotFound
	}

	return nil
}

// SearchByCriteria busca tenant_onboardings usando criterios
func (r *TenantonboardingPostgresRepository) SearchByCriteria(ctx context.Context, crit criteria.Criteria) ([]*entity.Tenantonboarding, error) {
	baseQuery := `
		SELECT id, tenant_id, name, active, created_at, updated_at
		FROM tenant_onboardings
	`

	query, params := r.converter.ToSelectSQL(baseQuery, crit)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanTenantonboardings(rows)
}

// CountByCriteria cuenta tenant_onboardings usando criterios
func (r *TenantonboardingPostgresRepository) CountByCriteria(ctx context.Context, crit criteria.Criteria) (int, error) {
	baseCountQuery := "SELECT COUNT(*) FROM tenant_onboardings"

	query, params := r.converter.ToCountSQL(baseCountQuery, crit)

	var count int
	err := r.db.QueryRowContext(ctx, query, params...).Scan(&count)
	return count, err
}

// scanTenantonboarding escanea una fila y devuelve un tenant_onboarding
func (r *TenantonboardingPostgresRepository) scanTenantonboarding(row *sql.Row) (*entity.Tenantonboarding, error) {
	var tenant_onboarding entity.Tenantonboarding

	err := row.Scan(
		&tenant_onboarding.ID,
		&tenant_onboarding.TenantID,
		&tenant_onboarding.Name,
		&tenant_onboarding.Active,
		&tenant_onboarding.CreatedAt,
		&tenant_onboarding.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &tenant_onboarding, nil
}

// scanTenantonboardings escanea múltiples filas y devuelve una lista de tenant_onboardings
func (r *TenantonboardingPostgresRepository) scanTenantonboardings(rows *sql.Rows) ([]*entity.Tenantonboarding, error) {
	var tenant_onboardings []*entity.Tenantonboarding

	for rows.Next() {
		var tenant_onboarding entity.Tenantonboarding

		err := rows.Scan(
			&tenant_onboarding.ID,
			&tenant_onboarding.TenantID,
			&tenant_onboarding.Name,
			&tenant_onboarding.Active,
			&tenant_onboarding.CreatedAt,
			&tenant_onboarding.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		tenant_onboardings = append(tenant_onboardings, &tenant_onboarding)
	}

	return tenant_onboardings, nil
}
