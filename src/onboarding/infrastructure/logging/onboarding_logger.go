package logging

import (
	"io"

	"onboarding/src/onboarding/domain/port"

	sharedlog "github.com/hornosg/go-shared/infrastructure/logging"
)

// OnboardingLogger implementa port.OnboardingEventLogger emitiendo una línea JSON canónica
// (ADR-001) por evento, delegando el envelope (ts/level/service/event + campos flat
// omitempty) en go-shared CanonicalLogger (>= v0.8.0). El mapeo struct→fields y las
// reglas de nivel por evento viven acá; el formato canónico es compartido por la flota.
//
// PRIVACIDAD / PII: nunca se loguea email, nombre ni datos del comercio en claro.
// Solo IDs y estados — la decisión vive en el struct OnboardingEvent (puerto).
type OnboardingLogger struct {
	canonical *sharedlog.CanonicalLogger
}

// NewOnboardingLogger crea el adapter escribiendo a stdout. El service se fija acá, nunca por-call.
func NewOnboardingLogger(service string) *OnboardingLogger {
	return &OnboardingLogger{canonical: sharedlog.NewCanonicalLogger(service)}
}

// NewOnboardingLoggerWithWriter permite inyectar un io.Writer (tests).
func NewOnboardingLoggerWithWriter(service string, w io.Writer) *OnboardingLogger {
	return &OnboardingLogger{canonical: sharedlog.NewCanonicalLoggerWithWriter(service, w)}
}

// levelFor aplica las reglas de nivel del ADR-001 por tipo de evento.
func levelFor(event string) string {
	switch event {
	// Flujo OK → info
	case "onboarding.flow_started",
		"onboarding.tenant_registered",
		"onboarding.email_verified",
		"onboarding.verification_email_sent",
		"onboarding.verification_email_resent",
		"onboarding.store_setup_completed",
		"onboarding.plan_selected",
		"onboarding.step_completed",
		"onboarding.completed",
		"onboarding.process_status_retrieved",
		"onboarding.tenant_config_bootstrapped",
		"onboarding.welcome_email_sent":
		return "info"
	// Anomalías recuperables → warn
	case "onboarding.auto_login_failed",
		"onboarding.verification_email_send_failed",
		"onboarding.verification_resend_throttled",
		"onboarding.tenant_config_bootstrap_failed",
		"onboarding.welcome_email_send_failed",
		"onboarding.welcome_email_user_fetch_failed",
		"onboarding.pim_categories_fetch_failed",
		"onboarding.pim_products_import_failed",
		"onboarding.tenant_info_update_failed",
		"onboarding.already_completed":
		return "warn"
	// Violaciones / estado inesperado → error
	case "onboarding.tenant_registration_failed",
		"onboarding.user_registration_failed",
		"onboarding.process_creation_failed",
		"onboarding.verification_code_save_failed",
		"onboarding.email_verification_failed",
		"onboarding.store_setup_failed",
		"onboarding.plan_selection_failed",
		"onboarding.completion_failed":
		return "error"
	default:
		return "info"
	}
}

func (l *OnboardingLogger) Log(e port.OnboardingEvent) {
	fields := map[string]any{
		"tenant_id":  e.TenantID,
		"user_id":    e.UserID,
		"process_id": e.ProcessID,
		"plan":       e.Plan,
		"reason":     e.Reason,
	}
	if e.Step > 0 {
		fields["step"] = e.Step
	}
	l.canonical.Emit(levelFor(e.Event), e.Event, fields)
}
