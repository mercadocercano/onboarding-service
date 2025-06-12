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

// OnboardingstepPostgresRepository implementa el repositorio usando PostgreSQL
type OnboardingstepPostgresRepository struct {
	db        *sql.DB
	converter *sharedCriteria.SQLCriteriaConverter
}

// NewOnboardingstepPostgresRepository crea una nueva instancia del repositorio
func NewOnboardingstepPostgresRepository(db *sql.DB) *OnboardingstepPostgresRepository {
	return &OnboardingstepPostgresRepository{
		db:        db,
		converter: sharedCriteria.NewSQLCriteriaConverter(),
	}
}

// Create crea un nuevo onboarding_step
func (r *OnboardingstepPostgresRepository) Create(ctx context.Context, onboarding_step *entity.Onboardingstep) error {
	query := `
		INSERT INTO onboarding_steps (
			id, tenant_id, name, active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		onboarding_step.ID,
		onboarding_step.TenantID,
		onboarding_step.Name,
		onboarding_step.Active,
		onboarding_step.CreatedAt,
		onboarding_step.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error creando onboarding_step: %v", err)
		return fmt.Errorf("%w: %v", exception.ErrOnboardingstepCreateFailed, err)
	}

	return nil
}

// Update actualiza un onboarding_step existente
func (r *OnboardingstepPostgresRepository) Update(ctx context.Context, onboarding_step *entity.Onboardingstep) error {
	query := `
		UPDATE onboarding_steps SET
			name = $3,
			active = $4,
			updated_at = $5
		WHERE id = $1 AND tenant_id = $2
	`

	result, err := r.db.ExecContext(ctx, query,
		onboarding_step.ID,
		onboarding_step.TenantID,
		onboarding_step.Name,
		onboarding_step.Active,
		onboarding_step.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error actualizando onboarding_step: %v", err)
		return fmt.Errorf("%w: %v", exception.ErrOnboardingstepUpdateFailed, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return exception.ErrOnboardingstepNotFound
	}

	return nil
}

// FindByID busca un onboarding_step por su ID
func (r *OnboardingstepPostgresRepository) FindByID(ctx context.Context, id string, tenantID string) (*entity.Onboardingstep, error) {
	query := `
		SELECT id, tenant_id, name, active, created_at, updated_at
		FROM onboarding_steps 
		WHERE id = $1 AND tenant_id = $2
	`

	row := r.db.QueryRowContext(ctx, query, id, tenantID)
	return r.scanOnboardingstep(row)
}

// FindByTenant obtiene todos los onboarding_steps de un tenant
func (r *OnboardingstepPostgresRepository) FindByTenant(ctx context.Context, tenantID string) ([]*entity.Onboardingstep, error) {
	query := `
		SELECT id, tenant_id, name, active, created_at, updated_at
		FROM onboarding_steps 
		WHERE tenant_id = $1 AND active = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanOnboardingsteps(rows)
}

// Delete elimina un onboarding_step
func (r *OnboardingstepPostgresRepository) Delete(ctx context.Context, id string, tenantID string) error {
	query := `DELETE FROM onboarding_steps WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		log.Printf("Error eliminando onboarding_step: %v", err)
		return fmt.Errorf("%w: %v", exception.ErrOnboardingstepDeleteFailed, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return exception.ErrOnboardingstepNotFound
	}

	return nil
}

// SearchByCriteria busca onboarding_steps usando criterios
func (r *OnboardingstepPostgresRepository) SearchByCriteria(ctx context.Context, crit criteria.Criteria) ([]*entity.Onboardingstep, error) {
	baseQuery := `
		SELECT id, tenant_id, name, active, created_at, updated_at
		FROM onboarding_steps
	`

	query, params := r.converter.ToSelectSQL(baseQuery, crit)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanOnboardingsteps(rows)
}

// CountByCriteria cuenta onboarding_steps usando criterios
func (r *OnboardingstepPostgresRepository) CountByCriteria(ctx context.Context, crit criteria.Criteria) (int, error) {
	baseCountQuery := "SELECT COUNT(*) FROM onboarding_steps"

	query, params := r.converter.ToCountSQL(baseCountQuery, crit)

	var count int
	err := r.db.QueryRowContext(ctx, query, params...).Scan(&count)
	return count, err
}

// scanOnboardingstep escanea una fila y devuelve un onboarding_step
func (r *OnboardingstepPostgresRepository) scanOnboardingstep(row *sql.Row) (*entity.Onboardingstep, error) {
	var onboarding_step entity.Onboardingstep

	err := row.Scan(
		&onboarding_step.ID,
		&onboarding_step.TenantID,
		&onboarding_step.Name,
		&onboarding_step.Active,
		&onboarding_step.CreatedAt,
		&onboarding_step.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &onboarding_step, nil
}

// scanOnboardingsteps escanea múltiples filas y devuelve una lista de onboarding_steps
func (r *OnboardingstepPostgresRepository) scanOnboardingsteps(rows *sql.Rows) ([]*entity.Onboardingstep, error) {
	var onboarding_steps []*entity.Onboardingstep

	for rows.Next() {
		var onboarding_step entity.Onboardingstep

		err := rows.Scan(
			&onboarding_step.ID,
			&onboarding_step.TenantID,
			&onboarding_step.Name,
			&onboarding_step.Active,
			&onboarding_step.CreatedAt,
			&onboarding_step.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		onboarding_steps = append(onboarding_steps, &onboarding_step)
	}

	return onboarding_steps, nil
}
