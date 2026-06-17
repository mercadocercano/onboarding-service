package usecase

import (
	"fmt"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/application/response"
	"onboarding/src/onboarding/domain/port"

	"github.com/google/uuid"
)

// VerifyEmailUseCase maneja la verificación de email
type VerifyEmailUseCase struct {
	onboardingRepo port.OnboardingRepository
	iamClient      port.IAMClient
	logger         port.OnboardingEventLogger
}

// NewVerifyEmailUseCase crea una nueva instancia del caso de uso
func NewVerifyEmailUseCase(onboardingRepo port.OnboardingRepository, iamClient port.IAMClient, logger ...port.OnboardingEventLogger) *VerifyEmailUseCase {
	uc := &VerifyEmailUseCase{
		onboardingRepo: onboardingRepo,
		iamClient:      iamClient,
	}
	if len(logger) > 0 && logger[0] != nil {
		uc.logger = logger[0]
	}
	return uc
}

func (uc *VerifyEmailUseCase) log(e port.OnboardingEvent) {
	if uc.logger != nil {
		uc.logger.Log(e)
	}
}

// Execute ejecuta el caso de uso de verificación de email
func (uc *VerifyEmailUseCase) Execute(req *request.VerifyEmailRequest) (*response.VerifyEmailResponse, error) {
	// Sanitizar y validar request
	req.Sanitize()
	if err := req.Validate(); err != nil {
		return response.NewVerifyEmailErrorResponse("Datos de solicitud inválidos", err), nil
	}

	// Convertir ProcessID a UUID
	processUUID, err := uuid.Parse(req.ProcessID)
	if err != nil {
		return response.NewVerifyEmailErrorResponse("ID de proceso inválido", err), nil
	}

	// Obtener proceso de onboarding
	process, err := uc.onboardingRepo.GetProcessByID(processUUID)
	if err != nil {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.email_verification_failed",
			ProcessID: req.ProcessID,
			Reason:    "error getting process: " + err.Error(),
		})
		return response.NewVerifyEmailErrorResponse("Proceso no encontrado", err), nil
	}

	if process == nil {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.email_verification_failed",
			ProcessID: req.ProcessID,
			Reason:    "process not found",
		})
		return response.NewVerifyEmailErrorResponse("Proceso no encontrado", fmt.Errorf("process not found")), nil
	}

	// Verificar que el proceso esté en el paso correcto (debe estar en paso 3)
	if process.CurrentStepNumber != 3 {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.email_verification_failed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
			Step:      process.CurrentStepNumber,
			Reason:    "process not in verification step",
		})
		return response.NewVerifyEmailErrorResponse("El proceso no está en el paso de verificación", fmt.Errorf("invalid step")), nil
	}

	// Verificar el código con la base de datos
	isValid, err := uc.verifyCodeWithDatabase(processUUID, req.VerificationCode)
	if err != nil {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.email_verification_failed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
			Step:      3,
			Reason:    "error verifying code: " + err.Error(),
		})
		return response.NewVerifyEmailErrorResponse("Error al verificar código", err), err
	}

	if !isValid {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.email_verification_failed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
			Step:      3,
			Reason:    "invalid verification code",
		})
		return response.NewVerifyEmailInvalidCodeResponse(req.ProcessID), nil
	}

	currentStep := process.CurrentStepNumber
	nextStep := currentStep + 1

	process.CompleteStep(currentStep)
	process.AdvanceToStep(nextStep)

	// Actualizar proceso en la base de datos
	if err := uc.onboardingRepo.UpdateProcess(process); err != nil {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.email_verification_failed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
			Step:      currentStep,
			Reason:    "error updating process: " + err.Error(),
		})
		return response.NewVerifyEmailErrorResponse("Error interno del servidor", err), err
	}

	uc.log(port.OnboardingEvent{
		Event:     "onboarding.email_verified",
		TenantID:  process.TenantID.String(),
		UserID:    process.UserID.String(),
		ProcessID: req.ProcessID,
		Step:      nextStep,
	})

	return response.NewVerifyEmailSuccessResponse(req.ProcessID), nil
}

// verifyCodeWithDatabase verifica el código de verificación con la base de datos
func (uc *VerifyEmailUseCase) verifyCodeWithDatabase(processID uuid.UUID, code string) (bool, error) {
	verificationCode, err := uc.onboardingRepo.GetVerificationCodeByProcessID(processID)
	if err != nil {
		return false, err
	}

	if verificationCode == nil {
		return false, nil
	}

	if !verificationCode.IsValid() {
		return false, nil
	}

	if verificationCode.Code != code {
		return false, nil
	}

	// Marcar código como usado
	verificationCode.MarkAsUsed()
	if err := uc.onboardingRepo.UpdateVerificationCode(verificationCode); err != nil {
		// No fallar la verificación por esto
		_ = err
	}

	return true, nil
}
