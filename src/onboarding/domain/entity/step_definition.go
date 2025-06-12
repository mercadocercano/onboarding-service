package entity

import (
	"time"

	"github.com/google/uuid"
)

// StepDefinition define la estructura de un paso de onboarding
type StepDefinition struct {
	ID          uuid.UUID `json:"id" db:"id"`
	StepNumber  int       `json:"step_number" db:"step_number"`
	StepName    string    `json:"step_name" db:"step_name"`
	StepTitle   string    `json:"step_title" db:"step_title"`
	Description string    `json:"description" db:"description"`
	IsRequired  bool      `json:"is_required" db:"is_required"`

	// Configuración del paso
	HasUI             bool `json:"has_ui" db:"has_ui"`
	RequiresUserInput bool `json:"requires_user_input" db:"requires_user_input"`
	CanBeSkipped      bool `json:"can_be_skipped" db:"can_be_skipped"`

	// Orden y agrupación
	DisplayOrder int  `json:"display_order" db:"display_order"`
	IsActive     bool `json:"is_active" db:"is_active"`

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// NewStepDefinition crea una nueva definición de paso
func NewStepDefinition(stepNumber int, stepName, stepTitle, description string) *StepDefinition {
	now := time.Now()
	return &StepDefinition{
		ID:                uuid.New(),
		StepNumber:        stepNumber,
		StepName:          stepName,
		StepTitle:         stepTitle,
		Description:       description,
		IsRequired:        true,
		HasUI:             true,
		RequiresUserInput: true,
		CanBeSkipped:      false,
		DisplayOrder:      stepNumber,
		IsActive:          true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

// GetDefaultStepDefinitions retorna las definiciones por defecto de los pasos
func GetDefaultStepDefinitions() []*StepDefinition {
	return []*StepDefinition{
		{
			ID:                uuid.New(),
			StepNumber:        1,
			StepName:          "bienvenida",
			StepTitle:         "Bienvenida",
			Description:       "Pantalla de bienvenida al onboarding",
			IsRequired:        true,
			HasUI:             true,
			RequiresUserInput: false,
			CanBeSkipped:      false,
			DisplayOrder:      1,
			IsActive:          true,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		{
			ID:                uuid.New(),
			StepNumber:        2,
			StepName:          "registro",
			StepTitle:         "Datos de Usuario",
			Description:       "Captura nombre, email y contraseña",
			IsRequired:        true,
			HasUI:             true,
			RequiresUserInput: true,
			CanBeSkipped:      false,
			DisplayOrder:      2,
			IsActive:          true,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		{
			ID:                uuid.New(),
			StepNumber:        3,
			StepName:          "verificacion",
			StepTitle:         "Verificar Email",
			Description:       "Validación del código de verificación",
			IsRequired:        true,
			HasUI:             true,
			RequiresUserInput: true,
			CanBeSkipped:      false,
			DisplayOrder:      3,
			IsActive:          true,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		{
			ID:                uuid.New(),
			StepNumber:        4,
			StepName:          "configuracion_tienda",
			StepTitle:         "Detalles de la Tienda",
			Description:       "Configuración básica del negocio",
			IsRequired:        true,
			HasUI:             true,
			RequiresUserInput: true,
			CanBeSkipped:      false,
			DisplayOrder:      4,
			IsActive:          true,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		{
			ID:                uuid.New(),
			StepNumber:        5,
			StepName:          "completar",
			StepTitle:         "Finalizar",
			Description:       "Completar onboarding y redirección",
			IsRequired:        true,
			HasUI:             true,
			RequiresUserInput: false,
			CanBeSkipped:      false,
			DisplayOrder:      5,
			IsActive:          true,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
	}
}

// IsValid valida la definición del paso
func (sd *StepDefinition) IsValid() bool {
	return sd.StepNumber > 0 &&
		sd.StepName != "" &&
		sd.StepTitle != "" &&
		sd.DisplayOrder > 0
}

// IsAutomaticStep verifica si el paso es automático (no requiere input del usuario)
func (sd *StepDefinition) IsAutomaticStep() bool {
	return !sd.RequiresUserInput
}

// IsInteractiveStep verifica si el paso requiere interacción del usuario
func (sd *StepDefinition) IsInteractiveStep() bool {
	return sd.RequiresUserInput && sd.HasUI
}
