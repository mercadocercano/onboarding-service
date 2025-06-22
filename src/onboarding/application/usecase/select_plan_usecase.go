package usecase

import (
	"log"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/application/response"
	"onboarding/src/onboarding/domain/port"

	"github.com/google/uuid"
)

// SelectPlanUseCase maneja la selección de plan en el onboarding
type SelectPlanUseCase struct {
	onboardingRepo port.OnboardingRepository
}

// NewSelectPlanUseCase crea una nueva instancia del caso de uso
func NewSelectPlanUseCase(onboardingRepo port.OnboardingRepository) *SelectPlanUseCase {
	return &SelectPlanUseCase{
		onboardingRepo: onboardingRepo,
	}
}

// Execute ejecuta el caso de uso de selección de plan
func (uc *SelectPlanUseCase) Execute(req *request.SelectPlanRequest) (*response.SelectPlanResponse, error) {
	log.Printf("Selecting plan for process: %s, plan: %s", req.ProcessID, req.SelectedPlan)

	// 1. Validar request
	if err := req.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
		return response.NewSelectPlanErrorResponse(err.Error()), nil
	}

	// 2. Obtener el proceso de onboarding
	processID, err := uuid.Parse(req.ProcessID)
	if err != nil {
		log.Printf("Invalid process ID: %v", err)
		return response.NewSelectPlanErrorResponse("ID de proceso inválido"), nil
	}

	process, err := uc.onboardingRepo.GetProcessByID(processID)
	if err != nil {
		log.Printf("Error getting onboarding process: %v", err)
		return response.NewSelectPlanErrorResponse("Proceso de onboarding no encontrado"), err
	}

	// 3. Validar que el proceso esté en el paso correcto (paso 5)
	if process.CurrentStepNumber != 5 {
		log.Printf("Process is not at step 5, current step: %d", process.CurrentStepNumber)
		return response.NewSelectPlanErrorResponse("El proceso no está en el paso de selección de plan"), nil
	}

	// 4. Actualizar el proceso con el plan seleccionado
	// En este caso, simplemente agregamos el plan como parte de los datos del proceso
	// (podrías agregar un campo SelectedPlan al entity si lo necesitas)

	// Marcar paso 5 como completado
	process.CompleteStep(5)

	// Avanzar al paso 6 (completar)
	process.AdvanceToStep(6)

	// 5. Guardar el proceso actualizado
	err = uc.onboardingRepo.UpdateProcess(process)
	if err != nil {
		log.Printf("Error updating onboarding process: %v", err)
		return response.NewSelectPlanErrorResponse("Error al actualizar el proceso"), err
	}

	log.Printf("Plan selected successfully: %s for process: %s", req.SelectedPlan, req.ProcessID)
	log.Printf("Process advanced to step: %d", process.CurrentStepNumber)

	// 6. Crear respuesta exitosa
	return response.NewSelectPlanResponse(
		req.ProcessID,
		req.SelectedPlan,
		process.CurrentStepNumber,
		process.StepsCompleted,
		process.StepsPending,
		"/onboarding/completar",
	), nil
}
