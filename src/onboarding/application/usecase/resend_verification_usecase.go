package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/application/response"
	"onboarding/src/onboarding/domain/entity"
	"onboarding/src/onboarding/domain/port"
	"onboarding/src/onboarding/infrastructure/client"
)

// ResendVerificationUseCase maneja el reenvío de códigos de verificación
type ResendVerificationUseCase struct {
	onboardingRepo     port.OnboardingRepository
	notificationClient port.NotificationClient
	logger             port.OnboardingEventLogger
}

// NewResendVerificationUseCase crea una nueva instancia del caso de uso
func NewResendVerificationUseCase(
	onboardingRepo port.OnboardingRepository,
	notificationClient port.NotificationClient,
	logger ...port.OnboardingEventLogger,
) *ResendVerificationUseCase {
	uc := &ResendVerificationUseCase{
		onboardingRepo:     onboardingRepo,
		notificationClient: notificationClient,
	}
	if len(logger) > 0 && logger[0] != nil {
		uc.logger = logger[0]
	}
	return uc
}

func (uc *ResendVerificationUseCase) log(e port.OnboardingEvent) {
	if uc.logger != nil {
		uc.logger.Log(e)
	}
}

// Execute ejecuta el caso de uso de reenvío de email de verificación
func (uc *ResendVerificationUseCase) Execute(req *request.ResendVerificationRequest) (*response.ResendVerificationResponse, error) {
	// Sanitizar y validar request
	req.Sanitize()
	if err := req.Validate(); err != nil {
		return response.NewResendVerificationErrorResponse("Datos de solicitud inválidos", err), nil
	}

	// Convertir ProcessID a UUID
	processUUID, err := uuid.Parse(req.ProcessID)
	if err != nil {
		return response.NewResendVerificationErrorResponse("ID de proceso inválido", err), nil
	}

	// Obtener proceso de onboarding
	process, err := uc.onboardingRepo.GetProcessByID(processUUID)
	if err != nil {
		return response.NewResendVerificationErrorResponse("Proceso no encontrado", err), nil
	}

	if process == nil {
		return response.NewResendVerificationErrorResponse("Proceso no encontrado", fmt.Errorf("process not found")), nil
	}

	// Verificar que el proceso esté en el paso correcto (debe estar en paso 3 - verificación)
	if process.CurrentStepNumber != 3 {
		return response.NewResendVerificationErrorResponse("El proceso no está en el paso de verificación", fmt.Errorf("invalid step")), nil
	}

	// Verificar throttling: obtener el último código enviado
	existingCode, err := uc.onboardingRepo.GetVerificationCodeByProcessID(processUUID)
	if err != nil {
		// Continuar sin error, ya que es para throttling
		_ = err
	}

	// Aplicar throttling si existe un código reciente
	if existingCode != nil {
		timeSinceCreated := time.Since(existingCode.CreatedAt)
		minWaitTime := 1 * time.Minute

		if timeSinceCreated < minWaitTime {
			waitSeconds := int(minWaitTime.Seconds() - timeSinceCreated.Seconds())
			uc.log(port.OnboardingEvent{
				Event:     "onboarding.verification_resend_throttled",
				TenantID:  process.TenantID.String(),
				UserID:    process.UserID.String(),
				ProcessID: req.ProcessID,
				Step:      3,
				Reason:    fmt.Sprintf("throttled, wait %d seconds", waitSeconds),
			})
			return response.NewResendVerificationThrottleResponse(req.ProcessID, waitSeconds), nil
		}
	}

	// Generar nuevo código de verificación
	newVerificationCode := client.GenerateVerificationCode()

	// Invalidar código anterior si existe
	if existingCode != nil {
		existingCode.MarkAsUsed()
		if err := uc.onboardingRepo.UpdateVerificationCode(existingCode); err != nil {
			// No fallar por esto
			_ = err
		}
	}

	// Crear y guardar nuevo código
	verificationCodeEntity := entity.NewVerificationCode(processUUID, req.Email, newVerificationCode)
	if err := uc.onboardingRepo.SaveVerificationCode(verificationCodeEntity); err != nil {
		return response.NewResendVerificationErrorResponse("Error interno del servidor", err), err
	}

	// Obtener nombre del usuario del proceso (simplificado por ahora)
	userName := "Usuario"

	// Enviar email de verificación
	ctx := context.Background()
	if err := uc.notificationClient.SendEmailVerification(ctx, req.Email, userName, newVerificationCode); err != nil {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.verification_email_send_failed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
			Step:      3,
			Reason:    err.Error(),
		})
		return response.NewResendVerificationErrorResponse("Error al enviar email de verificación", err), err
	}

	uc.log(port.OnboardingEvent{
		Event:     "onboarding.verification_email_resent",
		TenantID:  process.TenantID.String(),
		UserID:    process.UserID.String(),
		ProcessID: req.ProcessID,
		Step:      3,
	})

	return response.NewResendVerificationSuccessResponse(req.ProcessID), nil
}
