package usecase

import (
	"fmt"
	"log"

	"github.com/google/uuid"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/application/response"
	"onboarding/src/onboarding/domain/port"
)

// SetupStoreUseCase maneja la configuración de la tienda
type SetupStoreUseCase struct {
	onboardingRepo port.OnboardingRepository
	pimClient      port.PIMClient
	iamClient      port.IAMClient
}

// NewSetupStoreUseCase crea una nueva instancia del caso de uso
func NewSetupStoreUseCase(
	onboardingRepo port.OnboardingRepository,
	pimClient port.PIMClient,
	iamClient port.IAMClient,
) *SetupStoreUseCase {
	return &SetupStoreUseCase{
		onboardingRepo: onboardingRepo,
		pimClient:      pimClient,
		iamClient:      iamClient,
	}
}

// Execute ejecuta la configuración de la tienda
func (uc *SetupStoreUseCase) Execute(req *request.SetupStoreRequest) (*response.SetupStoreResponse, error) {
	// 1. Validar request
	if err := req.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
		return response.NewSetupStoreErrorResponse(err.Error()), nil
	}

	// 2. Obtener el proceso de onboarding
	processID, err := uuid.Parse(req.ProcessID)
	if err != nil {
		log.Printf("Invalid process ID: %v", err)
		return response.NewSetupStoreErrorResponse("ID de proceso inválido"), nil
	}

	process, err := uc.onboardingRepo.GetProcessByID(processID)
	if err != nil {
		log.Printf("Error getting onboarding process: %v", err)
		return response.NewSetupStoreErrorResponse("Proceso de onboarding no encontrado"), err
	}

	// 3. Validar tipos de negocio desde PIM
	businessTypes, err := uc.pimClient.GetBusinessTypes()
	if err != nil {
		log.Printf("Error getting business types from PIM: %v", err)
		return response.NewSetupStoreErrorResponse("Error al obtener tipos de negocio"), err
	}

	if !uc.isValidBusinessType(req.BusinessType, businessTypes) {
		return response.NewSetupStoreErrorResponse("Tipo de negocio no válido"), nil
	}

	// 4. Validar categorías desde PIM
	categories, err := uc.pimClient.GetCategoriesByBusinessType(req.BusinessType)
	if err != nil {
		log.Printf("Error getting categories from PIM: %v", err)
		return response.NewSetupStoreErrorResponse("Error al obtener categorías"), err
	}

	if !uc.areValidCategories(req.SelectedCategories, categories) {
		return response.NewSetupStoreErrorResponse("Una o más categorías seleccionadas no son válidas"), nil
	}

	// 5. Actualizar proceso de onboarding
	process.SetStoreConfiguration(
		req.GetCleanStoreName(),
		req.BusinessType,
		req.StoreSize,
		req.SelectedCategories,
	)

	// Marcar paso 4 como completado
	process.CompleteStep(4)

	// Avanzar al paso 5 (finalización)
	process.AdvanceToStep(5)

	err = uc.onboardingRepo.UpdateProcess(process)
	if err != nil {
		log.Printf("Error updating onboarding process: %v", err)
		return response.NewSetupStoreErrorResponse("Error al actualizar el proceso"), err
	}

	// 6. Actualizar tenant con información del negocio
	err = uc.updateTenantInfo(process.TenantID.String(), req)
	if err != nil {
		log.Printf("Error updating tenant info: %v", err)
		// No fallar el proceso por esto, solo logear
	}

	// 7. Preparar configuración PIM para aplicar después (en backoffice)
	err = uc.preparePIMConfiguration(process.TenantID.String(), req)
	if err != nil {
		log.Printf("Error preparing PIM configuration: %v", err)
		// No fallar el proceso por esto, solo logear
	}

	// 8. Preparar respuesta exitosa
	businessTypeInfo := uc.getBusinessTypeInfo(req.BusinessType, businessTypes)

	log.Printf("Store configured successfully: tenant=%s, store=%s, business_type=%s",
		process.TenantID.String(), req.StoreName, req.BusinessType)

	return response.NewSetupStoreResponse(
		process.ID.String(),
		process.TenantID.String(),
		req.GetCleanStoreName(),
		businessTypeInfo,
		req.StoreSize,
		req.GetRecommendedPlan(),
		req.SelectedCategories,
		5, // Siguiente paso: finalización
	), nil
}

// isValidBusinessType valida que el business type existe en PIM
func (uc *SetupStoreUseCase) isValidBusinessType(businessType string, businessTypes []*port.BusinessType) bool {
	for _, bt := range businessTypes {
		if bt.ID == businessType {
			return true
		}
	}
	return false
}

// areValidCategories valida que todas las categorías existen para el business type
func (uc *SetupStoreUseCase) areValidCategories(selectedCategories []string, availableCategories []*port.Category) bool {
	categoryMap := make(map[string]bool)
	for _, cat := range availableCategories {
		categoryMap[cat.ID] = true
		categoryMap[cat.Name] = true // Permitir búsqueda por nombre también
	}

	for _, selected := range selectedCategories {
		if !categoryMap[selected] {
			return false
		}
	}
	return true
}

// updateTenantInfo actualiza la información del tenant con datos del negocio
func (uc *SetupStoreUseCase) updateTenantInfo(tenantID string, req *request.SetupStoreRequest) error {
	tenantType := uc.getTenantTypeByStoreSize(req.StoreSize)

	updateReq := &port.UpdateTenantRequest{
		Name: req.GetCleanStoreName(),
		Type: tenantType,
		Description: fmt.Sprintf("%s - %s",
			req.GetBusinessTypeDisplayName(),
			req.GetCleanStoreName()),
	}

	_, err := uc.iamClient.UpdateTenant(tenantID, updateReq)
	return err
}

// preparePIMConfiguration prepara la configuración PIM para aplicar después
func (uc *SetupStoreUseCase) preparePIMConfiguration(tenantID string, req *request.SetupStoreRequest) error {
	// Por ahora solo validamos que el template existe
	// La aplicación real se hará en el backoffice
	_, err := uc.pimClient.GetQuickstartTemplate(req.BusinessType)
	if err != nil {
		return fmt.Errorf("template de quickstart no encontrado para %s: %w", req.BusinessType, err)
	}

	log.Printf("PIM configuration prepared for tenant %s with business type %s", tenantID, req.BusinessType)
	return nil
}

// getBusinessTypeInfo obtiene información detallada del business type
func (uc *SetupStoreUseCase) getBusinessTypeInfo(businessType string, businessTypes []*port.BusinessType) *response.BusinessTypeInfo {
	for _, bt := range businessTypes {
		if bt.ID == businessType {
			return &response.BusinessTypeInfo{
				ID:          bt.ID,
				Name:        bt.Name,
				Description: bt.Description,
				Icon:        bt.Icon,
			}
		}
	}

	// Fallback si no se encuentra
	return &response.BusinessTypeInfo{
		ID:   businessType,
		Name: businessType,
	}
}

// getTenantTypeByStoreSize mapea el tamaño de tienda al tipo de tenant
func (uc *SetupStoreUseCase) getTenantTypeByStoreSize(storeSize string) string {
	switch storeSize {
	case "micro":
		return "STARTUP"
	case "pyme":
		return "BUSINESS"
	case "multiple":
		return "ENTERPRISE"
	default:
		return "STARTUP"
	}
}
