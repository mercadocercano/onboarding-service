package usecase

import (
	"log"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/application/response"
	"onboarding/src/onboarding/domain/port"

	"github.com/google/uuid"
)

// StartOnboardingUseCase maneja el inicio del proceso de onboarding
type StartOnboardingUseCase struct {
	onboardingRepo port.OnboardingRepository
}

// NewStartOnboardingUseCase crea una nueva instancia del caso de uso
func NewStartOnboardingUseCase(onboardingRepo port.OnboardingRepository) *StartOnboardingUseCase {
	return &StartOnboardingUseCase{
		onboardingRepo: onboardingRepo,
	}
}

// Execute ejecuta el caso de uso de inicio de onboarding
func (uc *StartOnboardingUseCase) Execute(req *request.StartOnboardingRequest) (*response.StartOnboardingResponse, error) {
	log.Printf("Starting onboarding process with source: %s", req.Source)

	// Validar request
	if err := req.Validate(); err != nil {
		log.Printf("Validation error in StartOnboarding: %v", err)
		return response.NewStartOnboardingErrorResponse("Datos de solicitud inválidos", err), nil
	}

	// Establecer valores por defecto
	req.GetDefaults()

	// Generar un ID temporal para el flujo del frontend
	// El proceso real se creará en el paso 2 (registro)
	tempProcessID := uuid.New().String()

	log.Printf("Onboarding flow started, temporary ID: %s", tempProcessID)

	// Retornar respuesta exitosa con ID temporal
	return response.NewStartOnboardingSuccessResponse(tempProcessID), nil
}
