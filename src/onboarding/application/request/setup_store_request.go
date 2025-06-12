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
	SelectedCategories []string `json:"selected_categories" validate:"required,min=1"`
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

	if err := r.validateCategories(); err != nil {
		return err
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
	if len(r.SelectedCategories) == 0 {
		return errors.New("debe seleccionar al menos una categoría")
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
	validBusinessTypes := map[string]bool{
		"home-construction":   true,
		"fashion":             true,
		"electronics":         true,
		"automotive":          true,
		"books-media":         true,
		"food-beverage":       true,
		"health-pharmacy":     true,
		"sports-fitness":      true,
		"beauty-cosmetics":    true,
		"toys-games":          true,
		"pet-supplies":        true,
		"office-supplies":     true,
		"jewelry-accessories": true,
		"polirubro":           true,
	}

	if !validBusinessTypes[r.BusinessType] {
		return errors.New("tipo de negocio no válido")
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
		return errors.New("debe seleccionar al menos una categoría")
	}

	// Limitar a máximo 10 categorías
	if len(r.SelectedCategories) > 10 {
		return errors.New("no puede seleccionar más de 10 categorías")
	}

	// Validar que no haya categorías vacías o duplicadas
	seen := make(map[string]bool)
	cleanCategories := []string{}

	for _, category := range r.SelectedCategories {
		clean := strings.TrimSpace(category)
		if clean == "" {
			continue // Saltar categorías vacías
		}
		if seen[clean] {
			continue // Saltar duplicadas
		}
		seen[clean] = true
		cleanCategories = append(cleanCategories, clean)
	}

	if len(cleanCategories) == 0 {
		return errors.New("debe seleccionar al menos una categoría válida")
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
		return "professional"
	case "multiple":
		return "enterprise"
	default:
		return "basic"
	}
}

// GetBusinessTypeDisplayName retorna el nombre para mostrar del tipo de negocio
func (r *SetupStoreRequest) GetBusinessTypeDisplayName() string {
	displayNames := map[string]string{
		"home-construction":   "Hogar y Construcción",
		"fashion":             "Moda y Vestimenta",
		"electronics":         "Electrónicos y Tecnología",
		"automotive":          "Automotriz y Repuestos",
		"books-media":         "Libros y Medios",
		"food-beverage":       "Alimentos y Bebidas",
		"health-pharmacy":     "Salud y Farmacia",
		"sports-fitness":      "Deportes y Fitness",
		"beauty-cosmetics":    "Belleza y Cosméticos",
		"toys-games":          "Juguetes y Juegos",
		"pet-supplies":        "Mascotas y Suministros",
		"office-supplies":     "Oficina y Papelería",
		"jewelry-accessories": "Joyería y Accesorios",
		"polirubro":           "Polirubro",
	}

	if displayName, exists := displayNames[r.BusinessType]; exists {
		return displayName
	}
	return r.BusinessType
}
