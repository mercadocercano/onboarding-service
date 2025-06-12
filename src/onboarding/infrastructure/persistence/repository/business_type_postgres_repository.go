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

// BusinesstypePostgresRepository implementa el repositorio usando PostgreSQL
type BusinesstypePostgresRepository struct {
	db        *sql.DB
	converter *sharedCriteria.SQLCriteriaConverter
}

// NewBusinesstypePostgresRepository crea una nueva instancia del repositorio
func NewBusinesstypePostgresRepository(db *sql.DB) *BusinesstypePostgresRepository {
	return &BusinesstypePostgresRepository{
		db:        db,
		converter: sharedCriteria.NewSQLCriteriaConverter(),
	}
}

// Create crea un nuevo business_type
func (r *BusinesstypePostgresRepository) Create(ctx context.Context, business_type *entity.Businesstype) error {
	query := `
		INSERT INTO business_types (
			id, tenant_id, name, active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6
		)
	`

	_, err := r.db.ExecContext(ctx, query,
		business_type.ID,
		business_type.TenantID,
		business_type.Name,
		business_type.Active,
		business_type.CreatedAt,
		business_type.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error creando business_type: %v", err)
		return fmt.Errorf("%w: %v", exception.ErrBusinesstypeCreateFailed, err)
	}

	return nil
}

// Update actualiza un business_type existente
func (r *BusinesstypePostgresRepository) Update(ctx context.Context, business_type *entity.Businesstype) error {
	query := `
		UPDATE business_types SET
			name = $3,
			active = $4,
			updated_at = $5
		WHERE id = $1 AND tenant_id = $2
	`

	result, err := r.db.ExecContext(ctx, query,
		business_type.ID,
		business_type.TenantID,
		business_type.Name,
		business_type.Active,
		business_type.UpdatedAt,
	)

	if err != nil {
		log.Printf("Error actualizando business_type: %v", err)
		return fmt.Errorf("%w: %v", exception.ErrBusinesstypeUpdateFailed, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return exception.ErrBusinesstypeNotFound
	}

	return nil
}

// FindByID busca un business_type por su ID
func (r *BusinesstypePostgresRepository) FindByID(ctx context.Context, id string, tenantID string) (*entity.Businesstype, error) {
	query := `
		SELECT id, tenant_id, name, active, created_at, updated_at
		FROM business_types 
		WHERE id = $1 AND tenant_id = $2
	`

	row := r.db.QueryRowContext(ctx, query, id, tenantID)
	return r.scanBusinesstype(row)
}

// FindByTenant obtiene todos los business_types de un tenant
func (r *BusinesstypePostgresRepository) FindByTenant(ctx context.Context, tenantID string) ([]*entity.Businesstype, error) {
	query := `
		SELECT id, tenant_id, name, active, created_at, updated_at
		FROM business_types 
		WHERE tenant_id = $1 AND active = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanBusinesstypes(rows)
}

// Delete elimina un business_type
func (r *BusinesstypePostgresRepository) Delete(ctx context.Context, id string, tenantID string) error {
	query := `DELETE FROM business_types WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		log.Printf("Error eliminando business_type: %v", err)
		return fmt.Errorf("%w: %v", exception.ErrBusinesstypeDeleteFailed, err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return exception.ErrBusinesstypeNotFound
	}

	return nil
}

// SearchByCriteria busca business_types usando criterios
func (r *BusinesstypePostgresRepository) SearchByCriteria(ctx context.Context, crit criteria.Criteria) ([]*entity.Businesstype, error) {
	baseQuery := `
		SELECT id, tenant_id, name, active, created_at, updated_at
		FROM business_types
	`

	query, params := r.converter.ToSelectSQL(baseQuery, crit)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanBusinesstypes(rows)
}

// CountByCriteria cuenta business_types usando criterios
func (r *BusinesstypePostgresRepository) CountByCriteria(ctx context.Context, crit criteria.Criteria) (int, error) {
	baseCountQuery := "SELECT COUNT(*) FROM business_types"

	query, params := r.converter.ToCountSQL(baseCountQuery, crit)

	var count int
	err := r.db.QueryRowContext(ctx, query, params...).Scan(&count)
	return count, err
}

// scanBusinesstype escanea una fila y devuelve un business_type
func (r *BusinesstypePostgresRepository) scanBusinesstype(row *sql.Row) (*entity.Businesstype, error) {
	var business_type entity.Businesstype

	err := row.Scan(
		&business_type.ID,
		&business_type.TenantID,
		&business_type.Name,
		&business_type.Active,
		&business_type.CreatedAt,
		&business_type.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &business_type, nil
}

// scanBusinesstypes escanea múltiples filas y devuelve una lista de business_types
func (r *BusinesstypePostgresRepository) scanBusinesstypes(rows *sql.Rows) ([]*entity.Businesstype, error) {
	var business_types []*entity.Businesstype

	for rows.Next() {
		var business_type entity.Businesstype

		err := rows.Scan(
			&business_type.ID,
			&business_type.TenantID,
			&business_type.Name,
			&business_type.Active,
			&business_type.CreatedAt,
			&business_type.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		business_types = append(business_types, &business_type)
	}

	return business_types, nil
}
