package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"onboarding/src/onboarding/domain/port"
)

// PIMHTTPClient implementa PIMClient usando HTTP (S2S)
type PIMHTTPClient struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// NewPIMClient crea una nueva instancia del cliente PIM con autenticación S2S
func NewPIMClient() port.PIMClient {
	baseURL := getEnvPIM("PIM_SERVICE_URL", "http://localhost:8090")
	apiKey := os.Getenv("S2S_API_KEY")

	return &PIMHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey: apiKey,
	}
}

// GetBusinessTypes obtiene todos los tipos de negocio
func (c *PIMHTTPClient) GetBusinessTypes() ([]*port.BusinessType, error) {
	url := fmt.Sprintf("%s/api/v1/business-types?page_size=100", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PIM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Items []*port.BusinessType `json:"items"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling business types response: %w", err)
	}

	if len(result.Items) == 0 {
		return nil, fmt.Errorf("no business types found in PIM response")
	}

	log.Printf("Retrieved %d business types from PIM service", len(result.Items))
	return result.Items, nil
}

// GetBusinessType obtiene un tipo de negocio específico
func (c *PIMHTTPClient) GetBusinessType(businessTypeID string) (*port.BusinessType, error) {
	url := fmt.Sprintf("%s/api/v1/business-types/%s", c.baseURL, businessTypeID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PIM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		BusinessType *port.BusinessType `json:"business_type"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return result.BusinessType, nil
}

// GetCategoriesByBusinessType obtiene categorías filtradas por tipo de negocio
func (c *PIMHTTPClient) GetCategoriesByBusinessType(businessType string) ([]*port.Category, error) {
	url := fmt.Sprintf("%s/api/v1/quickstart/categories/%s", c.baseURL, businessType)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PIM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Categories []*port.Category `json:"categories"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	log.Printf("Retrieved %d categories for business type %s", len(result.Categories), businessType)
	return result.Categories, nil
}

// GetCategories obtiene todas las categorías
func (c *PIMHTTPClient) GetCategories() ([]*port.Category, error) {
	url := fmt.Sprintf("%s/api/v1/categories", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PIM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Categories []*port.Category `json:"categories"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return result.Categories, nil
}

// GetCategory obtiene una categoría específica
func (c *PIMHTTPClient) GetCategory(categoryID string) (*port.Category, error) {
	url := fmt.Sprintf("%s/api/v1/categories/%s", c.baseURL, categoryID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PIM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Category *port.Category `json:"category"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return result.Category, nil
}

// ApplyQuickstartTemplate aplica un template de configuración rápida
func (c *PIMHTTPClient) ApplyQuickstartTemplate(tenantID string, config *port.QuickstartConfig) (*port.QuickstartResponse, error) {
	url := fmt.Sprintf("%s/api/v1/quickstart/setup", c.baseURL)

	requestData := struct {
		BusinessType       string   `json:"businessType"`
		SelectedCategories []string `json:"selectedCategories"`
		SelectedAttributes []string `json:"selectedAttributes,omitempty"`
		SelectedVariants   []string `json:"selectedVariants,omitempty"`
		SelectedProducts   []string `json:"selectedProducts,omitempty"`
	}{
		BusinessType:       config.BusinessType,
		SelectedCategories: config.SelectedCategories,
		SelectedAttributes: config.SelectedAttributes,
		SelectedVariants:   config.SelectedVariants,
		SelectedProducts:   []string{},
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID)

	log.Printf("Calling PIM setup endpoint: %s for tenant %s with business type %s", url, tenantID, config.BusinessType)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	log.Printf("PIM setup response status: %d, body: %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("PIM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var pimResponse struct {
		ID           string `json:"id"`
		TenantID     string `json:"tenantId"`
		BusinessType string `json:"businessType"`
		Status       string `json:"status"`
		SetupData    string `json:"setupData"`
		CreatedAt    string `json:"createdAt"`
		UpdatedAt    string `json:"updatedAt"`
	}

	if err := json.Unmarshal(body, &pimResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	quickstartResponse := &port.QuickstartResponse{
		Success:           pimResponse.Status == "COMPLETED",
		Message:           fmt.Sprintf("Business type setup completed for tenant %s", tenantID),
		TenantID:          pimResponse.TenantID,
		TemplateID:        pimResponse.BusinessType,
		CategoriesCreated: len(config.SelectedCategories),
		AttributesCreated: len(config.SelectedAttributes),
		VariantsCreated:   len(config.SelectedVariants),
		ProductsCreated:   0,
	}

	log.Printf("Quickstart template applied successfully for tenant %s with business type %s", tenantID, config.BusinessType)
	return quickstartResponse, nil
}

// GetQuickstartTemplate obtiene un template de quickstart
func (c *PIMHTTPClient) GetQuickstartTemplate(businessType string) (*port.QuickstartTemplate, error) {
	url := fmt.Sprintf("%s/api/v1/quickstart/templates/%s", c.baseURL, businessType)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("PIM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Template *port.QuickstartTemplate `json:"template"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return result.Template, nil
}

// CreateInitialProducts crea productos iniciales para un tenant
func (c *PIMHTTPClient) CreateInitialProducts(tenantID string, config *port.InitialProductsConfig) (*port.ProductSetupResponse, error) {
	url := fmt.Sprintf("%s/api/v1/quickstart/products", c.baseURL)

	requestData := struct {
		TenantID string                      `json:"tenant_id"`
		Config   *port.InitialProductsConfig `json:"config"`
	}{
		TenantID: tenantID,
		Config:   config,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("PIM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var productResponse port.ProductSetupResponse
	if err := json.Unmarshal(body, &productResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	log.Printf("Initial products created successfully for tenant %s: %d products",
		tenantID, productResponse.ProductsCreated)
	return &productResponse, nil
}

// ImportProductsFromBusinessType importa productos del catálogo global al tenant
func (c *PIMHTTPClient) ImportProductsFromBusinessType(tenantID, businessTypeCode string) (*port.ImportProductsResponse, error) {
	url := fmt.Sprintf("%s/api/v1/quickstart/products/import-from-business-type", c.baseURL)

	body := map[string]interface{}{
		"business_type_id": businessTypeCode,
		"import_all":       true,
		"initial_status":   "draft",
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenantID)

	log.Printf("Importing products from global catalog for tenant %s, business_type=%s", tenantID, businessTypeCode)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("PIM import error: %s (status: %d)", string(respBody), resp.StatusCode)
	}

	var result port.ImportProductsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	log.Printf("Imported %d products for tenant %s", result.Summary.TotalImported, tenantID)
	return &result, nil
}

// getEnvPIM obtiene una variable de entorno con valor por defecto
func getEnvPIM(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
