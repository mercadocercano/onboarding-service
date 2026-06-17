package usecase

import (
	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/application/response"
	"onboarding/src/onboarding/domain/port"

	"github.com/google/uuid"
)

// SelectPlanUseCase maneja la selección de plan en el onboarding
type SelectPlanUseCase struct {
	onboardingRepo port.OnboardingRepository
	logger         port.OnboardingEventLogger
}

// NewSelectPlanUseCase crea una nueva instancia del caso de uso
func NewSelectPlanUseCase(onboardingRepo port.OnboardingRepository, logger ...port.OnboardingEventLogger) *SelectPlanUseCase {
	uc := &SelectPlanUseCase{onboardingRepo: onboardingRepo}
	if len(logger) > 0 && logger[0] != nil {
		uc.logger = logger[0]
	}
	return uc
}

func (uc *SelectPlanUseCase) log(e port.OnboardingEvent) {
	if uc.logger != nil {
		uc.logger.Log(e)
	}
}

// Execute ejecuta el caso de uso de selección de plan
func (uc *SelectPlanUseCase) Execute(req *request.SelectPlanRequest) (*response.SelectPlanResponse, error) {
	// 1. Validar request
	if err := req.Validate(); err != nil {
		return response.NewSelectPlanErrorResponse(err.Error()), nil
	}

	// 2. Obtener el proceso de onboarding
	processID, err := uuid.Parse(req.ProcessID)
	if err != nil {
		return response.NewSelectPlanErrorResponse("ID de proceso inválido"), nil
	}

	process, err := uc.onboardingRepo.GetProcessByID(processID)
	if err != nil {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.plan_selection_failed",
			ProcessID: req.ProcessID,
			Reason:    "error getting process: " + err.Error(),
		})
		return response.NewSelectPlanErrorResponse("Proceso de onboarding no encontrado"), err
	}

	// 3. Validar que el proceso esté en el paso correcto (paso 5)
	if process.CurrentStepNumber != 5 {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.plan_selection_failed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
			Step:      process.CurrentStepNumber,
			Reason:    "process not at step 5",
		})
		return response.NewSelectPlanErrorResponse("El proceso no está en el paso de selección de plan"), nil
	}

	// 4. Marcar paso 5 como completado y avanzar al paso 6 (completar)
	process.CompleteStep(5)
	process.AdvanceToStep(6)

	// 5. Guardar el proceso actualizado
	err = uc.onboardingRepo.UpdateProcess(process)
	if err != nil {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.plan_selection_failed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
			Plan:      req.SelectedPlan,
			Reason:    "error updating process: " + err.Error(),
		})
		return response.NewSelectPlanErrorResponse("Error al actualizar el proceso"), err
	}

	uc.log(port.OnboardingEvent{
		Event:     "onboarding.plan_selected",
		TenantID:  process.TenantID.String(),
		UserID:    process.UserID.String(),
		ProcessID: req.ProcessID,
		Step:      process.CurrentStepNumber,
		Plan:      req.SelectedPlan,
	})

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
