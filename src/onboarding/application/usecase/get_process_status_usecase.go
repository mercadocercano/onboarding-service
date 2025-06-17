package usecase

import (
	"log"

	"github.com/google/uuid"

	"onboarding/src/onboarding/application/response"
	"onboarding/src/onboarding/domain/port"
)

// GetProcessStatusUseCase maneja la obtención del estado del proceso de onboarding
type GetProcessStatusUseCase struct {
	onboardingRepo port.OnboardingRepository
}

// NewGetProcessStatusUseCase crea una nueva instancia del caso de uso
func NewGetProcessStatusUseCase(onboardingRepo port.OnboardingRepository) *GetProcessStatusUseCase {
	return &GetProcessStatusUseCase{
		onboardingRepo: onboardingRepo,
	}
}

// Execute obtiene el estado actual del proceso de onboarding
func (uc *GetProcessStatusUseCase) Execute(processIDStr string) (*response.GetProcessStatusResponse, error) {
	// 1. Validar y parsear process ID
	processID, err := uuid.Parse(processIDStr)
	if err != nil {
		log.Printf("Invalid process ID: %v", err)
		return response.NewGetProcessStatusErrorResponse("ID de proceso inválido"), nil
	}

	// 2. Obtener proceso desde la base de datos
	process, err := uc.onboardingRepo.GetProcessByID(processID)
	if err != nil {
		log.Printf("Error getting process by ID %s: %v", processIDStr, err)
		return response.NewGetProcessStatusErrorResponse("Proceso de onboarding no encontrado"), err
	}

	// 3. Determinar estado del proceso
	status := "in_progress"
	if process.IsCompleted {
		status = "completed"
	}

	// 4. Calcular progreso
	progressPercent := process.GetProgress()

	log.Printf("Process status retrieved: ID=%s, Status=%s, Progress=%.1f%%",
		processIDStr, status, progressPercent)

	// 5. Retornar respuesta con datos reales
	return response.NewGetProcessStatusResponse(
		process.ID.String(),
		status,
		process.CurrentStepNumber,
		process.StepsCompleted,
		progressPercent,
	), nil
}
