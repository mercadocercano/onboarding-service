package usecase

import (
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/application/response"
	"onboarding/src/onboarding/domain/entity"
	"onboarding/src/onboarding/domain/port"
)

// RegisterUserUseCase maneja el registro completo de usuario
type RegisterUserUseCase struct {
	onboardingRepo port.OnboardingRepository
	iamClient      port.IAMClient
}

// NewRegisterUserUseCase crea una nueva instancia del caso de uso
func NewRegisterUserUseCase(
	onboardingRepo port.OnboardingRepository,
	iamClient port.IAMClient,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		onboardingRepo: onboardingRepo,
		iamClient:      iamClient,
	}
}

// Execute ejecuta el proceso completo de registro de usuario
func (uc *RegisterUserUseCase) Execute(req *request.RegisterUserRequest) (*response.RegisterUserResponse, error) {
	// 1. Validar request
	if err := req.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
		return response.NewRegisterUserErrorResponse(err.Error()), nil
	}

	// 2. Crear tenant temporal (se completará después)
	tenant, err := uc.createTenant(req)
	if err != nil {
		log.Printf("Error creating tenant: %v", err)
		return response.NewRegisterUserErrorResponse("Error al crear la organización"), err
	}

	// 3. Obtener rol TENANT_ADMIN
	tenantAdminRole, err := uc.iamClient.GetRoleByType("TENANT_ADMIN")
	if err != nil {
		log.Printf("Error getting tenant admin role: %v", err)
		// Rollback: eliminar tenant
		uc.iamClient.DeleteTenant(tenant.ID)
		return response.NewRegisterUserErrorResponse("Error en la configuración de permisos"), err
	}

	// 4. Crear usuario con rol de administrador
	user, err := uc.createUser(req, tenant.ID, tenantAdminRole.ID)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		// Rollback: eliminar tenant
		uc.iamClient.DeleteTenant(tenant.ID)
		return response.NewRegisterUserErrorResponse("Error al crear el usuario"), err
	}

	// 5. Actualizar owner del tenant
	err = uc.iamClient.UpdateTenantOwner(tenant.ID, user.ID)
	if err != nil {
		log.Printf("Error updating tenant owner: %v", err)
		// Rollback: eliminar usuario y tenant
		uc.iamClient.DeleteUser(user.ID)
		uc.iamClient.DeleteTenant(tenant.ID)
		return response.NewRegisterUserErrorResponse("Error en la configuración final"), err
	}

	// 6. Crear proceso de onboarding
	process, err := uc.createOnboardingProcess(tenant.ID, user.ID)
	if err != nil {
		log.Printf("Error creating onboarding process: %v", err)
		// No hacer rollback aquí, el usuario ya está creado exitosamente
		// Solo logear el error y continuar
	}

	// 7. Preparar respuesta exitosa
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

	processID := ""
	if process != nil {
		processID = process.ID.String()
	}

	log.Printf("User registered successfully: email=%s, tenant=%s, user=%s", req.Email, tenant.ID, user.ID)

	return response.NewRegisterUserResponse(processID, tenant.ID, user.ID, userData, tenantData), nil
}

// createTenant crea un tenant temporal para el usuario
func (uc *RegisterUserUseCase) createTenant(req *request.RegisterUserRequest) (*port.TenantResponse, error) {
	tenantName := fmt.Sprintf("Tienda de %s", req.GetCleanName())
	slug := uc.generateSlugFromName(req.GetCleanName())

	tenantData := &port.CreateTenantRequest{
		Name:        tenantName,
		Slug:        slug,
		Description: fmt.Sprintf("Tienda administrada por %s", req.GetCleanEmail()),
		Type:        "STARTUP",   // Default temporal, se actualiza en paso 4
		OwnerID:     "temp-uuid", // Se actualiza después de crear el usuario
	}

	return uc.iamClient.CreateTenant(tenantData)
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

// createOnboardingProcess crea el proceso de onboarding
func (uc *RegisterUserUseCase) createOnboardingProcess(tenantID, userID string) (*entity.OnboardingProcess, error) {
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

	err = uc.onboardingRepo.SaveProcess(process)
	if err != nil {
		return nil, fmt.Errorf("error saving onboarding process: %w", err)
	}

	return process, nil
}

// generateSlugFromName genera un slug único para el tenant
func (uc *RegisterUserUseCase) generateSlugFromName(name string) string {
	// Convertir a minúsculas y reemplazar espacios
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, ".", "")
	slug = strings.ReplaceAll(slug, ",", "")

	// Agregar sufijo aleatorio para garantizar unicidad
	randomSuffix := uuid.New().String()[:8]
	slug = fmt.Sprintf("%s-%s", slug, randomSuffix)

	return slug
}
