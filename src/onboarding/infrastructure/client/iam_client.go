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
	baseURL        string
	httpClient     *http.Client
	superToken     string // Token de super admin para operaciones privilegiadas
	systemTenantID string // Tenant de sistema para operaciones service-to-service
}

// NewIAMClient crea una nueva instancia del cliente IAM
func NewIAMClient() port.IAMClient {
	baseURL := getEnv("IAM_SERVICE_URL", "http://localhost:8080")
	superToken := getEnv("IAM_SUPER_ADMIN_TOKEN", "")
	systemTenantID := getEnv("SYSTEM_TENANT_ID", "123e4567-e89b-12d3-a456-426614174003")

	log.Printf("[IAM_CLIENT] Initialized - superToken configured: %t", superToken != "")

	return &IAMHTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		superToken:     superToken,
		systemTenantID: systemTenantID,
	}
}

// setSystemHeaders agrega X-Tenant-ID del sistema y Authorization para operaciones service-to-service
func (c *IAMHTTPClient) setSystemHeaders(req *http.Request) {
	if c.superToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.superToken)
	}
	req.Header.Set("X-Tenant-ID", c.systemTenantID)
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
	c.setSystemHeaders(req)

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

	if tenant.ID == "" {
		return nil, fmt.Errorf("IAM service returned empty tenant data")
	}

	log.Printf("[IAM_CLIENT] Tenant created: %s", tenant.ID)
	return &tenant, nil
}

// GetTenant obtiene un tenant por ID
func (c *IAMHTTPClient) GetTenant(tenantID string) (*port.TenantResponse, error) {
	url := fmt.Sprintf("%s/api/v1/tenants/%s", c.baseURL, tenantID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	c.setSystemHeaders(req)

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
	c.setSystemHeaders(req)

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
	c.setSystemHeaders(req)

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

	c.setSystemHeaders(req)

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
	c.setSystemHeaders(req)

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

	if user.ID == "" {
		return nil, fmt.Errorf("IAM service returned empty user data")
	}

	log.Printf("[IAM_CLIENT] User created: %s", user.ID)
	return &user, nil
}

// GetUser obtiene un usuario por ID
func (c *IAMHTTPClient) GetUser(userID string) (*port.UserResponse, error) {
	url := fmt.Sprintf("%s/api/v1/users/%s", c.baseURL, userID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	c.setSystemHeaders(req)

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
		return nil, fmt.Errorf("IAM service error: status %d", resp.StatusCode)
	}

	var user port.UserResponse
	if err := json.Unmarshal(body, &user); err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %w", err)
	}

	log.Printf("[IAM_CLIENT] GetUser success - ID: %s", user.ID)
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
	c.setSystemHeaders(req)

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

	c.setSystemHeaders(req)

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

	c.setSystemHeaders(req)

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

	c.setSystemHeaders(req)

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
