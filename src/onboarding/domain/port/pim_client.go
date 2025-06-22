package port

// PIMClient define las operaciones de integración con el servicio PIM
type PIMClient interface {
	// Business Types operations
	GetBusinessTypes() ([]*BusinessType, error)
	GetBusinessType(businessTypeID string) (*BusinessType, error)

	// Categories operations
	GetCategoriesByBusinessType(businessType string) ([]*Category, error)
	GetCategories() ([]*Category, error)
	GetCategory(categoryID string) (*Category, error)

	// Quickstart operations
	ApplyQuickstartTemplate(tenantID string, config *QuickstartConfig) (*QuickstartResponse, error)
	GetQuickstartTemplate(businessType string) (*QuickstartTemplate, error)

	// Product setup operations
	CreateInitialProducts(tenantID string, config *InitialProductsConfig) (*ProductSetupResponse, error)
}

// BusinessType representa un tipo de negocio en PIM
// ACTUALIZADO: Para mapear correctamente la respuesta REAL del PIM Service
type BusinessType struct {
	// Campos principales que devuelve el PIM Service
	ID          string `json:"id"`          // ID funcional (retail, food-beverage, etc.)
	Name        string `json:"name"`        // Nombre para mostrar
	Description string `json:"description"` // Descripción del tipo de negocio
	Icon        string `json:"icon"`        // Icono para la UI

	// Timestamps en camelCase como devuelve el PIM real
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`

	// Campos adicionales opcionales (no siempre presentes en respuesta del PIM)
	Color     string `json:"color,omitempty"`      // Color para la UI
	IsActive  bool   `json:"is_active,omitempty"`  // Si está activo
	SortOrder int    `json:"sort_order,omitempty"` // Orden de visualización

	// Campos de estructura interna para responses del onboarding (agregados por este servicio)
	DefaultAttributes []string `json:"default_attributes,omitempty"`
	DefaultCategories []string `json:"default_categories,omitempty"`
	DefaultVariants   []string `json:"default_variants,omitempty"`

	// Metadata adicional
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GetID retorna el ID funcional
func (bt *BusinessType) GetID() string {
	return bt.ID
}

// GetBusinessTypeCode retorna el código funcional
func (bt *BusinessType) GetBusinessTypeCode() string {
	return bt.ID
}

// Category representa una categoría de productos
type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ParentID    string `json:"parent_id,omitempty"`
	Level       int    `json:"level"`
	Path        string `json:"path"`
	IsActive    bool   `json:"is_active"`

	// Metadata para onboarding
	BusinessTypes []string `json:"business_types"`
	Icon          string   `json:"icon,omitempty"`
	SortOrder     int      `json:"sort_order"`
}

// QuickstartConfig representa la configuración para el quickstart
type QuickstartConfig struct {
	BusinessType       string   `json:"business_type"`
	StoreSize          string   `json:"store_size"`
	SelectedCategories []string `json:"selected_categories"`

	// Configuración opcional
	SelectedAttributes   []string `json:"selected_attributes,omitempty"`
	SelectedVariants     []string `json:"selected_variants,omitempty"`
	CreateSampleProducts bool     `json:"create_sample_products"`
}

// QuickstartTemplate representa un template de configuración rápida
type QuickstartTemplate struct {
	BusinessType string `json:"business_type"`
	Name         string `json:"name"`
	Description  string `json:"description"`

	// Configuración por defecto
	DefaultCategories []string `json:"default_categories"`
	DefaultAttributes []string `json:"default_attributes"`
	DefaultVariants   []string `json:"default_variants"`
	SampleProducts    []string `json:"sample_products"`

	// Configuración por tamaño de tienda
	ConfigBySize map[string]*SizeConfig `json:"config_by_size"`
}

// SizeConfig representa configuración específica por tamaño de tienda
type SizeConfig struct {
	MaxCategories int      `json:"max_categories"`
	MaxProducts   int      `json:"max_products"`
	MaxVariants   int      `json:"max_variants"`
	Features      []string `json:"features"`
}

// QuickstartResponse representa la respuesta del quickstart
type QuickstartResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	TenantID   string `json:"tenant_id"`
	TemplateID string `json:"template_id"`

	// Estadísticas de configuración
	CategoriesCreated int `json:"categories_created"`
	AttributesCreated int `json:"attributes_created"`
	VariantsCreated   int `json:"variants_created"`
	ProductsCreated   int `json:"products_created"`

	// URLs de recursos creados
	CatalogURL   string `json:"catalog_url"`
	DashboardURL string `json:"dashboard_url"`
}

// InitialProductsConfig representa configuración para productos iniciales
type InitialProductsConfig struct {
	Categories          []string `json:"categories"`
	ProductsPerCategory int      `json:"products_per_category"`
	IncludeVariants     bool     `json:"include_variants"`
	IncludeImages       bool     `json:"include_images"`
	IncludePricing      bool     `json:"include_pricing"`
}

// ProductSetupResponse representa la respuesta de configuración de productos
type ProductSetupResponse struct {
	Success         bool     `json:"success"`
	Message         string   `json:"message"`
	ProductsCreated int      `json:"products_created"`
	CategoriesUsed  []string `json:"categories_used"`

	// Información de los productos creados
	ProductIDs []string `json:"product_ids"`

	// URLs útiles
	ProductsURL string `json:"products_url"`
	CatalogURL  string `json:"catalog_url"`
}
