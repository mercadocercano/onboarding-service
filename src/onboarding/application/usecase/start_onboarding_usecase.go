package usecase

import (
	"context"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/application/response"
	"onboarding/src/onboarding/domain/port"

	"github.com/google/uuid"
)

// StartOnboardingUseCase maneja el inicio del proceso de onboarding
type StartOnboardingUseCase struct {
	onboardingRepo port.OnboardingRepository
	logger         port.OnboardingEventLogger
}

// NewStartOnboardingUseCase crea una nueva instancia del caso de uso
func NewStartOnboardingUseCase(onboardingRepo port.OnboardingRepository, logger ...port.OnboardingEventLogger) *StartOnboardingUseCase {
	uc := &StartOnboardingUseCase{onboardingRepo: onboardingRepo}
	if len(logger) > 0 && logger[0] != nil {
		uc.logger = logger[0]
	}
	return uc
}

func (uc *StartOnboardingUseCase) log(e port.OnboardingEvent) {
	if uc.logger != nil {
		uc.logger.Log(e)
	}
}

// Execute ejecuta el caso de uso de inicio de onboarding
func (uc *StartOnboardingUseCase) Execute(_ context.Context, req *request.StartOnboardingRequest) (*response.StartOnboardingResponse, error) {
	// Validar request
	if err := req.Validate(); err != nil {
		return response.NewStartOnboardingErrorResponse("Datos de solicitud inválidos", err), nil
	}

	// Establecer valores por defecto
	req.GetDefaults()

	// Generar un ID temporal para el flujo del frontend
	// El proceso real se creará en el paso 2 (registro)
	tempProcessID := uuid.New().String()

	uc.log(port.OnboardingEvent{
		Event:     "onboarding.flow_started",
		ProcessID: tempProcessID,
	})

	// Retornar respuesta exitosa con ID temporal
	return response.NewStartOnboardingSuccessResponse(tempProcessID), nil
}
