package port

// OnboardingEvent es el payload canónico para eventos de dominio de onboarding (ADR-001).
// Campos flat, named. Los nombres comunes (tenant_id, user_id) son idénticos al resto
// de la flota para que el LogQL cross-service funcione. Todos opcionales salvo Event.
//
// PRIVACIDAD / PII: este struct NO debe contener email, nombre, ni datos del comercio
// en claro. Solo IDs (tenant_id, user_id, process_id) y estados/pasos.
type OnboardingEvent struct {
	Event     string // <domain>.<action>_<result>, p.ej. "onboarding.tenant_registered"
	TenantID  string
	UserID    string
	ProcessID string
	Step      int    // número de paso del onboarding (1-6), 0 si no aplica
	Plan      string // plan seleccionado (p.ej. "basic", "pro"), vacío si no aplica
	Reason    string // causa del error, vacío en caso de éxito
}

// OnboardingEventLogger es el puerto para emitir eventos canónicos de onboarding.
// El código de aplicación depende de esta interfaz; el adapter (JSON a stdout,
// Loki push, etc.) la implementa. Nunca al revés.
type OnboardingEventLogger interface {
	Log(e OnboardingEvent)
}
