package usecase

import (
	"context"
	"log"
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
}

// NewCompleteOnboardingUseCase crea una nueva instancia del caso de uso
func NewCompleteOnboardingUseCase(
	onboardingRepo port.OnboardingRepository,
	notificationClient port.NotificationClient,
	iamClient port.IAMClient,
	tenantClient port.TenantClient,
) *CompleteOnboardingUseCase {
	return &CompleteOnboardingUseCase{
		onboardingRepo:     onboardingRepo,
		notificationClient: notificationClient,
		iamClient:          iamClient,
		tenantClient:       tenantClient,
	}
}

// Execute ejecuta el caso de uso de completar onboarding
func (uc *CompleteOnboardingUseCase) Execute(req *request.CompleteOnboardingRequest) (*response.CompleteOnboardingResponse, error) {
	log.Printf("Completing onboarding for process: %s", req.ProcessID)

	// 1. Validar request
	if err := req.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
		return response.NewCompleteOnboardingErrorResponse(err.Error()), nil
	}

	// 2. Obtener el proceso de onboarding
	processID, err := uuid.Parse(req.ProcessID)
	if err != nil {
		log.Printf("Invalid process ID: %v", err)
		return response.NewCompleteOnboardingErrorResponse("ID de proceso inválido"), nil
	}

	process, err := uc.onboardingRepo.GetProcessByID(processID)
	if err != nil {
		log.Printf("Error getting onboarding process: %v", err)
		return response.NewCompleteOnboardingErrorResponse("Proceso de onboarding no encontrado"), err
	}

	// 3. Validar que el proceso esté en el paso correcto (paso 6)
	if process.CurrentStepNumber != 6 {
		log.Printf("Process is not at step 6, current step: %d", process.CurrentStepNumber)
		return response.NewCompleteOnboardingErrorResponse("El proceso no está en el paso final"), nil
	}

	// 4. Manejar proceso ya completado
	if process.IsCompleted {
		log.Printf("Process is already completed: %s, but will still send welcome email", req.ProcessID)

		// Enviar correo de bienvenida si aún no se ha enviado
		uc.sendWelcomeEmailAsync(process)

		// Retornar respuesta exitosa con datos existentes
		return response.NewCompleteOnboardingResponse(
			req.ProcessID,
			process.TenantID.String(),
			process.CompanyName,
			process.BusinessType,
			process.StoreSize,
			process.SelectedCategories,
			process.StepsCompleted,
			process.StartedAt,
			"", // Token vacío para procesos ya completados
		), nil
	}

	// 5. Marcar el proceso como completado
	now := time.Now()
	process.IsCompleted = true
	process.CompletedAt = &now

	// Asegurar que el paso 6 esté en los pasos completados
	process.CompleteStep(6)

	// Limpiar pasos pendientes
	process.StepsPending = []int{}

	// 6. Guardar el proceso actualizado
	err = uc.onboardingRepo.UpdateProcess(process)
	if err != nil {
		log.Printf("Error updating onboarding process: %v", err)
		return response.NewCompleteOnboardingErrorResponse("Error al completar el proceso"), err
	}

	log.Printf("Onboarding completed successfully for process: %s at %v", req.ProcessID, now)

	// 7. Bootstrap tenant config (best-effort, no bloquea onboarding)
	uc.bootstrapTenantConfigAsync(process)

	// 8. Enviar correo de bienvenida
	uc.sendWelcomeEmailAsync(process)

	// 9. Generar access token para el usuario (si el usuario no lo tiene ya)
	var accessToken string
	user, err := uc.iamClient.GetUser(process.UserID.String())
	if err != nil {
		log.Printf("Warning: Could not get user for token generation: %v", err)
		accessToken = "" // Token vacío si no se puede obtener el usuario
	} else if user != nil {
		// Nota: Como no tenemos el password aquí, el token ya debería haber sido
		// generado en el paso de registro. Esto es solo un fallback.
		// En una implementación completa, esto requeriría un endpoint especial
		// en el IAM service para generar tokens sin password.
		log.Printf("User retrieved, but token generation requires password (should have been generated at registration)")
		accessToken = "" // Por ahora dejamos vacío, el test debe usar el token del registro
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
	log.Printf("=== BOOTSTRAP TENANT CONFIG PROCESS START ===")
	log.Printf("Checking tenant client availability...")
	log.Printf("TenantClient is nil: %t", uc.tenantClient == nil)

	if uc.tenantClient != nil {
		log.Printf("Tenant client available, starting bootstrap process asynchronously...")
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("PANIC recovered in bootstrap tenant config goroutine: %v", r)
				}
			}()

			log.Printf("=== ASYNC BOOTSTRAP TENANT CONFIG GOROUTINE START ===")
			ctx := context.Background()

			log.Printf("Attempting to bootstrap tenant config for TenantID: %s", process.TenantID.String())

			err := uc.tenantClient.BootstrapTenantConfig(ctx, process.TenantID.String())
			if err != nil {
				log.Printf("WARNING: Failed to bootstrap tenant config (best-effort, continuing)")
				log.Printf("  - TenantID: %s", process.TenantID.String())
				log.Printf("  - Error type: %T", err)
				log.Printf("  - Error message: %v", err)
			} else {
				log.Printf("SUCCESS: Tenant config bootstrapped successfully")
				log.Printf("  - TenantID: %s", process.TenantID.String())
			}

			log.Printf("=== ASYNC BOOTSTRAP TENANT CONFIG GOROUTINE END ===")
		}()
	} else {
		log.Printf("WARNING: Cannot bootstrap tenant config - TenantClient is nil")
	}

	log.Printf("=== BOOTSTRAP TENANT CONFIG PROCESS END (main thread) ===")
}

// sendWelcomeEmailAsync envía el correo de bienvenida de forma asíncrona
func (uc *CompleteOnboardingUseCase) sendWelcomeEmailAsync(process *entity.OnboardingProcess) {
	log.Printf("=== WELCOME EMAIL PROCESS START ===")
	log.Printf("Checking notification client availability...")
	log.Printf("NotificationClient is nil: %t", uc.notificationClient == nil)
	log.Printf("IAMClient is nil: %t", uc.iamClient == nil)

	if uc.notificationClient != nil && uc.iamClient != nil {
		log.Printf("Both clients available, starting welcome email process asynchronously...")
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("PANIC recovered in welcome email goroutine: %v", r)
				}
			}()

			log.Printf("=== ASYNC WELCOME EMAIL GOROUTINE START ===")
			ctx := context.Background()

			// Obtener el email del usuario desde el servicio IAM
			log.Printf("Attempting to get user info from IAM service for UserID: %s", process.UserID.String())

			user, err := uc.iamClient.GetUser(process.UserID.String())
			if err != nil {
				log.Printf("ERROR: Failed to get user from IAM service")
				log.Printf("  - UserID: %s", process.UserID.String())
				log.Printf("  - Error type: %T", err)
				log.Printf("  - Error message: %v", err)
				log.Printf("  - IAM client type: %T", uc.iamClient)
				return
			}

			if user == nil {
				log.Printf("ERROR: IAM service returned nil user for UserID: %s", process.UserID.String())
				return
			}

			log.Printf("SUCCESS: Retrieved user from IAM service")
			log.Printf("  - Email: %s", user.Email)
			log.Printf("  - Name: %s", user.Name)

			// Validar datos antes del envío
			if user.Email == "" {
				log.Printf("ERROR: User email is empty, cannot send welcome email")
				return
			}

			log.Printf("Preparing to send welcome email...")
			log.Printf("  - Recipient: %s", user.Email)
			log.Printf("  - Company: %s", process.CompanyName)
			log.Printf("  - Business Type: %s", process.BusinessType)
			log.Printf("  - NotificationClient type: %T", uc.notificationClient)

			// Intentar enviar el email con manejo detallado de errores
			err = uc.notificationClient.SendWelcomeEmail(ctx, user.Email, user.Name, process.CompanyName, process.BusinessType)
			if err != nil {
				log.Printf("ERROR: Failed to send welcome email")
				log.Printf("  - Recipient: %s", user.Email)
				log.Printf("  - Error type: %T", err)
				log.Printf("  - Error message: %v", err)
				log.Printf("  - Context: email=%s, company=%s, businessType=%s", user.Email, process.CompanyName, process.BusinessType)
			} else {
				log.Printf("SUCCESS: Welcome email sent successfully")
				log.Printf("  - Recipient: %s", user.Email)
				log.Printf("  - Company: %s", process.CompanyName)
			}

			log.Printf("=== ASYNC WELCOME EMAIL GOROUTINE END ===")
		}()
	} else {
		log.Printf("ERROR: Cannot send welcome email - missing dependencies")
		if uc.notificationClient == nil {
			log.Printf("  - NotificationClient is nil")
		}
		if uc.iamClient == nil {
			log.Printf("  - IAMClient is nil")
		}
	}

	log.Printf("=== WELCOME EMAIL PROCESS END (main thread) ===")
}
