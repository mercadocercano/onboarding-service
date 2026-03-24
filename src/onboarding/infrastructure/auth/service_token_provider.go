package auth

import (
	"log"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	tokenExpiry      = 1 * time.Hour
	refreshThreshold = 10 * time.Minute
)

// ServiceTokenProvider genera y renueva automáticamente un JWT de servicio
// para comunicación service-to-service con IAM.
type ServiceTokenProvider struct {
	jwtSecret     string
	staticToken   string // fallback: IAM_SUPER_ADMIN_TOKEN env var
	currentToken  string
	expiresAt     time.Time
	mu            sync.RWMutex
}

// NewServiceTokenProvider crea un provider que auto-genera tokens.
// Si staticToken != "", se usa como fallback (backward compatibility).
func NewServiceTokenProvider(jwtSecret, staticToken string) *ServiceTokenProvider {
	p := &ServiceTokenProvider{
		jwtSecret:   jwtSecret,
		staticToken: staticToken,
	}

	if jwtSecret != "" {
		if err := p.refresh(); err != nil {
			log.Printf("[SERVICE_TOKEN] Failed to generate initial token: %v", err)
		} else {
			log.Printf("[SERVICE_TOKEN] Auto-generated service token (expires in %v)", tokenExpiry)
		}
	} else if staticToken != "" {
		log.Printf("[SERVICE_TOKEN] Using static IAM_SUPER_ADMIN_TOKEN (no auto-rotation)")
	} else {
		log.Printf("[SERVICE_TOKEN] WARNING: No JWT_SECRET nor IAM_SUPER_ADMIN_TOKEN configured")
	}

	return p
}

// GetToken devuelve un token válido. Auto-refresh si está por expirar.
func (p *ServiceTokenProvider) GetToken() string {
	// Si no hay JWT_SECRET, usar token estático
	if p.jwtSecret == "" {
		return p.staticToken
	}

	p.mu.RLock()
	token := p.currentToken
	expiresAt := p.expiresAt
	p.mu.RUnlock()

	if time.Until(expiresAt) < refreshThreshold {
		if err := p.refresh(); err != nil {
			log.Printf("[SERVICE_TOKEN] Refresh failed, using current token: %v", err)
			return token
		}
		p.mu.RLock()
		token = p.currentToken
		p.mu.RUnlock()
	}

	return token
}

// IsConfigured indica si el provider tiene un token disponible.
func (p *ServiceTokenProvider) IsConfigured() bool {
	return p.jwtSecret != "" || p.staticToken != ""
}

func (p *ServiceTokenProvider) refresh() error {
	now := time.Now()
	exp := now.Add(tokenExpiry)

	claims := jwt.MapClaims{
		"user_id": "onboarding-service",
		"iss":     "iam-service",
		"iat":     now.Unix(),
		"exp":     exp.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(p.jwtSecret))
	if err != nil {
		return err
	}

	p.mu.Lock()
	p.currentToken = signed
	p.expiresAt = exp
	p.mu.Unlock()

	return nil
}
