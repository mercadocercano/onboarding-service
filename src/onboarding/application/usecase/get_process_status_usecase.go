package usecase

import (
	"github.com/google/uuid"

	"onboarding/src/onboarding/application/response"
	"onboarding/src/onboarding/domain/port"
)

// GetProcessStatusUseCase maneja la obtención del estado del proceso de onboarding
type GetProcessStatusUseCase struct {
	onboardingRepo port.OnboardingRepository
	logger         port.OnboardingEventLogger
}

// NewGetProcessStatusUseCase crea una nueva instancia del caso de uso
func NewGetProcessStatusUseCase(onboardingRepo port.OnboardingRepository, logger ...port.OnboardingEventLogger) *GetProcessStatusUseCase {
	uc := &GetProcessStatusUseCase{onboardingRepo: onboardingRepo}
	if len(logger) > 0 && logger[0] != nil {
		uc.logger = logger[0]
	}
	return uc
}

func (uc *GetProcessStatusUseCase) log(e port.OnboardingEvent) {
	if uc.logger != nil {
		uc.logger.Log(e)
	}
}

// Execute obtiene el estado actual del proceso de onboarding
func (uc *GetProcessStatusUseCase) Execute(processIDStr string) (*response.GetProcessStatusResponse, error) {
	// 1. Validar y parsear process ID
	processID, err := uuid.Parse(processIDStr)
	if err != nil {
		return response.NewGetProcessStatusErrorResponse("ID de proceso inválido"), nil
	}

	// 2. Obtener proceso desde la base de datos
	process, err := uc.onboardingRepo.GetProcessByID(processID)
	if err != nil {
		return response.NewGetProcessStatusErrorResponse("Proceso de onboarding no encontrado"), err
	}

	// 3. Determinar estado del proceso
	status := "in_progress"
	if process.IsCompleted {
		status = "completed"
	}

	// 4. Calcular progreso
	progressPercent := process.GetProgress()

	uc.log(port.OnboardingEvent{
		Event:     "onboarding.process_status_retrieved",
		TenantID:  process.TenantID.String(),
		UserID:    process.UserID.String(),
		ProcessID: processIDStr,
		Step:      process.CurrentStepNumber,
	})

	// 5. Retornar respuesta con datos reales
	return response.NewGetProcessStatusResponse(
		process.ID.String(),
		status,
		process.CurrentStepNumber,
		process.StepsCompleted,
		progressPercent,
	), nil
}
