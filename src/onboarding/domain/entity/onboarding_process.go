package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// OnboardingProcess representa el proceso completo de onboarding de un usuario
type OnboardingProcess struct {
	ID       uuid.UUID `json:"id" db:"id"`
	TenantID uuid.UUID `json:"tenant_id" db:"tenant_id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`

	// Estado del proceso
	CurrentStepNumber int  `json:"current_step_number" db:"current_step_number"`
	IsCompleted       bool `json:"is_completed" db:"is_completed"`

	// Datos capturados durante el onboarding
	CompanyName  string `json:"company_name" db:"company_name"`
	BusinessType string `json:"business_type" db:"business_type"` // 'pintureria', 'ferreteria', etc.
	StoreSize    string `json:"store_size" db:"store_size"`       // 'micro', 'pyme', 'multiple'

	// Datos adicionales del negocio
	SelectedCategories []string `json:"selected_categories"`

	// Control de progreso
	StepsCompleted []int `json:"steps_completed"`
	StepsPending   []int `json:"steps_pending"`
	StepsSkipped   []int `json:"steps_skipped"`

	// Timing
	StartedAt   time.Time  `json:"started_at" db:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewOnboardingProcess crea un nuevo proceso de onboarding
func NewOnboardingProcess(tenantID, userID uuid.UUID) *OnboardingProcess {
	now := time.Now()
	return &OnboardingProcess{
		ID:                uuid.New(),
		TenantID:          tenantID,
		UserID:            userID,
		CurrentStepNumber: 1,
		IsCompleted:       false,
		StepsCompleted:    []int{},
		StepsPending:      []int{1, 2, 3, 4, 5},
		StepsSkipped:      []int{},
		StartedAt:         now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// AdvanceToStep avanza el proceso al siguiente paso
func (op *OnboardingProcess) AdvanceToStep(stepNumber int) {
	// Marcar paso anterior como completado
	if !op.isStepCompleted(op.CurrentStepNumber) {
		op.StepsCompleted = append(op.StepsCompleted, op.CurrentStepNumber)
		op.removePendingStep(op.CurrentStepNumber)
	}

	// Avanzar al siguiente paso
	op.CurrentStepNumber = stepNumber
	op.UpdatedAt = time.Now()
}

// CompleteStep marca un paso como completado
func (op *OnboardingProcess) CompleteStep(stepNumber int) {
	if !op.isStepCompleted(stepNumber) {
		op.StepsCompleted = append(op.StepsCompleted, stepNumber)
		op.removePendingStep(stepNumber)
	}
	op.UpdatedAt = time.Now()
}

// CompleteOnboarding marca todo el proceso como completado
func (op *OnboardingProcess) CompleteOnboarding() {
	op.IsCompleted = true
	now := time.Now()
	op.CompletedAt = &now
	op.UpdatedAt = now

	// Marcar todos los pasos pendientes como completados
	for _, step := range op.StepsPending {
		if !op.isStepCompleted(step) {
			op.StepsCompleted = append(op.StepsCompleted, step)
		}
	}
	op.StepsPending = []int{}
}

// SetStoreConfiguration establece la configuración de la tienda
func (op *OnboardingProcess) SetStoreConfiguration(storeName, businessType, storeSize string, categories []string) {
	op.CompanyName = storeName
	op.BusinessType = businessType
	op.StoreSize = storeSize
	op.SelectedCategories = categories
	op.UpdatedAt = time.Now()
}

// GetProgress calcula el progreso del onboarding
func (op *OnboardingProcess) GetProgress() float64 {
	totalSteps := 5.0
	completedSteps := float64(len(op.StepsCompleted))
	return (completedSteps / totalSteps) * 100
}

// IsStepCompleted verifica si un paso está completado
func (op *OnboardingProcess) IsStepCompleted(stepNumber int) bool {
	return op.isStepCompleted(stepNumber)
}

// GetNextStep obtiene el siguiente paso a completar
func (op *OnboardingProcess) GetNextStep() int {
	if len(op.StepsPending) > 0 {
		return op.StepsPending[0]
	}
	if op.IsCompleted {
		return 0 // No hay más pasos
	}
	return op.CurrentStepNumber
}

// Métodos auxiliares privados
func (op *OnboardingProcess) isStepCompleted(stepNumber int) bool {
	for _, completed := range op.StepsCompleted {
		if completed == stepNumber {
			return true
		}
	}
	return false
}

func (op *OnboardingProcess) removePendingStep(stepNumber int) {
	newPending := []int{}
	for _, pending := range op.StepsPending {
		if pending != stepNumber {
			newPending = append(newPending, pending)
		}
	}
	op.StepsPending = newPending
}

// ToJSON convierte la entidad a JSON
func (op *OnboardingProcess) ToJSON() ([]byte, error) {
	return json.Marshal(op)
}

// Validation methods
func (op *OnboardingProcess) IsValid() bool {
	return op.TenantID != uuid.Nil &&
		op.UserID != uuid.Nil &&
		op.CurrentStepNumber >= 1 &&
		op.CurrentStepNumber <= 5
}

func (op *OnboardingProcess) HasValidStoreConfig() bool {
	return op.CompanyName != "" &&
		op.BusinessType != "" &&
		op.StoreSize != ""
}
