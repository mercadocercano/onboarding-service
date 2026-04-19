package request

import (
	"errors"
	"strings"
)

// SetupStoreRequest representa la solicitud de configuración de la tienda
type SetupStoreRequest struct {
	ProcessID          string   `json:"process_id" validate:"required"`
	StoreName          string   `json:"store_name" validate:"required,min=2,max=100"`
	BusinessType       string   `json:"business_type" validate:"required"`
	StoreSize          string   `json:"store_size" validate:"required"`
	SelectedCategories []string `json:"selected_categories,omitempty"`
}

// Validate valida los datos de la solicitud
func (r *SetupStoreRequest) Validate() error {
	if err := r.validateRequired(); err != nil {
		return err
	}

	if err := r.validateStoreName(); err != nil {
		return err
	}

	if err := r.validateBusinessType(); err != nil {
		return err
	}

	if err := r.validateStoreSize(); err != nil {
		return err
	}

	if len(r.SelectedCategories) > 0 {
		if err := r.validateCategories(); err != nil {
			return err
		}
	}

	return nil
}

func (r *SetupStoreRequest) validateRequired() error {
	if strings.TrimSpace(r.ProcessID) == "" {
		return errors.New("el ID del proceso es requerido")
	}
	if strings.TrimSpace(r.StoreName) == "" {
		return errors.New("el nombre de la tienda es requerido")
	}
	if strings.TrimSpace(r.BusinessType) == "" {
		return errors.New("el tipo de negocio es requerido")
	}
	if strings.TrimSpace(r.StoreSize) == "" {
		return errors.New("el tamaño de la tienda es requerido")
	}
	return nil
}

func (r *SetupStoreRequest) validateStoreName() error {
	name := strings.TrimSpace(r.StoreName)
	if len(name) < 2 {
		return errors.New("el nombre de la tienda debe tener al menos 2 caracteres")
	}
	if len(name) > 100 {
		return errors.New("el nombre de la tienda no puede tener más de 100 caracteres")
	}
	r.StoreName = name // Normalizar
	return nil
}

func (r *SetupStoreRequest) validateBusinessType() error {
	// SIMPLIFICADO: La validación real se hace contra el PIM service en el usecase
	// No necesitamos mantener una lista hardcodeada aquí
	if strings.TrimSpace(r.BusinessType) == "" {
		return errors.New("el tipo de negocio es requerido")
	}

	// Validación básica de formato
	if len(r.BusinessType) < 2 || len(r.BusinessType) > 50 {
		return errors.New("el tipo de negocio debe tener entre 2 y 50 caracteres")
	}

	return nil
}

func (r *SetupStoreRequest) validateStoreSize() error {
	validSizes := map[string]bool{
		"micro":    true,
		"pyme":     true,
		"multiple": true,
	}

	if !validSizes[r.StoreSize] {
		return errors.New("tamaño de tienda no válido. Debe ser: micro, pyme o multiple")
	}
	return nil
}

func (r *SetupStoreRequest) validateCategories() error {
	if len(r.SelectedCategories) == 0 {
		return nil
	}

	if len(r.SelectedCategories) > 10 {
		return errors.New("no puede seleccionar más de 10 categorías")
	}

	seen := make(map[string]bool)
	cleanCategories := []string{}

	for _, category := range r.SelectedCategories {
		clean := strings.TrimSpace(category)
		if clean == "" {
			continue
		}
		if seen[clean] {
			continue
		}
		seen[clean] = true
		cleanCategories = append(cleanCategories, clean)
	}

	r.SelectedCategories = cleanCategories
	return nil
}

// GetCleanStoreName retorna el nombre de la tienda limpio
func (r *SetupStoreRequest) GetCleanStoreName() string {
	return strings.TrimSpace(r.StoreName)
}

// GetRecommendedPlan determina el plan recomendado basado en el tamaño
func (r *SetupStoreRequest) GetRecommendedPlan() string {
	switch r.StoreSize {
	case "micro":
		return "basic"
	case "pyme":
		return "premium"
	case "multiple":
		return "enterprise"
	default:
		return "basic"
	}
}

// GetBusinessTypeDisplayName retorna el nombre para mostrar del tipo de negocio
// SIMPLIFICADO: Solo retorna el business type tal como viene del PIM
func (r *SetupStoreRequest) GetBusinessTypeDisplayName() string {
	// El nombre display se obtiene dinámicamente del PIM service
	// Este método se mantiene por compatibilidad pero ahora es simple
	if r.BusinessType == "" {
		return "Tipo de negocio no especificado"
	}
	return r.BusinessType
}
