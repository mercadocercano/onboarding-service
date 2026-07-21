package persistence

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/hornosg/go-shared/infrastructure/postgres"

	"onboarding/src/onboarding/domain/entity"
	"onboarding/src/onboarding/domain/port"
)

// PostgresOnboardingRepository implementa OnboardingRepository usando PostgreSQL.
//
// RLS (E28, RULE-09/RULE-10): `onboarding_processes` y `verification_codes` tienen ROW
// LEVEL SECURITY forzado con la policy `tenant_isolation` (migración 006). Cada operación
// sobre esas tablas corre dentro de postgres.WithRLSInTransaction, que fija `app.tenant_id`
// con SET LOCAL — sin él, cualquier query erra (fail-closed) bajo el rol NOBYPASSRLS
// `onboarding_app`. El filtro manual `WHERE ... = $` se mantiene como defensa en
// profundidad (la policy ya lo garantiza).
//
// El wizard es PRE-AUTH: no hay JWT del cual sacar el tenant. GetProcessByID resuelve el
// tenant en dos pasos (D1): primero la función SECURITY DEFINER
// `onboarding_tenant_for_process` (migración 006) mapea el process_id — capability UUID no
// adivinable — al tenant_id, y recién entonces la query real corre bajo el scope RLS de ese
// tenant. `onboarding_step_definitions` es catálogo GLOBAL sin RLS (legible pre-auth por
// diseño) y se consulta sin transacción RLS.
type PostgresOnboardingRepository struct {
	db *sql.DB
}

// NewPostgresOnboardingRepository crea una nueva instancia del repositorio.
// Sin side-effects de escritura: el seed del catálogo vive en la migración 002.
func NewPostgresOnboardingRepository(db *sql.DB) port.OnboardingRepository {
	return &PostgresOnboardingRepository{db: db}
}

// SaveProcess guarda un proceso de onboarding. Tenant: process.TenantID (creado
// server-side vía IAM — path del WITH CHECK, nunca input del cliente).
func (r *PostgresOnboardingRepository) SaveProcess(ctx context.Context, process *entity.OnboardingProcess) error {
	stepsCompletedJSON, _ := json.Marshal(process.StepsCompleted)
	stepsPendingJSON, _ := json.Marshal(process.StepsPending)
	stepsSkippedJSON, _ := json.Marshal(process.StepsSkipped)

	query := `
		INSERT INTO onboarding_processes (
			id, tenant_id, user_id, current_step_number, is_completed,
			company_name, business_type, store_size, steps_completed,
			steps_pending, steps_skipped, started_at, completed_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (id) DO UPDATE SET
			current_step_number = EXCLUDED.current_step_number,
			is_completed = EXCLUDED.is_completed,
			company_name = EXCLUDED.company_name,
			business_type = EXCLUDED.business_type,
			store_size = EXCLUDED.store_size,
			steps_completed = EXCLUDED.steps_completed,
			steps_pending = EXCLUDED.steps_pending,
			steps_skipped = EXCLUDED.steps_skipped,
			completed_at = EXCLUDED.completed_at,
			updated_at = EXCLUDED.updated_at
	`

	rc := postgres.RLSContext{TenantID: process.TenantID.String()}
	return postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, query,
			process.ID,
			process.TenantID,
			process.UserID,
			process.CurrentStepNumber,
			process.IsCompleted,
			process.CompanyName,
			process.BusinessType,
			process.StoreSize,
			stepsCompletedJSON,
			stepsPendingJSON,
			stepsSkippedJSON,
			process.StartedAt,
			process.CompletedAt,
			process.CreatedAt,
			process.UpdatedAt,
		)
		return err
	})
}

// GetProcessByID obtiene un proceso por ID — two-step D1 (lookup pre-auth).
func (r *PostgresOnboardingRepository) GetProcessByID(ctx context.Context, id uuid.UUID) (*entity.OnboardingProcess, error) {
	// Paso 1: resolver el tenant vía la función SECURITY DEFINER, directo sobre la
	// conexión y SIN WithRLSInTransaction (todavía no hay tenant que fijar).
	var tenantID uuid.NullUUID
	if err := r.db.QueryRowContext(ctx,
		`SELECT onboarding_tenant_for_process($1)`, id,
	).Scan(&tenantID); err != nil {
		return nil, err
	}
	if !tenantID.Valid {
		// Proceso inexistente: mismo error que una fila no encontrada — no se
		// distingue "existe pero no es tuyo".
		return nil, sql.ErrNoRows
	}

	query := `
		SELECT id, tenant_id, user_id, current_step_number, is_completed,
			   company_name, business_type, store_size, steps_completed,
			   steps_pending, steps_skipped, started_at, completed_at,
			   created_at, updated_at
		FROM onboarding_processes
		WHERE id = $1
	`

	// Paso 2: la query real bajo el scope RLS del tenant resuelto.
	rc := postgres.RLSContext{TenantID: tenantID.UUID.String()}
	var process *entity.OnboardingProcess
	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		p, scanErr := r.scanProcess(tx.QueryRowContext(ctx, query, id))
		if scanErr != nil {
			return scanErr
		}
		process = p
		return nil
	})
	if err != nil {
		return nil, err
	}

	return process, nil
}

// UpdateProcess actualiza un proceso existente. Tenant: process.TenantID.
func (r *PostgresOnboardingRepository) UpdateProcess(ctx context.Context, process *entity.OnboardingProcess) error {
	stepsCompletedJSON, _ := json.Marshal(process.StepsCompleted)
	stepsPendingJSON, _ := json.Marshal(process.StepsPending)
	stepsSkippedJSON, _ := json.Marshal(process.StepsSkipped)

	query := `
		UPDATE onboarding_processes SET
			current_step_number = $2,
			is_completed = $3,
			company_name = $4,
			business_type = $5,
			store_size = $6,
			steps_completed = $7,
			steps_pending = $8,
			steps_skipped = $9,
			completed_at = $10,
			updated_at = $11
		WHERE id = $1
	`

	rc := postgres.RLSContext{TenantID: process.TenantID.String()}
	return postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, query,
			process.ID,
			process.CurrentStepNumber,
			process.IsCompleted,
			process.CompanyName,
			process.BusinessType,
			process.StoreSize,
			stepsCompletedJSON,
			stepsPendingJSON,
			stepsSkippedJSON,
			process.CompletedAt,
			process.UpdatedAt,
		)
		return err
	})
}

// GetStepDefinitions obtiene todas las definiciones de pasos (catálogo global sin RLS,
// pre-auth por diseño — sin transacción RLS a propósito).
func (r *PostgresOnboardingRepository) GetStepDefinitions(ctx context.Context) ([]*entity.StepDefinition, error) {
	query := `
		SELECT id, step_number, step_name, step_title, description, is_required,
			   has_ui, requires_user_input, can_be_skipped, display_order,
			   is_active, created_at, updated_at
		FROM onboarding_step_definitions
		WHERE is_active = true
		ORDER BY step_number
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stepDefinitions []*entity.StepDefinition
	for rows.Next() {
		stepDef, err := r.scanStepDefinition(rows)
		if err != nil {
			log.Printf("Error scanning step definition: %v", err)
			continue
		}
		stepDefinitions = append(stepDefinitions, stepDef)
	}

	return stepDefinitions, nil
}

// GetStepDefinitionByNumber obtiene una definición de paso por número (catálogo global
// sin RLS — sin transacción RLS a propósito).
func (r *PostgresOnboardingRepository) GetStepDefinitionByNumber(ctx context.Context, stepNumber int) (*entity.StepDefinition, error) {
	query := `
		SELECT id, step_number, step_name, step_title, description, is_required,
			   has_ui, requires_user_input, can_be_skipped, display_order,
			   is_active, created_at, updated_at
		FROM onboarding_step_definitions
		WHERE step_number = $1 AND is_active = true
	`

	row := r.db.QueryRowContext(ctx, query, stepNumber)
	return r.scanStepDefinitionRow(row)
}

// ===============================
// VERIFICATION CODES METHODS
// ===============================

// SaveVerificationCode guarda un código de verificación. Tenant: code.TenantID (derivado
// del proceso ya cargado por el caller — D3).
func (r *PostgresOnboardingRepository) SaveVerificationCode(ctx context.Context, code *entity.VerificationCode) error {
	query := `
		INSERT INTO verification_codes (
			id, process_id, tenant_id, email, code, is_used, expires_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			code = EXCLUDED.code,
			is_used = EXCLUDED.is_used,
			expires_at = EXCLUDED.expires_at,
			updated_at = EXCLUDED.updated_at
	`

	rc := postgres.RLSContext{TenantID: code.TenantID.String()}
	return postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, query,
			code.ID,
			code.ProcessID,
			code.TenantID,
			code.UserEmail,
			code.Code,
			code.IsUsed,
			code.ExpiresAt,
			code.CreatedAt,
			code.UpdatedAt,
		)
		return err
	})
}

// GetVerificationCodeByProcessID obtiene el último código de verificación de un proceso.
// tenantID explícito en la firma: los callers (VerifyEmail/ResendVerification) ya cargaron
// el proceso y tienen el tenant.
func (r *PostgresOnboardingRepository) GetVerificationCodeByProcessID(ctx context.Context, tenantID, processID uuid.UUID) (*entity.VerificationCode, error) {
	query := `
		SELECT id, process_id, tenant_id, email, code, is_used, expires_at, created_at, updated_at
		FROM verification_codes
		WHERE process_id = $1 AND tenant_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	rc := postgres.RLSContext{TenantID: tenantID.String()}
	var code *entity.VerificationCode
	err := postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		c, scanErr := r.scanVerificationCodeRow(tx.QueryRowContext(ctx, query, processID, tenantID))
		if scanErr != nil {
			return scanErr
		}
		code = c
		return nil
	})
	if err != nil {
		return nil, err
	}

	return code, nil
}

// UpdateVerificationCode actualiza un código de verificación. Tenant: code.TenantID.
func (r *PostgresOnboardingRepository) UpdateVerificationCode(ctx context.Context, code *entity.VerificationCode) error {
	query := `
		UPDATE verification_codes SET
			is_used = $2,
			updated_at = $3
		WHERE id = $1
	`

	rc := postgres.RLSContext{TenantID: code.TenantID.String()}
	return postgres.WithRLSInTransaction(ctx, r.db, rc, func(ctx context.Context, tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, query,
			code.ID,
			code.IsUsed,
			code.UpdatedAt,
		)
		return err
	})
}

// ===============================
// HELPER SCANNING METHODS
// ===============================

func (r *PostgresOnboardingRepository) scanProcess(row *sql.Row) (*entity.OnboardingProcess, error) {
	var process entity.OnboardingProcess
	var stepsCompletedJSON, stepsPendingJSON, stepsSkippedJSON []byte
	var completedAt sql.NullTime

	err := row.Scan(
		&process.ID,
		&process.TenantID,
		&process.UserID,
		&process.CurrentStepNumber,
		&process.IsCompleted,
		&process.CompanyName,
		&process.BusinessType,
		&process.StoreSize,
		&stepsCompletedJSON,
		&stepsPendingJSON,
		&stepsSkippedJSON,
		&process.StartedAt,
		&completedAt,
		&process.CreatedAt,
		&process.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if completedAt.Valid {
		process.CompletedAt = &completedAt.Time
	}

	// Deserializar arrays JSON
	json.Unmarshal(stepsCompletedJSON, &process.StepsCompleted)
	json.Unmarshal(stepsPendingJSON, &process.StepsPending)
	json.Unmarshal(stepsSkippedJSON, &process.StepsSkipped)

	return &process, nil
}

func (r *PostgresOnboardingRepository) scanStepDefinition(rows *sql.Rows) (*entity.StepDefinition, error) {
	var stepDef entity.StepDefinition

	err := rows.Scan(
		&stepDef.ID,
		&stepDef.StepNumber,
		&stepDef.StepName,
		&stepDef.StepTitle,
		&stepDef.Description,
		&stepDef.IsRequired,
		&stepDef.HasUI,
		&stepDef.RequiresUserInput,
		&stepDef.CanBeSkipped,
		&stepDef.DisplayOrder,
		&stepDef.IsActive,
		&stepDef.CreatedAt,
		&stepDef.UpdatedAt,
	)

	return &stepDef, err
}

func (r *PostgresOnboardingRepository) scanStepDefinitionRow(row *sql.Row) (*entity.StepDefinition, error) {
	var stepDef entity.StepDefinition

	err := row.Scan(
		&stepDef.ID,
		&stepDef.StepNumber,
		&stepDef.StepName,
		&stepDef.StepTitle,
		&stepDef.Description,
		&stepDef.IsRequired,
		&stepDef.HasUI,
		&stepDef.RequiresUserInput,
		&stepDef.CanBeSkipped,
		&stepDef.DisplayOrder,
		&stepDef.IsActive,
		&stepDef.CreatedAt,
		&stepDef.UpdatedAt,
	)

	return &stepDef, err
}

// scanVerificationCodeRow escanea una fila de código de verificación
func (r *PostgresOnboardingRepository) scanVerificationCodeRow(row *sql.Row) (*entity.VerificationCode, error) {
	var code entity.VerificationCode

	err := row.Scan(
		&code.ID,
		&code.ProcessID,
		&code.TenantID,
		&code.UserEmail,
		&code.Code,
		&code.IsUsed,
		&code.ExpiresAt,
		&code.CreatedAt,
		&code.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &code, nil
}
