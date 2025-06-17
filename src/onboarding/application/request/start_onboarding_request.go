package request

import "fmt"

// StartOnboardingRequest representa la solicitud para iniciar un proceso de onboarding
type StartOnboardingRequest struct {
	Source      string `json:"source,omitempty"`       // Fuente del registro (ej: "landing_page")
	UTMCampaign string `json:"utm_campaign,omitempty"` // Campaña UTM
	Referrer    string `json:"referrer,omitempty"`     // URL de referencia
	UserAgent   string `json:"user_agent,omitempty"`   // User agent del navegador
	IP          string `json:"ip,omitempty"`           // IP del usuario (se captura del request)
}

// Validate valida la solicitud de inicio de onboarding
func (r *StartOnboardingRequest) Validate() error {
	// Las validaciones son opcionales para el inicio
	// Solo verificamos que no sea malicioso

	if len(r.Source) > 100 {
		return fmt.Errorf("source demasiado largo (máximo 100 caracteres)")
	}

	if len(r.UTMCampaign) > 200 {
		return fmt.Errorf("utm_campaign demasiado largo (máximo 200 caracteres)")
	}

	if len(r.Referrer) > 500 {
		return fmt.Errorf("referrer demasiado largo (máximo 500 caracteres)")
	}

	return nil
}

// GetDefaults establece valores por defecto
func (r *StartOnboardingRequest) GetDefaults() {
	if r.Source == "" {
		r.Source = "direct"
	}
}
