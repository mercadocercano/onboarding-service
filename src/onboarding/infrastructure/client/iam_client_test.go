package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestIAMClient_T6_S2SSoloPorAPIKey: desde ACC-E03.T6 el S2S a IAM se autentica sólo con
// X-API-Key. Si volviera a viajar un Bearer, sería un JWT de servicio firmado por onboarding
// (HS256 con JWT_SECRET e iss=iam-service forjado): exactamente lo que el corte retiró.
func TestIAMClient_T6_S2SSoloPorAPIKey(t *testing.T) {
	var gotKey, gotAuth, gotTenant string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotKey = r.Header.Get("X-API-Key")
		gotAuth = r.Header.Get("Authorization")
		gotTenant = r.Header.Get("X-Tenant-ID")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"id":"r1","name":"owner","type":"tenant"}]}`))
	}))
	defer srv.Close()

	t.Setenv("IAM_SERVICE_URL", srv.URL)
	t.Setenv("S2S_KEY_ONBOARDING", "s2s-onboarding-test")
	t.Setenv("JWT_SECRET", "dev-secret-key-change-me-in-production-min-32-chars")
	t.Setenv("IAM_SUPER_ADMIN_TOKEN", "static-god-token")

	if _, err := NewIAMClient().GetRoleByType("tenant"); err != nil {
		t.Fatalf("GetRoleByType: %v", err)
	}
	if gotKey != "s2s-onboarding-test" {
		t.Errorf("X-API-Key = %q, want la key S2S de onboarding", gotKey)
	}
	if gotAuth != "" {
		t.Errorf("Authorization = %q: el S2S a IAM no debe llevar ningún Bearer", gotAuth)
	}
	if gotTenant == "" {
		t.Error("X-Tenant-ID vacío: el S2S debe seguir mandando el tenant de sistema")
	}
}
