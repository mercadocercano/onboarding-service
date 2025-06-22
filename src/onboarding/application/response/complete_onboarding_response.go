package response

import (
	"time"
)

// CompleteOnboardingResponse representa la respuesta de completar el onboarding
type CompleteOnboardingResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ProcessID string `json:"process_id"`

	// Datos del proceso completado
	Completed   bool      `json:"completed"`
	CompletedAt time.Time `json:"completed_at"`
	TenantID    string    `json:"tenant_id"`

	// URLs y próximos pasos
	BackofficeURL string   `json:"backoffice_url"`
	NextSteps     []string `json:"next_steps"`

	// Resumen del onboarding
	Summary OnboardingSummary `json:"summary"`

	CreatedAt time.Time `json:"created_at"`
}

// OnboardingSummary contiene un resumen del proceso completado
type OnboardingSummary struct {
	CompanyName        string   `json:"company_name"`
	BusinessType       string   `json:"business_type"`
	StoreSize          string   `json:"store_size"`
	SelectedCategories []string `json:"selected_categories"`
	StepsCompleted     []int    `json:"steps_completed"`
	TotalSteps         int      `json:"total_steps"`
	StartedAt          string   `json:"started_at"`
	DurationMinutes    int      `json:"duration_minutes"`
}

// NewCompleteOnboardingResponse crea una nueva respuesta de completar exitosa
func NewCompleteOnboardingResponse(
	processID, tenantID, companyName, businessType, storeSize string,
	selectedCategories []string,
	stepsCompleted []int,
	startedAt time.Time,
) *CompleteOnboardingResponse {
	now := time.Now()
	durationMinutes := int(now.Sub(startedAt).Minutes())

	summary := OnboardingSummary{
		CompanyName:        companyName,
		BusinessType:       businessType,
		StoreSize:          storeSize,
		SelectedCategories: selectedCategories,
		StepsCompleted:     stepsCompleted,
		TotalSteps:         6,
		StartedAt:          startedAt.Format("2006-01-02 15:04:05"),
		DurationMinutes:    durationMinutes,
	}

	return &CompleteOnboardingResponse{
		Success:       true,
		Message:       "Onboarding completado exitosamente",
		ProcessID:     processID,
		Completed:     true,
		CompletedAt:   now,
		TenantID:      tenantID,
		BackofficeURL: "/backoffice/dashboard",
		NextSteps: []string{
			"Configurar catálogo de productos",
			"Personalizar la tienda",
			"Añadir productos",
			"Configurar métodos de pago",
			"Configurar envíos",
		},
		Summary:   summary,
		CreatedAt: now,
	}
}

// NewCompleteOnboardingErrorResponse crea una respuesta de error
func NewCompleteOnboardingErrorResponse(message string) *CompleteOnboardingResponse {
	return &CompleteOnboardingResponse{
		Success:   false,
		Message:   message,
		CreatedAt: time.Now(),
	}
}
