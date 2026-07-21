package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	appport "onboarding/src/onboarding/application/port"
	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/application/response"
	"onboarding/src/onboarding/domain/entity"
	domainport "onboarding/src/onboarding/domain/port"
)

// RegisterUserUseCase maneja el registro completo de usuario
type RegisterUserUseCase struct {
	onboardingRepo     domainport.OnboardingRepository
	iamClient          domainport.IAMClient
	notificationClient domainport.NotificationClient
	codeGenerator      appport.VerificationCodeGenerator
	logger             domainport.OnboardingEventLogger
}

// NewRegisterUserUseCase crea una nueva instancia del caso de uso
func NewRegisterUserUseCase(
	onboardingRepo domainport.OnboardingRepository,
	iamClient domainport.IAMClient,
	notificationClient domainport.NotificationClient,
	codeGenerator appport.VerificationCodeGenerator,
	logger ...domainport.OnboardingEventLogger,
) *RegisterUserUseCase {
	uc := &RegisterUserUseCase{
		onboardingRepo:     onboardingRepo,
		iamClient:          iamClient,
		notificationClient: notificationClient,
		codeGenerator:      codeGenerator,
	}
	if len(logger) > 0 && logger[0] != nil {
		uc.logger = logger[0]
	}
	return uc
}

func (uc *RegisterUserUseCase) log(e domainport.OnboardingEvent) {
	if uc.logger != nil {
		uc.logger.Log(e)
	}
}

// Execute ejecuta el proceso completo de registro de usuario
func (uc *RegisterUserUseCase) Execute(ctx context.Context, req *request.RegisterUserRequest) (*response.RegisterUserResponse, error) {
	// 1. Validar request
	if err := req.Validate(); err != nil {
		return response.NewRegisterUserErrorResponse(err.Error()), nil
	}

	// 2. Obtener rol TENANT_ADMIN primero
	tenantAdminRole, err := uc.iamClient.GetRoleByType("TENANT_ADMIN")
	if err != nil {
		uc.log(domainport.OnboardingEvent{
			Event:  "onboarding.tenant_registration_failed",
			Reason: "error getting tenant admin role: " + err.Error(),
		})
		return response.NewRegisterUserErrorResponse("Error en la configuración de permisos"), err
	}

	// 3. Crear tenant temporal con owner temporal
	tenant, err := uc.createTenant(req)
	if err != nil {
		uc.log(domainport.OnboardingEvent{
			Event:  "onboarding.tenant_registration_failed",
			Reason: "error creating tenant: " + err.Error(),
		})
		return response.NewRegisterUserErrorResponse("Error al crear la organización"), err
	}

	// 4. Crear usuario con rol de administrador
	user, err := uc.createUser(req, tenant.ID, tenantAdminRole.ID)
	if err != nil {
		uc.log(domainport.OnboardingEvent{
			Event:    "onboarding.user_registration_failed",
			TenantID: tenant.ID,
			Reason:   "error creating user: " + err.Error(),
		})
		// Rollback: eliminar tenant
		uc.iamClient.DeleteTenant(tenant.ID)
		return response.NewRegisterUserErrorResponse("Error al crear el usuario"), err
	}

	// 5. Nota: El owner se actualiza automáticamente en el IAM service cuando se crea el usuario
	// No necesitamos llamar UpdateTenantOwner por separado

	// 6. Crear proceso de onboarding real
	process, err := uc.createOnboardingProcess(ctx, tenant.ID, user.ID)
	if err != nil {
		uc.log(domainport.OnboardingEvent{
			Event:    "onboarding.process_creation_failed",
			TenantID: tenant.ID,
			UserID:   user.ID,
			Reason:   err.Error(),
		})
		// No hacer rollback aquí, el usuario ya está creado exitosamente
	}

	// 7. *** Generar y enviar código de verificación ***
	verificationCode := uc.codeGenerator.Generate()
	processID := uuid.Nil
	codeTenantID := uuid.Nil
	if process != nil {
		processID = process.ID
		codeTenantID = process.TenantID
	}

	// Guardar código de verificación en la base de datos (tenant del proceso recién
	// creado server-side — D3)
	if err := uc.saveVerificationCode(ctx, codeTenantID, processID, req.GetCleanEmail(), verificationCode); err != nil {
		uc.log(domainport.OnboardingEvent{
			Event:     "onboarding.verification_code_save_failed",
			TenantID:  tenant.ID,
			UserID:    user.ID,
			ProcessID: processID.String(),
			Reason:    err.Error(),
		})
		// No fallar todo el proceso por esto
	}

	// Enviar email de verificación
	if err := uc.notificationClient.SendEmailVerification(ctx, req.GetCleanEmail(), req.GetCleanName(), verificationCode); err != nil {
		uc.log(domainport.OnboardingEvent{
			Event:     "onboarding.verification_email_send_failed",
			TenantID:  tenant.ID,
			UserID:    user.ID,
			ProcessID: processID.String(),
			Reason:    err.Error(),
		})
	} else {
		uc.log(domainport.OnboardingEvent{
			Event:     "onboarding.verification_email_sent",
			TenantID:  tenant.ID,
			UserID:    user.ID,
			ProcessID: processID.String(),
		})
	}

	// 8. Preparar respuesta exitosa
	userData := response.UserData{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  tenantAdminRole.Type,
	}

	tenantData := response.TenantData{
		ID:          tenant.ID,
		Name:        tenant.Name,
		Slug:        tenant.Slug,
		Type:        tenant.Type,
		Description: tenant.Description,
	}

	processIDStr := ""
	if process != nil {
		processIDStr = process.ID.String()
	}

	uc.log(domainport.OnboardingEvent{
		Event:     "onboarding.tenant_registered",
		TenantID:  tenant.ID,
		UserID:    user.ID,
		ProcessID: processIDStr,
		Step:      2,
	})

	// 9. Hacer login automático para obtener access_token y refresh_token
	var accessToken, refreshToken string
	loginReq := &domainport.LoginRequest{
		Email:    req.GetCleanEmail(),
		Password: req.Password,
		Provider: "LOCAL",
	}

	loginResp, err := uc.iamClient.Login(loginReq)
	if err != nil {
		uc.log(domainport.OnboardingEvent{
			Event:    "onboarding.auto_login_failed",
			TenantID: tenant.ID,
			UserID:   user.ID,
			Reason:   err.Error(),
		})
	} else {
		accessToken = loginResp.AccessToken
		refreshToken = loginResp.RefreshToken
	}

	return response.NewRegisterUserResponse(processIDStr, tenant.ID, user.ID, accessToken, refreshToken, userData, tenantData), nil
}

// saveVerificationCode guarda el código de verificación en la base de datos
func (uc *RegisterUserUseCase) saveVerificationCode(ctx context.Context, tenantID, processID uuid.UUID, email, code string) error {
	verificationCode := entity.NewVerificationCode(tenantID, processID, email, code)
	return uc.onboardingRepo.SaveVerificationCode(ctx, verificationCode)
}

// createTenant crea un tenant temporal para el usuario
func (uc *RegisterUserUseCase) createTenant(req *request.RegisterUserRequest) (*domainport.TenantResponse, error) {
	tenantName := fmt.Sprintf("Tienda de %s", req.GetCleanName())
	slug := uc.generateSlugFromName(req.GetCleanName())

	// Generar UUID temporal válido para el owner (se actualizará después)
	tempOwnerUUID := uuid.New().String()

	tenantData := &domainport.CreateTenantRequest{
		Name:        tenantName,
		Slug:        slug,
		Description: fmt.Sprintf("Tienda administrada por %s", req.GetCleanEmail()),
		Type:        "STARTUP",
		OwnerID:     tempOwnerUUID,
		Domain:      fmt.Sprintf("%s.mitienda.com", slug),
	}

	tenant, err := uc.iamClient.CreateTenant(tenantData)
	if err != nil {
		return nil, err
	}

	return tenant, nil
}

// createUser crea el usuario administrador del tenant
func (uc *RegisterUserUseCase) createUser(req *request.RegisterUserRequest, tenantID, roleID string) (*domainport.UserResponse, error) {
	userData := &domainport.CreateUserRequest{
		Email:    req.GetCleanEmail(),
		Password: req.Password,
		Name:     req.GetCleanName(),
		TenantID: tenantID,
		RoleID:   roleID,
		Provider: "LOCAL",
	}

	return uc.iamClient.CreateUser(userData)
}

// createOnboardingProcess crea el proceso de onboarding real
func (uc *RegisterUserUseCase) createOnboardingProcess(ctx context.Context, tenantID, userID string) (*entity.OnboardingProcess, error) {
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant UUID: %w", err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user UUID: %w", err)
	}

	process := entity.NewOnboardingProcess(tenantUUID, userUUID)

	// Marcar paso 1 (bienvenida) como completado automáticamente
	process.CompleteStep(1)

	// Marcar paso 2 (registro) como completado
	process.CompleteStep(2)

	// Avanzar al paso 3 (verificación)
	process.AdvanceToStep(3)

	err = uc.onboardingRepo.SaveProcess(ctx, process)
	if err != nil {
		return nil, fmt.Errorf("error saving onboarding process: %w", err)
	}

	return process, nil
}

// generateSlugFromName genera un slug único para el tenant (solo alfanumérico)
func (uc *RegisterUserUseCase) generateSlugFromName(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "")
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, ",", "")
	slug = strings.ReplaceAll(slug, "-", "")
	slug = strings.ReplaceAll(slug, "_", "")
	slug = strings.ReplaceAll(slug, "@", "")

	var result strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			result.WriteRune(char)
		}
	}

	cleanSlug := result.String()

	if len(cleanSlug) < 3 {
		cleanSlug = "store"
	}

	timestamp := time.Now().Unix()
	uniqueID := strings.ReplaceAll(uuid.New().String(), "-", "")

	uniqueSuffix := ""
	for _, char := range uniqueID {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			uniqueSuffix += string(char)
			if len(uniqueSuffix) >= 8 {
				break
			}
		}
	}

	finalSlug := fmt.Sprintf("%s%d%s", cleanSlug, timestamp, uniqueSuffix)

	if len(finalSlug) > 50 {
		finalSlug = finalSlug[:50]
	}

	return finalSlug
}
