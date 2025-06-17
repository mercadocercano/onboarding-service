package port

// IAMClient define las operaciones de integración con el servicio IAM
type IAMClient interface {
	// Tenant operations
	CreateTenant(request *CreateTenantRequest) (*TenantResponse, error)
	GetTenant(tenantID string) (*TenantResponse, error)
	UpdateTenant(tenantID string, request *UpdateTenantRequest) (*TenantResponse, error)
	UpdateTenantOwner(tenantID, userID string) error
	DeleteTenant(tenantID string) error

	// User operations
	CreateUser(request *CreateUserRequest) (*UserResponse, error)
	GetUser(userID string) (*UserResponse, error)
	UpdateUser(userID string, request *UpdateUserRequest) (*UserResponse, error)
	DeleteUser(userID string) error

	// Role operations
	GetRoleByType(roleType string) (*RoleResponse, error)
	GetRoles() ([]*RoleResponse, error)

	// Auth operations
	Login(request *LoginRequest) (*LoginResponse, error)
	ValidateToken(token string) (*TokenValidationResponse, error)
}

// CreateTenantRequest representa una solicitud de creación de tenant
type CreateTenantRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Type        string `json:"type"`
	OwnerID     string `json:"owner_id"`
	Domain      string `json:"domain,omitempty"`
}

// UpdateTenantRequest representa una solicitud de actualización de tenant
type UpdateTenantRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`
}

// TenantResponse representa la respuesta de operaciones de tenant
type TenantResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Type        string `json:"type"`
	OwnerID     string `json:"owner_id"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// CreateUserRequest representa una solicitud de creación de usuario
type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
	TenantID string `json:"tenant_id"`
	RoleID   string `json:"role_id"`
	Provider string `json:"provider"`
}

// UpdateUserRequest representa una solicitud de actualización de usuario
type UpdateUserRequest struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	RoleID   string `json:"role_id,omitempty"`
}

// UserResponse representa la respuesta de operaciones de usuario
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	TenantID  string `json:"tenant_id"`
	RoleID    string `json:"role_id"`
	Provider  string `json:"provider"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// RoleResponse representa la respuesta de operaciones de rol
type RoleResponse struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	IsActive    bool     `json:"is_active"`
}

// LoginRequest representa una solicitud de login
type LoginRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Provider    string `json:"provider"`
	GoogleToken string `json:"google_token,omitempty"`
	TenantID    string `json:"tenant_id,omitempty"`
}

// LoginResponse representa la respuesta de login
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	User         struct {
		ID       string `json:"id"`
		Email    string `json:"email"`
		Name     string `json:"name"`
		TenantID string `json:"tenant_id"`
		Role     string `json:"role"`
	} `json:"user"`
}

// TokenValidationResponse representa la respuesta de validación de token
type TokenValidationResponse struct {
	Valid    bool   `json:"valid"`
	UserID   string `json:"user_id,omitempty"`
	TenantID string `json:"tenant_id,omitempty"`
	Role     string `json:"role,omitempty"`
	Email    string `json:"email,omitempty"`
}
