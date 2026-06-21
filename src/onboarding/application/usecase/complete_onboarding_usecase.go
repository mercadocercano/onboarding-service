package usecase

import (
	"context"
	"time"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/application/response"
	"onboarding/src/onboarding/domain/entity"
	"onboarding/src/onboarding/domain/port"

	"github.com/google/uuid"
)

// CompleteOnboardingUseCase maneja la finalización del proceso de onboarding
type CompleteOnboardingUseCase struct {
	onboardingRepo     port.OnboardingRepository
	notificationClient port.NotificationClient
	iamClient          port.IAMClient
	tenantClient       port.TenantClient
	eventPublisher     port.EventPublisher // opcional: si está, el welcome va por evento; si no, HTTP.
	logger             port.OnboardingEventLogger
}

// WithEventPublisher inyecta el publisher del EventBus (Plan F1). Nil-safe: si no se llama,
// el welcome email se manda por HTTP sincrónico (comportamiento legacy).
func (uc *CompleteOnboardingUseCase) WithEventPublisher(p port.EventPublisher) *CompleteOnboardingUseCase {
	uc.eventPublisher = p
	return uc
}

// NewCompleteOnboardingUseCase crea una nueva instancia del caso de uso
func NewCompleteOnboardingUseCase(
	onboardingRepo port.OnboardingRepository,
	notificationClient port.NotificationClient,
	iamClient port.IAMClient,
	tenantClient port.TenantClient,
	logger ...port.OnboardingEventLogger,
) *CompleteOnboardingUseCase {
	uc := &CompleteOnboardingUseCase{
		onboardingRepo:     onboardingRepo,
		notificationClient: notificationClient,
		iamClient:          iamClient,
		tenantClient:       tenantClient,
	}
	if len(logger) > 0 && logger[0] != nil {
		uc.logger = logger[0]
	}
	return uc
}

func (uc *CompleteOnboardingUseCase) log(e port.OnboardingEvent) {
	if uc.logger != nil {
		uc.logger.Log(e)
	}
}

// Execute ejecuta el caso de uso de completar onboarding
func (uc *CompleteOnboardingUseCase) Execute(req *request.CompleteOnboardingRequest) (*response.CompleteOnboardingResponse, error) {
	// 1. Validar request
	if err := req.Validate(); err != nil {
		return response.NewCompleteOnboardingErrorResponse(err.Error()), nil
	}

	// 2. Obtener el proceso de onboarding
	processID, err := uuid.Parse(req.ProcessID)
	if err != nil {
		return response.NewCompleteOnboardingErrorResponse("ID de proceso inválido"), nil
	}

	process, err := uc.onboardingRepo.GetProcessByID(processID)
	if err != nil {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.completion_failed",
			ProcessID: req.ProcessID,
			Reason:    "error getting process: " + err.Error(),
		})
		return response.NewCompleteOnboardingErrorResponse("Proceso de onboarding no encontrado"), err
	}

	// 3. Validar que el proceso esté en el paso correcto (paso 6)
	if process.CurrentStepNumber != 6 {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.completion_failed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
			Step:      process.CurrentStepNumber,
			Reason:    "process not at step 6",
		})
		return response.NewCompleteOnboardingErrorResponse("El proceso no está en el paso final"), nil
	}

	// 4. Manejar proceso ya completado
	if process.IsCompleted {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.already_completed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
		})

		uc.sendWelcomeEmailAsync(process)

		return response.NewCompleteOnboardingResponse(
			req.ProcessID,
			process.TenantID.String(),
			process.CompanyName,
			process.BusinessType,
			process.StoreSize,
			process.SelectedCategories,
			process.StepsCompleted,
			process.StartedAt,
			"",
		), nil
	}

	// 5. Marcar el proceso como completado
	now := time.Now()
	process.IsCompleted = true
	process.CompletedAt = &now

	process.CompleteStep(6)
	process.StepsPending = []int{}

	// 6. Guardar el proceso actualizado
	err = uc.onboardingRepo.UpdateProcess(process)
	if err != nil {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.completion_failed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
			Step:      6,
			Reason:    "error updating process: " + err.Error(),
		})
		return response.NewCompleteOnboardingErrorResponse("Error al completar el proceso"), err
	}

	uc.log(port.OnboardingEvent{
		Event:     "onboarding.completed",
		TenantID:  process.TenantID.String(),
		UserID:    process.UserID.String(),
		ProcessID: req.ProcessID,
		Step:      6,
	})

	// 7. Bootstrap tenant config (best-effort, no bloquea onboarding)
	uc.bootstrapTenantConfigAsync(process)

	// 8. Enviar correo de bienvenida
	uc.sendWelcomeEmailAsync(process)

	// 9. Generar access token para el usuario (si el usuario no lo tiene ya)
	var accessToken string
	_, err = uc.iamClient.GetUser(process.UserID.String())
	if err != nil {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.auto_login_failed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
			Reason:    "could not get user for token generation: " + err.Error(),
		})
		accessToken = ""
	} else {
		// Token ya debería haber sido generado en el paso de registro.
		accessToken = ""
	}

	// 10. Crear respuesta exitosa con resumen
	return response.NewCompleteOnboardingResponse(
		req.ProcessID,
		process.TenantID.String(),
		process.CompanyName,
		process.BusinessType,
		process.StoreSize,
		process.SelectedCategories,
		process.StepsCompleted,
		process.StartedAt,
		accessToken,
	), nil
}

// bootstrapTenantConfigAsync inicializa la configuración del tenant de forma asíncrona
// Esta operación es best-effort: no bloquea el onboarding si falla
func (uc *CompleteOnboardingUseCase) bootstrapTenantConfigAsync(process *entity.OnboardingProcess) {
	if uc.tenantClient == nil {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				uc.log(port.OnboardingEvent{
					Event:    "onboarding.tenant_config_bootstrap_failed",
					TenantID: process.TenantID.String(),
					UserID:   process.UserID.String(),
					Reason:   "panic recovered in goroutine",
				})
			}
		}()

		ctx := context.Background()
		err := uc.tenantClient.BootstrapTenantConfig(ctx, process.TenantID.String())
		if err != nil {
			uc.log(port.OnboardingEvent{
				Event:    "onboarding.tenant_config_bootstrap_failed",
				TenantID: process.TenantID.String(),
				UserID:   process.UserID.String(),
				Reason:   err.Error(),
			})
		} else {
			uc.log(port.OnboardingEvent{
				Event:    "onboarding.tenant_config_bootstrapped",
				TenantID: process.TenantID.String(),
				UserID:   process.UserID.String(),
			})
		}
	}()
}

// sendWelcomeEmailAsync envía el correo de bienvenida de forma asíncrona
func (uc *CompleteOnboardingUseCase) sendWelcomeEmailAsync(process *entity.OnboardingProcess) {
	// Necesitamos iamClient para obtener email/nombre, y al menos un canal de entrega
	// (evento o HTTP) para mandar el welcome.
	if uc.iamClient == nil || (uc.notificationClient == nil && uc.eventPublisher == nil) {
		return
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				uc.log(port.OnboardingEvent{
					Event:    "onboarding.welcome_email_send_failed",
					TenantID: process.TenantID.String(),
					UserID:   process.UserID.String(),
					Reason:   "panic recovered in goroutine",
				})
			}
		}()

		ctx := context.Background()

		user, err := uc.iamClient.GetUser(process.UserID.String())
		if err != nil {
			uc.log(port.OnboardingEvent{
				Event:    "onboarding.welcome_email_user_fetch_failed",
				TenantID: process.TenantID.String(),
				UserID:   process.UserID.String(),
				Reason:   err.Error(),
			})
			return
		}

		if user == nil || user.Email == "" {
			// Sin PII — solo logueamos user_id, no el email
			uc.log(port.OnboardingEvent{
				Event:    "onboarding.welcome_email_send_failed",
				TenantID: process.TenantID.String(),
				UserID:   process.UserID.String(),
				Reason:   "user email empty or user not found",
			})
			return
		}

		// Ingestión event-driven (Plan F1): si hay publisher, publicamos el evento y
		// notification-service lo consume; si no, caemos al HTTP sincrónico legacy.
		if uc.eventPublisher != nil {
			err = uc.eventPublisher.PublishTenantRegistered(ctx, port.TenantRegisteredEvent{
				TenantID:     process.TenantID.String(),
				UserID:       process.UserID.String(),
				Recipient:    user.Email,
				Name:         user.Name,
				Company:      process.CompanyName,
				BusinessType: process.BusinessType,
			})
		} else {
			err = uc.notificationClient.SendWelcomeEmail(ctx, user.Email, user.Name, process.CompanyName, process.BusinessType)
		}
		if err != nil {
			uc.log(port.OnboardingEvent{
				Event:    "onboarding.welcome_email_send_failed",
				TenantID: process.TenantID.String(),
				UserID:   process.UserID.String(),
				Reason:   err.Error(),
			})
		} else {
			uc.log(port.OnboardingEvent{
				Event:    "onboarding.welcome_email_sent",
				TenantID: process.TenantID.String(),
				UserID:   process.UserID.String(),
			})
		}
	}()
}
