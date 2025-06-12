package port

import (
	"onboarding/src/onboarding/domain/entity"

	"github.com/google/uuid"
)

// OnboardingRepository define las operaciones de persistencia para el onboarding
type OnboardingRepository interface {
	// Process operations
	SaveProcess(process *entity.OnboardingProcess) error
	GetProcessByID(id uuid.UUID) (*entity.OnboardingProcess, error)
	GetProcessByTenantID(tenantID uuid.UUID) (*entity.OnboardingProcess, error)
	GetProcessByUserID(userID uuid.UUID) (*entity.OnboardingProcess, error)
	UpdateProcess(process *entity.OnboardingProcess) error
	DeleteProcess(id uuid.UUID) error

	// Step definitions operations
	GetStepDefinitions() ([]*entity.StepDefinition, error)
	GetStepDefinitionByNumber(stepNumber int) (*entity.StepDefinition, error)
	SaveStepDefinition(stepDef *entity.StepDefinition) error

	// Queries
	GetActiveProcesses() ([]*entity.OnboardingProcess, error)
	GetCompletedProcesses() ([]*entity.OnboardingProcess, error)
	GetProcessesByDateRange(startDate, endDate string) ([]*entity.OnboardingProcess, error)
	GetProcessStats() (*ProcessStats, error)
}

// ProcessStats contiene estadísticas del proceso de onboarding
type ProcessStats struct {
	TotalProcesses     int     `json:"total_processes"`
	CompletedProcesses int     `json:"completed_processes"`
	ActiveProcesses    int     `json:"active_processes"`
	CompletionRate     float64 `json:"completion_rate"`
	AverageTimeMinutes float64 `json:"average_time_minutes"`
}
