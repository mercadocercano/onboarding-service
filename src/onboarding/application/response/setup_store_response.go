package response

import (
	"time"
)

// SetupStoreResponse representa la respuesta de la configuración de tienda
type SetupStoreResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	ProcessID string `json:"process_id"`
	TenantID  string `json:"tenant_id"`
	NextStep  int    `json:"next_step"`

	// Datos de la tienda configurada
	StoreData StoreData `json:"store_data"`

	// Plan recomendado
	RecommendedPlan string `json:"recommended_plan"`

	CreatedAt time.Time `json:"created_at"`
}

// StoreData contiene información de la tienda configurada
type StoreData struct {
	Name               string            `json:"name"`
	BusinessType       *BusinessTypeInfo `json:"business_type"`
	StoreSize          string            `json:"store_size"`
	SelectedCategories []string          `json:"selected_categories"`
}

// BusinessTypeInfo contiene información del tipo de negocio
type BusinessTypeInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
}

// NewSetupStoreResponse crea una nueva respuesta de configuración exitosa
func NewSetupStoreResponse(
	processID, tenantID, storeName string,
	businessType *BusinessTypeInfo,
	storeSize, recommendedPlan string,
	selectedCategories []string,
	nextStep int,
) *SetupStoreResponse {
	storeData := StoreData{
		Name:               storeName,
		BusinessType:       businessType,
		StoreSize:          storeSize,
		SelectedCategories: selectedCategories,
	}

	return &SetupStoreResponse{
		Success:         true,
		Message:         "Tienda configurada exitosamente",
		ProcessID:       processID,
		TenantID:        tenantID,
		NextStep:        nextStep,
		StoreData:       storeData,
		RecommendedPlan: recommendedPlan,
		CreatedAt:       time.Now(),
	}
}

// NewSetupStoreErrorResponse crea una respuesta de error
func NewSetupStoreErrorResponse(message string) *SetupStoreResponse {
	return &SetupStoreResponse{
		Success:   false,
		Message:   message,
		NextStep:  4, // Permanecer en el paso de configuración
		CreatedAt: time.Now(),
	}
}
