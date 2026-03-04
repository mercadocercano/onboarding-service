package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/application/response"
	"onboarding/src/onboarding/domain/entity"
	"onboarding/src/onboarding/domain/port"
	"onboarding/src/onboarding/infrastructure/client"
)

// RegisterUserUseCase maneja el registro completo de usuario
type RegisterUserUseCase struct {
	onboardingRepo     port.OnboardingRepository
	iamClient          port.IAMClient
	notificationClient port.NotificationClient
}

// NewRegisterUserUseCase crea una nueva instancia del caso de uso
func NewRegisterUserUseCase(
	onboardingRepo port.OnboardingRepository,
	iamClient port.IAMClient,
	notificationClient port.NotificationClient,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		onboardingRepo:     onboardingRepo,
		iamClient:          iamClient,
		notificationClient: notificationClient,
	}
}

// Execute ejecuta el proceso completo de registro de usuario
func (uc *RegisterUserUseCase) Execute(req *request.RegisterUserRequest) (*response.RegisterUserResponse, error) {
	// 1. Validar request
	if err := req.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
		return response.NewRegisterUserErrorResponse(err.Error()), nil
	}

	// 2. Obtener rol TENANT_ADMIN primero
	tenantAdminRole, err := uc.iamClient.GetRoleByType("TENANT_ADMIN")
	if err != nil {
		log.Printf("Error getting tenant admin role: %v", err)
		return response.NewRegisterUserErrorResponse("Error en la configuración de permisos"), err
	}

	// 3. Crear tenant temporal con owner temporal
	tenant, err := uc.createTenant(req)
	if err != nil {
		log.Printf("Error creating tenant: %v", err)
		return response.NewRegisterUserErrorResponse("Error al crear la organización"), err
	}

	// 4. Crear usuario con rol de administrador
	user, err := uc.createUser(req, tenant.ID, tenantAdminRole.ID)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		// Rollback: eliminar tenant
		uc.iamClient.DeleteTenant(tenant.ID)
		return response.NewRegisterUserErrorResponse("Error al crear el usuario"), err
	}

	// 5. Nota: El owner se actualiza automáticamente en el IAM service cuando se crea el usuario
	// No necesitamos llamar UpdateTenantOwner por separado

	// 6. Crear proceso de onboarding real
	process, err := uc.createOnboardingProcess(tenant.ID, user.ID)
	if err != nil {
		log.Printf("Error creating onboarding process: %v", err)
		// No hacer rollback aquí, el usuario ya está creado exitosamente
		// Solo logear el error y continuar
	}

	// 7. *** NUEVO: Generar y enviar código de verificación ***
	verificationCode := client.GenerateVerificationCode()
	processID := uuid.Nil
	if process != nil {
		processID = process.ID
	}

	// Guardar código de verificación en la base de datos
	if err := uc.saveVerificationCode(processID, req.GetCleanEmail(), verificationCode); err != nil {
		log.Printf("Error saving verification code: %v", err)
		// No fallar todo el proceso por esto, solo logear
	}

	// Enviar email de verificación
	ctx := context.Background()
	if err := uc.notificationClient.SendEmailVerification(ctx, req.GetCleanEmail(), req.GetCleanName(), verificationCode); err != nil {
		log.Printf("Error sending verification email: %v", err)
		// No fallar todo el proceso, pero logear el error
		// En producción podrías querer marcar esto para reintento
	} else {
		log.Printf("Verification email sent successfully to: %s with code: %s", req.GetCleanEmail(), verificationCode)
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

	log.Printf("User registered successfully: email=%s, tenant=%s, user=%s", req.Email, tenant.ID, user.ID)

	// 9. Hacer login automático para obtener access_token y refresh_token (misma estructura que login)
	var accessToken, refreshToken string
	loginReq := &port.LoginRequest{
		Email:    req.GetCleanEmail(),
		Password: req.Password, // Usar el password del request original
		Provider: "LOCAL",
	}

	loginResp, err := uc.iamClient.Login(loginReq)
	if err != nil {
		log.Printf("Warning: Could not auto-login user after registration: %v", err)
		// No fallar el registro por esto, solo logear el warning
	} else {
		accessToken = loginResp.AccessToken
		refreshToken = loginResp.RefreshToken
		log.Printf("User auto-logged in successfully with token")
	}

	return response.NewRegisterUserResponse(processIDStr, tenant.ID, user.ID, accessToken, refreshToken, userData, tenantData), nil
}

// saveVerificationCode guarda el código de verificación en la base de datos
func (uc *RegisterUserUseCase) saveVerificationCode(processID uuid.UUID, email, code string) error {
	// Crear entidad de código de verificación
	verificationCode := entity.NewVerificationCode(processID, email, code)

	// Guardar en el repositorio
	return uc.onboardingRepo.SaveVerificationCode(verificationCode)
}

// createTenant crea un tenant temporal para el usuario
func (uc *RegisterUserUseCase) createTenant(req *request.RegisterUserRequest) (*port.TenantResponse, error) {
	tenantName := fmt.Sprintf("Tienda de %s", req.GetCleanName())
	slug := uc.generateSlugFromName(req.GetCleanName())

	// Generar UUID temporal válido para el owner (se actualizará después)
	tempOwnerUUID := uuid.New().String()

	tenantData := &port.CreateTenantRequest{
		Name:        tenantName,
		Slug:        slug,
		Description: fmt.Sprintf("Tienda administrada por %s", req.GetCleanEmail()),
		Type:        "STARTUP",                            // Default temporal, se actualiza en paso 4
		OwnerID:     tempOwnerUUID,                        // UUID válido temporal, se actualiza después de crear el usuario
		Domain:      fmt.Sprintf("%s.mitienda.com", slug), // Generar dominio único basado en slug
	}

	log.Printf("Creating tenant with data: Name=%s, Slug=%s, OwnerID=%s", tenantData.Name, tenantData.Slug, tenantData.OwnerID)

	tenant, err := uc.iamClient.CreateTenant(tenantData)
	if err != nil {
		log.Printf("Error creating tenant: %v", err)
		return nil, err
	}

	return tenant, nil
}

// createUser crea el usuario administrador del tenant
func (uc *RegisterUserUseCase) createUser(req *request.RegisterUserRequest, tenantID, roleID string) (*port.UserResponse, error) {
	userData := &port.CreateUserRequest{
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
func (uc *RegisterUserUseCase) createOnboardingProcess(tenantID, userID string) (*entity.OnboardingProcess, error) {
	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant UUID: %w", err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user UUID: %w", err)
	}

	// Crear nuevo proceso de onboarding con datos reales
	process := entity.NewOnboardingProcess(tenantUUID, userUUID)

	// Marcar paso 1 (bienvenida) como completado automáticamente
	process.CompleteStep(1)

	// Marcar paso 2 (registro) como completado
	process.CompleteStep(2)

	// Avanzar al paso 3 (verificación)
	process.AdvanceToStep(3)

	err = uc.onboardingRepo.SaveProcess(process)
	if err != nil {
		return nil, fmt.Errorf("error saving onboarding process: %w", err)
	}

	return process, nil
}

// generateSlugFromName genera un slug único para el tenant (solo alfanumérico)
func (uc *RegisterUserUseCase) generateSlugFromName(name string) string {
	// Convertir a minúsculas y limpiar caracteres especiales
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "")
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, ",", "")
	slug = strings.ReplaceAll(slug, "-", "")
	slug = strings.ReplaceAll(slug, "_", "")
	slug = strings.ReplaceAll(slug, "@", "")

	// Mantener solo caracteres alfanuméricos
	var result strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') {
			result.WriteRune(char)
		}
	}

	cleanSlug := result.String()

	// Si el slug queda muy corto, usar un slug por defecto
	if len(cleanSlug) < 3 {
		cleanSlug = "store"
	}

	// Usar timestamp + UUID para garantizar unicidad absoluta
	timestamp := time.Now().Unix()
	uniqueID := strings.ReplaceAll(uuid.New().String(), "-", "")

	// Tomar solo los primeros 8 caracteres del UUID (alfanumérico)
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

	// Limitar longitud total del slug
	if len(finalSlug) > 50 {
		finalSlug = finalSlug[:50]
	}

	return finalSlug
}
