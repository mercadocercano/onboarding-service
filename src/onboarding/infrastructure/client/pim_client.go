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

// PIMHTTPClient implementa PIMClient usando HTTP
type PIMHTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewPIMClient crea una nueva instancia del cliente PIM
func NewPIMClient() port.PIMClient {
	baseURL := getEnvPIM("PIM_SERVICE_URL", "http://localhost:8090")

	return &PIMHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetBusinessTypes obtiene todos los tipos de negocio
func (c *PIMHTTPClient) GetBusinessTypes() ([]*port.BusinessType, error) {
	url := fmt.Sprintf("%s/api/v1/quickstart/business-types", c.baseURL)

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

	var businessTypes []*port.BusinessType
	if err := json.Unmarshal(body, &businessTypes); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	log.Printf("Retrieved %d business types from PIM service", len(businessTypes))
	return businessTypes, nil
}

// GetBusinessType obtiene un tipo de negocio específico
func (c *PIMHTTPClient) GetBusinessType(businessTypeID string) (*port.BusinessType, error) {
	url := fmt.Sprintf("%s/api/v1/quickstart/business-types/%s", c.baseURL, businessTypeID)

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
	url := fmt.Sprintf("%s/api/v1/quickstart/apply", c.baseURL)

	requestData := struct {
		TenantID string                 `json:"tenant_id"`
		Config   *port.QuickstartConfig `json:"config"`
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
	// TODO: Agregar header de tenant-id cuando esté implementado
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

	var quickstartResponse port.QuickstartResponse
	if err := json.Unmarshal(body, &quickstartResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	log.Printf("Quickstart template applied successfully for tenant %s", tenantID)
	return &quickstartResponse, nil
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

// getEnvPIM obtiene una variable de entorno con valor por defecto
func getEnvPIM(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
