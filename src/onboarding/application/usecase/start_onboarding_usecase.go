package usecase

import (
	"log"
	"time"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/application/response"
	"onboarding/src/onboarding/domain/entity"
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

	// Crear nueva instancia de proceso de onboarding
	// Nota: Usamos uuid.Nil temporalmente, se asignarán en el paso 2 (registro)
	process := &entity.OnboardingProcess{
		ID:                uuid.New(),
		TenantID:          uuid.Nil, // Se asigna en paso 2 (registro)
		UserID:            uuid.Nil, // Se asigna en paso 2 (registro)
		CurrentStepNumber: 1,
		IsCompleted:       false,
		StepsCompleted:    []int{},
		StepsPending:      []int{1, 2, 3, 4, 5},
		StepsSkipped:      []int{},
		StartedAt:         time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Guardar el proceso en la base de datos
	if err := uc.onboardingRepo.SaveProcess(process); err != nil {
		log.Printf("Error creating onboarding process: %v", err)
		return response.NewStartOnboardingErrorResponse("Error interno del servidor", err), err
	}

	log.Printf("Onboarding process created successfully with ID: %s", process.ID.String())

	// Retornar respuesta exitosa
	return response.NewStartOnboardingSuccessResponse(process.ID.String()), nil
}
