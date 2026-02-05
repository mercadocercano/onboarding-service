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

	// 4. OMITIDO: Validación de categorías - ahora es opcional y se maneja en el backoffice
	// Las categorías se configurarán después en el backoffice siguiendo la filosofía
	// "conseguir el registro rápido, luego guiar la configuración completa"

	// 5. Actualizar proceso de onboarding con información básica
	// Si no hay categorías, pasar array vacío
	categories := req.SelectedCategories
	if categories == nil {
		categories = []string{}
	}

	process.SetStoreConfiguration(
		req.GetCleanStoreName(),
		req.BusinessType,
		req.StoreSize,
		categories,
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

	// 7. Aplicar configuración PIM - vincular business_type con tenant
	err = uc.applyPIMConfiguration(process.TenantID.String(), req)
	if err != nil {
		log.Printf("Error applying PIM configuration: %v", err)
		// No fallar el proceso por esto, pero logearlo como error
		// El tenant queda configurado básicamente, pero sin el setup PIM completo
	}

	// 8. Preparar respuesta exitosa
	businessTypeInfo := uc.getBusinessTypeInfo(req.BusinessType, businessTypes)

	log.Printf("Store configured successfully: tenant=%s, store=%s, business_type=%s, categories_deferred=true",
		process.TenantID.String(), req.StoreName, req.BusinessType)

	return response.NewSetupStoreResponse(
		process.ID.String(),
		process.TenantID.String(),
		req.GetCleanStoreName(),
		businessTypeInfo,
		req.StoreSize,
		req.GetRecommendedPlan(),
		categories, // Array vacío o categorías si fueron proporcionadas
		5,          // Siguiente paso: finalización
	), nil
}

// isValidBusinessType valida que el business type existe en PIM
func (uc *SetupStoreUseCase) isValidBusinessType(businessType string, businessTypes []*port.BusinessType) bool {
	for _, bt := range businessTypes {
		// Comparar tanto con ID (UUID) como con Code (ej: "retail", "almacen")
		if bt.ID == businessType || bt.Code == businessType {
			return true
		}
	}
	return false
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

// applyPIMConfiguration aplica la configuración PIM para vincular business_type con tenant
func (uc *SetupStoreUseCase) applyPIMConfiguration(tenantID string, req *request.SetupStoreRequest) error {
	// 1. Validar que el business_type existe en PIM
	businessType, err := uc.pimClient.GetBusinessType(req.BusinessType)
	if err != nil {
		return fmt.Errorf("error validating business type: %w", err)
	}
	if businessType == nil {
		return fmt.Errorf("business type '%s' not found", req.BusinessType)
	}

	// 2. Obtener categorías sugeridas para este business type
	suggestedCategories, err := uc.pimClient.GetCategoriesByBusinessType(req.BusinessType)
	if err != nil {
		log.Printf("Warning: Could not get suggested categories for business type %s: %v", req.BusinessType, err)
		// Continuar con categorías vacías si no se pueden obtener
		suggestedCategories = []*port.Category{}
	}

	// 3. Preparar categorías seleccionadas (por ahora usar las sugeridas)
	selectedCategories := make([]string, 0, len(suggestedCategories))
	for _, cat := range suggestedCategories {
		if cat != nil && cat.ID != "" {
			selectedCategories = append(selectedCategories, cat.ID)
		}
	}

	// Limitar a máximo 5 categorías para el setup inicial
	if len(selectedCategories) > 5 {
		selectedCategories = selectedCategories[:5]
	}

	// 4. Crear configuración PIM
	pimConfig := &port.QuickstartConfig{
		BusinessType:       req.BusinessType,
		StoreSize:          "small", // Por defecto pequeña para onboarding
		SelectedCategories: selectedCategories,
		// Los atributos y variantes se configurarán después en el backoffice
		SelectedAttributes:   []string{},
		SelectedVariants:     []string{},
		CreateSampleProducts: false, // No crear productos de ejemplo en onboarding
	}

	// 5. Aplicar la configuración en PIM (vincular business_type con tenant)
	response, err := uc.pimClient.ApplyQuickstartTemplate(tenantID, pimConfig)
	if err != nil {
		return fmt.Errorf("error applying PIM quickstart template: %w", err)
	}

	log.Printf("PIM configuration applied successfully for tenant %s: %+v", tenantID, response)

	// 6. Log de estadísticas de configuración
	log.Printf("Business type '%s' linked to tenant %s with %d categories",
		req.BusinessType, tenantID, len(selectedCategories))

	return nil
}

// getBusinessTypeInfo obtiene información detallada del business type
func (uc *SetupStoreUseCase) getBusinessTypeInfo(businessType string, businessTypes []*port.BusinessType) *response.BusinessTypeInfo {
	for _, bt := range businessTypes {
		// Buscar por ID (UUID) o por Code (ej: "retail", "almacen")
		if bt.ID == businessType || bt.Code == businessType {
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
