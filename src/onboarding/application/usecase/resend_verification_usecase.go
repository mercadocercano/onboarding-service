package usecase

import (
	"context"
	"fmt"
	"log"
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
}

// NewResendVerificationUseCase crea una nueva instancia del caso de uso
func NewResendVerificationUseCase(
	onboardingRepo port.OnboardingRepository,
	notificationClient port.NotificationClient,
) *ResendVerificationUseCase {
	return &ResendVerificationUseCase{
		onboardingRepo:     onboardingRepo,
		notificationClient: notificationClient,
	}
}

// Execute ejecuta el caso de uso de reenvío de email de verificación
func (uc *ResendVerificationUseCase) Execute(req *request.ResendVerificationRequest) (*response.ResendVerificationResponse, error) {
	log.Printf("Resending verification email for process: %s", req.ProcessID)

	// Sanitizar y validar request
	req.Sanitize()
	if err := req.Validate(); err != nil {
		log.Printf("Validation error in ResendVerification: %v", err)
		return response.NewResendVerificationErrorResponse("Datos de solicitud inválidos", err), nil
	}

	// Convertir ProcessID a UUID
	processUUID, err := uuid.Parse(req.ProcessID)
	if err != nil {
		log.Printf("Invalid process ID format: %v", err)
		return response.NewResendVerificationErrorResponse("ID de proceso inválido", err), nil
	}

	// Obtener proceso de onboarding
	process, err := uc.onboardingRepo.GetProcessByID(processUUID)
	if err != nil {
		log.Printf("Error getting process: %v", err)
		return response.NewResendVerificationErrorResponse("Proceso no encontrado", err), nil
	}

	if process == nil {
		log.Printf("Process not found: %s", req.ProcessID)
		return response.NewResendVerificationErrorResponse("Proceso no encontrado", fmt.Errorf("process not found")), nil
	}

	// Verificar que el proceso esté en el paso correcto (debe estar en paso 3 - verificación)
	if process.CurrentStepNumber != 3 {
		log.Printf("Process is not in verification step. Current step: %d", process.CurrentStepNumber)
		return response.NewResendVerificationErrorResponse("El proceso no está en el paso de verificación", fmt.Errorf("invalid step")), nil
	}

	// Verificar throttling: obtener el último código enviado
	existingCode, err := uc.onboardingRepo.GetVerificationCodeByProcessID(processUUID)
	if err != nil {
		log.Printf("Error getting existing verification code: %v", err)
		// Continuar sin error, ya que es para throttling
	}

	// Aplicar throttling si existe un código reciente
	if existingCode != nil {
		timeSinceCreated := time.Since(existingCode.CreatedAt)
		minWaitTime := 1 * time.Minute // Esperar mínimo 1 minuto entre reenvíos

		if timeSinceCreated < minWaitTime {
			waitSeconds := int(minWaitTime.Seconds() - timeSinceCreated.Seconds())
			log.Printf("Throttling resend request. Wait %d seconds", waitSeconds)
			return response.NewResendVerificationThrottleResponse(req.ProcessID, waitSeconds), nil
		}
	}

	// Generar nuevo código de verificación
	newVerificationCode := client.GenerateVerificationCode()

	// Invalidar código anterior si existe
	if existingCode != nil {
		existingCode.MarkAsUsed()
		if err := uc.onboardingRepo.UpdateVerificationCode(existingCode); err != nil {
			log.Printf("Error invalidating old verification code: %v", err)
			// No fallar por esto, solo logear
		}
	}

	// Crear y guardar nuevo código
	verificationCodeEntity := entity.NewVerificationCode(processUUID, req.Email, newVerificationCode)
	if err := uc.onboardingRepo.SaveVerificationCode(verificationCodeEntity); err != nil {
		log.Printf("Error saving new verification code: %v", err)
		return response.NewResendVerificationErrorResponse("Error interno del servidor", err), err
	}

	// Obtener nombre del usuario del proceso (simplificado por ahora)
	userName := "Usuario" // En un escenario real, obtendrías esto del proceso o usuario

	// Enviar email de verificación
	ctx := context.Background()
	if err := uc.notificationClient.SendEmailVerification(ctx, req.Email, userName, newVerificationCode); err != nil {
		log.Printf("Error sending verification email: %v", err)
		return response.NewResendVerificationErrorResponse("Error al enviar email de verificación", err), err
	}

	log.Printf("Verification email resent successfully. Process: %s", req.ProcessID)

	// Retornar respuesta exitosa
	return response.NewResendVerificationSuccessResponse(req.ProcessID), nil
}
