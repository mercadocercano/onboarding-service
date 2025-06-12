package persistence

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/google/uuid"

	"onboarding/src/onboarding/domain/entity"
	"onboarding/src/onboarding/domain/port"
)

// PostgresOnboardingRepository implementa OnboardingRepository usando PostgreSQL
type PostgresOnboardingRepository struct {
	db *sql.DB
}

// NewPostgresOnboardingRepository crea una nueva instancia del repositorio
func NewPostgresOnboardingRepository(db *sql.DB) port.OnboardingRepository {
	repo := &PostgresOnboardingRepository{db: db}

	// Inicializar datos por defecto si es necesario
	if err := repo.initializeDefaultData(); err != nil {
		log.Printf("Warning: Could not initialize default data: %v", err)
	}

	return repo
}

// SaveProcess guarda un proceso de onboarding
func (r *PostgresOnboardingRepository) SaveProcess(process *entity.OnboardingProcess) error {
	query := `
		INSERT INTO onboarding_processes (
			id, tenant_id, user_id, current_step_number, is_completed,
			company_name, business_type, store_size, steps_completed,
			steps_pending, steps_skipped, started_at, completed_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`

	stepsCompletedJSON, _ := json.Marshal(process.StepsCompleted)
	stepsPendingJSON, _ := json.Marshal(process.StepsPending)
	stepsSkippedJSON, _ := json.Marshal(process.StepsSkipped)

	_, err := r.db.Exec(query,
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

	if err != nil {
		log.Printf("Error saving onboarding process: %v", err)
		return err
	}

	log.Printf("Onboarding process saved successfully: %s", process.ID)
	return nil
}

// GetProcessByID obtiene un proceso por ID
func (r *PostgresOnboardingRepository) GetProcessByID(id uuid.UUID) (*entity.OnboardingProcess, error) {
	query := `
		SELECT id, tenant_id, user_id, current_step_number, is_completed,
			   company_name, business_type, store_size, steps_completed,
			   steps_pending, steps_skipped, started_at, completed_at,
			   created_at, updated_at
		FROM onboarding_processes
		WHERE id = $1
	`

	row := r.db.QueryRow(query, id)
	return r.scanProcess(row)
}

// GetProcessByTenantID obtiene un proceso por tenant ID
func (r *PostgresOnboardingRepository) GetProcessByTenantID(tenantID uuid.UUID) (*entity.OnboardingProcess, error) {
	query := `
		SELECT id, tenant_id, user_id, current_step_number, is_completed,
			   company_name, business_type, store_size, steps_completed,
			   steps_pending, steps_skipped, started_at, completed_at,
			   created_at, updated_at
		FROM onboarding_processes
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.db.QueryRow(query, tenantID)
	return r.scanProcess(row)
}

// GetProcessByUserID obtiene un proceso por user ID
func (r *PostgresOnboardingRepository) GetProcessByUserID(userID uuid.UUID) (*entity.OnboardingProcess, error) {
	query := `
		SELECT id, tenant_id, user_id, current_step_number, is_completed,
			   company_name, business_type, store_size, steps_completed,
			   steps_pending, steps_skipped, started_at, completed_at,
			   created_at, updated_at
		FROM onboarding_processes
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.db.QueryRow(query, userID)
	return r.scanProcess(row)
}

// UpdateProcess actualiza un proceso existente
func (r *PostgresOnboardingRepository) UpdateProcess(process *entity.OnboardingProcess) error {
	query := `
		UPDATE onboarding_processes
		SET current_step_number = $1, is_completed = $2, company_name = $3,
			business_type = $4, store_size = $5, steps_completed = $6,
			steps_pending = $7, steps_skipped = $8, completed_at = $9,
			updated_at = $10
		WHERE id = $11
	`

	stepsCompletedJSON, _ := json.Marshal(process.StepsCompleted)
	stepsPendingJSON, _ := json.Marshal(process.StepsPending)
	stepsSkippedJSON, _ := json.Marshal(process.StepsSkipped)

	_, err := r.db.Exec(query,
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
		process.ID,
	)

	if err != nil {
		log.Printf("Error updating onboarding process: %v", err)
		return err
	}

	return nil
}

// DeleteProcess elimina un proceso
func (r *PostgresOnboardingRepository) DeleteProcess(id uuid.UUID) error {
	query := `DELETE FROM onboarding_processes WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// GetStepDefinitions obtiene todas las definiciones de pasos
func (r *PostgresOnboardingRepository) GetStepDefinitions() ([]*entity.StepDefinition, error) {
	query := `
		SELECT id, step_number, step_name, step_title, description, is_required,
			   has_ui, requires_user_input, can_be_skipped, display_order,
			   is_active, created_at, updated_at
		FROM onboarding_step_definitions
		WHERE is_active = true
		ORDER BY display_order
	`

	rows, err := r.db.Query(query)
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

// GetStepDefinitionByNumber obtiene una definición de paso por número
func (r *PostgresOnboardingRepository) GetStepDefinitionByNumber(stepNumber int) (*entity.StepDefinition, error) {
	query := `
		SELECT id, step_number, step_name, step_title, description, is_required,
			   has_ui, requires_user_input, can_be_skipped, display_order,
			   is_active, created_at, updated_at
		FROM onboarding_step_definitions
		WHERE step_number = $1 AND is_active = true
	`

	row := r.db.QueryRow(query, stepNumber)
	return r.scanStepDefinitionRow(row)
}

// SaveStepDefinition guarda una definición de paso
func (r *PostgresOnboardingRepository) SaveStepDefinition(stepDef *entity.StepDefinition) error {
	query := `
		INSERT INTO onboarding_step_definitions (
			id, step_number, step_name, step_title, description, is_required,
			has_ui, requires_user_input, can_be_skipped, display_order,
			is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (step_number) DO UPDATE SET
			step_name = EXCLUDED.step_name,
			step_title = EXCLUDED.step_title,
			description = EXCLUDED.description,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(query,
		stepDef.ID,
		stepDef.StepNumber,
		stepDef.StepName,
		stepDef.StepTitle,
		stepDef.Description,
		stepDef.IsRequired,
		stepDef.HasUI,
		stepDef.RequiresUserInput,
		stepDef.CanBeSkipped,
		stepDef.DisplayOrder,
		stepDef.IsActive,
		stepDef.CreatedAt,
		stepDef.UpdatedAt,
	)

	return err
}

// GetActiveProcesses obtiene procesos activos
func (r *PostgresOnboardingRepository) GetActiveProcesses() ([]*entity.OnboardingProcess, error) {
	query := `
		SELECT id, tenant_id, user_id, current_step_number, is_completed,
			   company_name, business_type, store_size, steps_completed,
			   steps_pending, steps_skipped, started_at, completed_at,
			   created_at, updated_at
		FROM onboarding_processes
		WHERE is_completed = false
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var processes []*entity.OnboardingProcess
	for rows.Next() {
		process, err := r.scanProcessRows(rows)
		if err != nil {
			log.Printf("Error scanning process: %v", err)
			continue
		}
		processes = append(processes, process)
	}

	return processes, nil
}

// GetCompletedProcesses obtiene procesos completados
func (r *PostgresOnboardingRepository) GetCompletedProcesses() ([]*entity.OnboardingProcess, error) {
	query := `
		SELECT id, tenant_id, user_id, current_step_number, is_completed,
			   company_name, business_type, store_size, steps_completed,
			   steps_pending, steps_skipped, started_at, completed_at,
			   created_at, updated_at
		FROM onboarding_processes
		WHERE is_completed = true
		ORDER BY completed_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var processes []*entity.OnboardingProcess
	for rows.Next() {
		process, err := r.scanProcessRows(rows)
		if err != nil {
			log.Printf("Error scanning process: %v", err)
			continue
		}
		processes = append(processes, process)
	}

	return processes, nil
}

// GetProcessesByDateRange obtiene procesos por rango de fechas
func (r *PostgresOnboardingRepository) GetProcessesByDateRange(startDate, endDate string) ([]*entity.OnboardingProcess, error) {
	query := `
		SELECT id, tenant_id, user_id, current_step_number, is_completed,
			   company_name, business_type, store_size, steps_completed,
			   steps_pending, steps_skipped, started_at, completed_at,
			   created_at, updated_at
		FROM onboarding_processes
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var processes []*entity.OnboardingProcess
	for rows.Next() {
		process, err := r.scanProcessRows(rows)
		if err != nil {
			log.Printf("Error scanning process: %v", err)
			continue
		}
		processes = append(processes, process)
	}

	return processes, nil
}

// GetProcessStats obtiene estadísticas de procesos
func (r *PostgresOnboardingRepository) GetProcessStats() (*port.ProcessStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_processes,
			COUNT(CASE WHEN is_completed = true THEN 1 END) as completed_processes,
			COUNT(CASE WHEN is_completed = false THEN 1 END) as active_processes,
			AVG(CASE WHEN completed_at IS NOT NULL AND started_at IS NOT NULL 
				THEN EXTRACT(EPOCH FROM (completed_at - started_at))/60 END) as avg_time_minutes
		FROM onboarding_processes
	`

	row := r.db.QueryRow(query)

	var stats port.ProcessStats
	var avgTimeMinutes sql.NullFloat64

	err := row.Scan(
		&stats.TotalProcesses,
		&stats.CompletedProcesses,
		&stats.ActiveProcesses,
		&avgTimeMinutes,
	)

	if err != nil {
		return nil, err
	}

	if avgTimeMinutes.Valid {
		stats.AverageTimeMinutes = avgTimeMinutes.Float64
	}

	if stats.TotalProcesses > 0 {
		stats.CompletionRate = float64(stats.CompletedProcesses) / float64(stats.TotalProcesses) * 100
	}

	return &stats, nil
}

// Métodos auxiliares para escanear resultados

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

func (r *PostgresOnboardingRepository) scanProcessRows(rows *sql.Rows) (*entity.OnboardingProcess, error) {
	var process entity.OnboardingProcess
	var stepsCompletedJSON, stepsPendingJSON, stepsSkippedJSON []byte
	var completedAt sql.NullTime

	err := rows.Scan(
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

// initializeDefaultData inicializa datos por defecto
func (r *PostgresOnboardingRepository) initializeDefaultData() error {
	// Verificar si ya existen step definitions
	count := 0
	err := r.db.QueryRow("SELECT COUNT(*) FROM onboarding_step_definitions").Scan(&count)
	if err != nil {
		log.Printf("Error checking existing step definitions: %v", err)
		return err
	}

	if count > 0 {
		log.Println("Step definitions already exist, skipping initialization")
		return nil
	}

	// Insertar step definitions por defecto
	stepDefinitions := entity.GetDefaultStepDefinitions()
	for _, stepDef := range stepDefinitions {
		err := r.SaveStepDefinition(stepDef)
		if err != nil {
			log.Printf("Error saving default step definition %d: %v", stepDef.StepNumber, err)
			continue
		}
	}

	log.Printf("Initialized %d default step definitions", len(stepDefinitions))
	return nil
}
