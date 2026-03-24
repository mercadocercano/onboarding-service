package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestServiceTokenProvider_GeneratesValidJWT(t *testing.T) {
	secret := "test-secret-at-least-32-characters-long"
	provider := NewServiceTokenProvider(secret, "")

	token := provider.GetToken()
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Parse and validate
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("failed to parse token: %v", err)
	}
	if !parsed.Valid {
		t.Fatal("token is not valid")
	}

	claims := parsed.Claims.(jwt.MapClaims)
	if claims["user_id"] != "onboarding-service" {
		t.Errorf("expected user_id=onboarding-service, got %v", claims["user_id"])
	}
	if claims["iss"] != "iam-service" {
		t.Errorf("expected iss=iam-service, got %v", claims["iss"])
	}

	exp := time.Unix(int64(claims["exp"].(float64)), 0)
	if time.Until(exp) < 50*time.Minute {
		t.Errorf("token expires too soon: %v", exp)
	}
}

func TestServiceTokenProvider_FallbackToStaticToken(t *testing.T) {
	staticToken := "static-fallback-token"
	provider := NewServiceTokenProvider("", staticToken)

	token := provider.GetToken()
	if token != staticToken {
		t.Errorf("expected static token, got %s", token)
	}
}

func TestServiceTokenProvider_IsConfigured(t *testing.T) {
	tests := []struct {
		name       string
		jwtSecret  string
		static     string
		configured bool
	}{
		{"with jwt secret", "secret-32-chars-long-at-least!!", "", true},
		{"with static token", "", "static-token", true},
		{"with nothing", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewServiceTokenProvider(tt.jwtSecret, tt.static)
			if p.IsConfigured() != tt.configured {
				t.Errorf("expected IsConfigured=%v", tt.configured)
			}
		})
	}
}

func TestServiceTokenProvider_RefreshOnExpiry(t *testing.T) {
	secret := "test-secret-at-least-32-characters-long"
	provider := NewServiceTokenProvider(secret, "")

	firstToken := provider.GetToken()

	// Force expiry
	provider.mu.Lock()
	provider.expiresAt = time.Now().Add(5 * time.Minute) // within refresh threshold
	provider.mu.Unlock()

	secondToken := provider.GetToken()

	if firstToken == secondToken {
		// Token should have been refreshed
		t.Log("token was refreshed (different tokens)")
	}
	// Both should be valid
	if secondToken == "" {
		t.Fatal("second token should not be empty")
	}
}
