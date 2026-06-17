package logging_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"onboarding/src/onboarding/domain/port"
	"onboarding/src/onboarding/infrastructure/logging"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseLogLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))
	return line
}

func TestOnboardingLogger_TenantRegistered_EmitsCanonicalEnvelope(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewOnboardingLoggerWithWriter("onboarding", &buf)

	logger.Log(port.OnboardingEvent{
		Event:     "onboarding.tenant_registered",
		TenantID:  "tenant-123",
		UserID:    "user-456",
		ProcessID: "proc-789",
		Step:      2,
	})

	line := parseLogLine(t, &buf)
	assert.Equal(t, "onboarding.tenant_registered", line["event"])
	assert.Equal(t, "info", line["level"])
	assert.Equal(t, "onboarding", line["service"])
	assert.NotEmpty(t, line["ts"])
	assert.Equal(t, "tenant-123", line["tenant_id"])
	assert.Equal(t, "user-456", line["user_id"])
	assert.Equal(t, "proc-789", line["process_id"])
	assert.EqualValues(t, 2, line["step"])
	// Sin PII: no debe haber email ni nombre
	assert.Nil(t, line["email"])
	assert.Nil(t, line["name"])
}

func TestOnboardingLogger_EmailVerified_LevelInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewOnboardingLoggerWithWriter("onboarding", &buf)

	logger.Log(port.OnboardingEvent{
		Event:     "onboarding.email_verified",
		TenantID:  "tenant-123",
		UserID:    "user-456",
		ProcessID: "proc-789",
		Step:      4,
	})

	line := parseLogLine(t, &buf)
	assert.Equal(t, "onboarding.email_verified", line["event"])
	assert.Equal(t, "info", line["level"])
}

func TestOnboardingLogger_VerificationEmailSendFailed_LevelWarn(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewOnboardingLoggerWithWriter("onboarding", &buf)

	logger.Log(port.OnboardingEvent{
		Event:    "onboarding.verification_email_send_failed",
		TenantID: "tenant-123",
		UserID:   "user-456",
		Reason:   "smtp timeout",
	})

	line := parseLogLine(t, &buf)
	assert.Equal(t, "warn", line["level"])
	assert.Equal(t, "smtp timeout", line["reason"])
}

func TestOnboardingLogger_TenantRegistrationFailed_LevelError(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewOnboardingLoggerWithWriter("onboarding", &buf)

	logger.Log(port.OnboardingEvent{
		Event:  "onboarding.tenant_registration_failed",
		Reason: "iam service unavailable",
	})

	line := parseLogLine(t, &buf)
	assert.Equal(t, "error", line["level"])
}

func TestOnboardingLogger_PlanSelected_IncludesPlan(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewOnboardingLoggerWithWriter("onboarding", &buf)

	logger.Log(port.OnboardingEvent{
		Event:     "onboarding.plan_selected",
		TenantID:  "tenant-123",
		UserID:    "user-456",
		ProcessID: "proc-789",
		Step:      6,
		Plan:      "pro",
	})

	line := parseLogLine(t, &buf)
	assert.Equal(t, "info", line["level"])
	assert.Equal(t, "pro", line["plan"])
}

func TestOnboardingLogger_OmitsEmptyFields(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewOnboardingLoggerWithWriter("onboarding", &buf)

	// Solo event, sin otros campos
	logger.Log(port.OnboardingEvent{
		Event: "onboarding.flow_started",
	})

	line := parseLogLine(t, &buf)
	assert.Equal(t, "onboarding.flow_started", line["event"])
	// Campos vacíos deben omitirse (omitempty)
	assert.Nil(t, line["tenant_id"])
	assert.Nil(t, line["user_id"])
	assert.Nil(t, line["reason"])
	assert.Nil(t, line["plan"])
	// step=0 se omite
	assert.Nil(t, line["step"])
}

func TestOnboardingLogger_StepZero_Omitted(t *testing.T) {
	var buf bytes.Buffer
	logger := logging.NewOnboardingLoggerWithWriter("onboarding", &buf)

	logger.Log(port.OnboardingEvent{
		Event:    "onboarding.completed",
		TenantID: "tenant-123",
		Step:     0, // debe omitirse
	})

	line := parseLogLine(t, &buf)
	assert.Nil(t, line["step"], "step=0 debe ser omitido del JSON")
}
