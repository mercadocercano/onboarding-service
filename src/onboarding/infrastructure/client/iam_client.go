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

// IAMHTTPClient implementa IAMClient usando HTTP
type IAMHTTPClient struct {
	baseURL    string
	httpClient *http.Client
	superToken string // Token de super admin para operaciones privilegiadas
}

// NewIAMClient crea una nueva instancia del cliente IAM
func NewIAMClient() port.IAMClient {
	baseURL := getEnv("IAM_SERVICE_URL", "http://localhost:8080")
	superToken := getEnv("IAM_SUPER_ADMIN_TOKEN", "") // Token para operaciones privilegiadas

	log.Printf("=== IAM CLIENT INITIALIZATION ===")
	log.Printf("Base URL: %s", baseURL)
	log.Printf("Super token configured: %t", superToken != "")

	return &IAMHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		superToken: superToken,
	}
}

// CreateTenant crea un nuevo tenant
func (c *IAMHTTPClient) CreateTenant(request *port.CreateTenantRequest) (*port.TenantResponse, error) {
	url := fmt.Sprintf("%s/api/v1/tenants", c.baseURL)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.superToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.superToken)
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

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IAM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var tenant port.TenantResponse
	if err := json.Unmarshal(body, &tenant); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	// Verificar que el tenant fue creado correctamente
	if tenant.ID == "" {
		log.Printf("IAM service response body: %s", string(body))
		return nil, fmt.Errorf("IAM service returned empty tenant data")
	}

	log.Printf("Tenant created successfully: %s", tenant.ID)
	return &tenant, nil
}

// GetTenant obtiene un tenant por ID
func (c *IAMHTTPClient) GetTenant(tenantID string) (*port.TenantResponse, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s", c.baseURL, tenantID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if c.superToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.superToken)
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
		return nil, fmt.Errorf("IAM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Tenant *port.TenantResponse `json:"tenant"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return result.Tenant, nil
}

// UpdateTenant actualiza un tenant
func (c *IAMHTTPClient) UpdateTenant(tenantID string, request *port.UpdateTenantRequest) (*port.TenantResponse, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s", c.baseURL, tenantID)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.superToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.superToken)
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
		return nil, fmt.Errorf("IAM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Tenant *port.TenantResponse `json:"tenant"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return result.Tenant, nil
}

// UpdateTenantOwner actualiza el propietario de un tenant
func (c *IAMHTTPClient) UpdateTenantOwner(tenantID, userID string) error {
	url := fmt.Sprintf("%s/api/v1/tenants/%s/owner", c.baseURL, tenantID)

	requestData := map[string]string{
		"owner_id": userID,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.superToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.superToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("IAM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// DeleteTenant elimina un tenant
func (c *IAMHTTPClient) DeleteTenant(tenantID string) error {
	url := fmt.Sprintf("%s/api/v1/tenants/%s", c.baseURL, tenantID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	if c.superToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.superToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("IAM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// CreateUser crea un nuevo usuario
func (c *IAMHTTPClient) CreateUser(request *port.CreateUserRequest) (*port.UserResponse, error) {
	url := fmt.Sprintf("%s/api/v1/users", c.baseURL)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.superToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.superToken)
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

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("IAM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var user port.UserResponse
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	// Verificar que el usuario fue creado correctamente
	if user.ID == "" {
		log.Printf("IAM service response body: %s", string(body))
		return nil, fmt.Errorf("IAM service returned empty user data")
	}

	log.Printf("User created successfully: %s", user.ID)
	return &user, nil
}

// GetUser obtiene un usuario por ID
func (c *IAMHTTPClient) GetUser(userID string) (*port.UserResponse, error) {
	log.Printf("=== IAM CLIENT: GetUser START ===")
	log.Printf("UserID: %s", userID)
	log.Printf("Base URL: %s", c.baseURL)

	url := fmt.Sprintf("%s/api/v1/users/%s", c.baseURL, userID)
	log.Printf("Request URL: %s", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("ERROR: Failed to create HTTP request: %v", err)
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if c.superToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.superToken)
		log.Printf("Added authorization header with super token")
	} else {
		log.Printf("No super token available")
	}

	log.Printf("Making HTTP request...")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("ERROR: HTTP request failed: %v", err)
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("HTTP response status: %d %s", resp.StatusCode, resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read response body: %v", err)
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	log.Printf("Response body: %s", string(body))
	log.Printf("Response size: %d bytes", len(body))

	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: Non-success status code received")
		log.Printf("=== IAM CLIENT: GetUser END (STATUS ERROR) ===")
		return nil, fmt.Errorf("IAM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	log.Printf("Parsing response JSON...")
	var user port.UserResponse
	if err := json.Unmarshal(body, &user); err != nil {
		log.Printf("ERROR: Failed to parse JSON response: %v", err)
		log.Printf("=== IAM CLIENT: GetUser END (PARSE ERROR) ===")
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	log.Printf("JSON parsed successfully")
	log.Printf("User found:")
	log.Printf("  - ID: %s", user.ID)
	log.Printf("  - Email: %s", user.Email)
	log.Printf("  - Name: %s", user.Name)
	log.Printf("=== IAM CLIENT: GetUser END (SUCCESS) ===")

	return &user, nil
}

// UpdateUser actualiza un usuario
func (c *IAMHTTPClient) UpdateUser(userID string, request *port.UpdateUserRequest) (*port.UserResponse, error) {
	url := fmt.Sprintf("%s/api/v1/users/%s", c.baseURL, userID)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.superToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.superToken)
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
		return nil, fmt.Errorf("IAM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		User *port.UserResponse `json:"user"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return result.User, nil
}

// DeleteUser elimina un usuario
func (c *IAMHTTPClient) DeleteUser(userID string) error {
	url := fmt.Sprintf("%s/api/v1/users/%s", c.baseURL, userID)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	if c.superToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.superToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("IAM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

// GetRoleByType obtiene un rol por tipo
func (c *IAMHTTPClient) GetRoleByType(roleType string) (*port.RoleResponse, error) {
	url := fmt.Sprintf("%s/api/v1/roles?type=%s", c.baseURL, roleType)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if c.superToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.superToken)
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
		return nil, fmt.Errorf("IAM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Items []*port.RoleResponse `json:"items"`
		Roles []*port.RoleResponse `json:"roles"` // Backward compatibility
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	// Try items first, then roles for backward compatibility
	roles := result.Items
	if len(roles) == 0 {
		roles = result.Roles
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("role type %s not found", roleType)
	}

	return roles[0], nil
}

// GetRoles obtiene todos los roles
func (c *IAMHTTPClient) GetRoles() ([]*port.RoleResponse, error) {
	url := fmt.Sprintf("%s/api/v1/roles", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	if c.superToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.superToken)
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
		return nil, fmt.Errorf("IAM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var result struct {
		Items []*port.RoleResponse `json:"items"`
		Roles []*port.RoleResponse `json:"roles"` // Backward compatibility
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	// Try items first, then roles for backward compatibility
	if len(result.Items) > 0 {
		return result.Items, nil
	}
	return result.Roles, nil
}

// Login realiza login de usuario
func (c *IAMHTTPClient) Login(request *port.LoginRequest) (*port.LoginResponse, error) {
	url := fmt.Sprintf("%s/api/v1/auth/login", c.baseURL)

	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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
		return nil, fmt.Errorf("IAM service error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var loginResponse port.LoginResponse
	if err := json.Unmarshal(body, &loginResponse); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	return &loginResponse, nil
}

// ValidateToken valida un token
func (c *IAMHTTPClient) ValidateToken(token string) (*port.TokenValidationResponse, error) {
	url := fmt.Sprintf("%s/api/v1/auth/validate", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var validationResponse port.TokenValidationResponse
	if resp.StatusCode == http.StatusOK {
		if err := json.Unmarshal(body, &validationResponse); err != nil {
			return nil, fmt.Errorf("error unmarshaling response: %w", err)
		}
		validationResponse.Valid = true
	} else {
		validationResponse.Valid = false
	}

	return &validationResponse, nil
}

// getEnv obtiene una variable de entorno con valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
