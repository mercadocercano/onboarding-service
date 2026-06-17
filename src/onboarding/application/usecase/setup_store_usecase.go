package usecase

import (
	"fmt"

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
	logger         port.OnboardingEventLogger
}

// NewSetupStoreUseCase crea una nueva instancia del caso de uso
func NewSetupStoreUseCase(
	onboardingRepo port.OnboardingRepository,
	pimClient port.PIMClient,
	iamClient port.IAMClient,
	logger ...port.OnboardingEventLogger,
) *SetupStoreUseCase {
	uc := &SetupStoreUseCase{
		onboardingRepo: onboardingRepo,
		pimClient:      pimClient,
		iamClient:      iamClient,
	}
	if len(logger) > 0 && logger[0] != nil {
		uc.logger = logger[0]
	}
	return uc
}

func (uc *SetupStoreUseCase) log(e port.OnboardingEvent) {
	if uc.logger != nil {
		uc.logger.Log(e)
	}
}

// Execute ejecuta la configuración de la tienda
func (uc *SetupStoreUseCase) Execute(req *request.SetupStoreRequest) (*response.SetupStoreResponse, error) {
	// 1. Validar request
	if err := req.Validate(); err != nil {
		return response.NewSetupStoreErrorResponse(err.Error()), nil
	}

	// 2. Obtener el proceso de onboarding
	processID, err := uuid.Parse(req.ProcessID)
	if err != nil {
		return response.NewSetupStoreErrorResponse("ID de proceso inválido"), nil
	}

	process, err := uc.onboardingRepo.GetProcessByID(processID)
	if err != nil {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.store_setup_failed",
			ProcessID: req.ProcessID,
			Reason:    "error getting process: " + err.Error(),
		})
		return response.NewSetupStoreErrorResponse("Proceso de onboarding no encontrado"), err
	}

	// 3. Validar tipos de negocio desde PIM
	businessTypes, err := uc.pimClient.GetBusinessTypes()
	if err != nil {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.store_setup_failed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
			Reason:    "error getting business types: " + err.Error(),
		})
		return response.NewSetupStoreErrorResponse("Error al obtener tipos de negocio"), err
	}

	if !uc.isValidBusinessType(req.BusinessType, businessTypes) {
		return response.NewSetupStoreErrorResponse("Tipo de negocio no válido"), nil
	}

	// 4. Actualizar proceso de onboarding con información básica
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
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.store_setup_failed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
			Step:      4,
			Reason:    "error updating process: " + err.Error(),
		})
		return response.NewSetupStoreErrorResponse("Error al actualizar el proceso"), err
	}

	// 5. Actualizar tenant con información del negocio
	if err := uc.updateTenantInfo(process.TenantID.String(), req); err != nil {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.tenant_info_update_failed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
			Reason:    err.Error(),
		})
		// No fallar el proceso por esto
	}

	// 6. Aplicar configuración PIM
	if err := uc.applyPIMConfiguration(process.TenantID.String(), req); err != nil {
		uc.log(port.OnboardingEvent{
			Event:     "onboarding.store_setup_failed",
			TenantID:  process.TenantID.String(),
			UserID:    process.UserID.String(),
			ProcessID: req.ProcessID,
			Reason:    "error applying PIM configuration: " + err.Error(),
		})
		// No fallar el proceso, el tenant queda configurado básicamente
	}

	// 7. Preparar respuesta exitosa
	businessTypeInfo := uc.getBusinessTypeInfo(req.BusinessType, businessTypes)

	uc.log(port.OnboardingEvent{
		Event:     "onboarding.store_setup_completed",
		TenantID:  process.TenantID.String(),
		UserID:    process.UserID.String(),
		ProcessID: req.ProcessID,
		Step:      5,
	})

	return response.NewSetupStoreResponse(
		process.ID.String(),
		process.TenantID.String(),
		req.GetCleanStoreName(),
		businessTypeInfo,
		req.StoreSize,
		req.GetRecommendedPlan(),
		categories,
		5,
	), nil
}

// isValidBusinessType valida que el business type existe en PIM
func (uc *SetupStoreUseCase) isValidBusinessType(businessType string, businessTypes []*port.BusinessType) bool {
	for _, bt := range businessTypes {
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
		uc.log(port.OnboardingEvent{
			Event:    "onboarding.pim_categories_fetch_failed",
			TenantID: tenantID,
			Reason:   err.Error(),
		})
		suggestedCategories = []*port.Category{}
	}

	// 3. Preparar categorías seleccionadas
	selectedCategories := make([]string, 0, len(suggestedCategories))
	for _, cat := range suggestedCategories {
		if cat != nil && cat.ID != "" {
			selectedCategories = append(selectedCategories, cat.ID)
		}
	}

	if len(selectedCategories) > 5 {
		selectedCategories = selectedCategories[:5]
	}

	// 4. Crear configuración PIM
	pimConfig := &port.QuickstartConfig{
		BusinessType:         req.BusinessType,
		StoreSize:            "small",
		SelectedCategories:   selectedCategories,
		SelectedAttributes:   []string{},
		SelectedVariants:     []string{},
		CreateSampleProducts: false,
	}

	// 5. Aplicar la configuración en PIM
	_, err = uc.pimClient.ApplyQuickstartTemplate(tenantID, pimConfig)
	if err != nil {
		return fmt.Errorf("error applying PIM quickstart template: %w", err)
	}

	// 6. Importar productos del catálogo global para este business type
	importResp, err := uc.pimClient.ImportProductsFromBusinessType(tenantID, req.BusinessType)
	if err != nil {
		uc.log(port.OnboardingEvent{
			Event:    "onboarding.pim_products_import_failed",
			TenantID: tenantID,
			Reason:   err.Error(),
		})
	} else {
		_ = importResp // importResp.Summary.TotalImported disponible si se necesita en el futuro
	}

	return nil
}

// getBusinessTypeInfo obtiene información detallada del business type
func (uc *SetupStoreUseCase) getBusinessTypeInfo(businessType string, businessTypes []*port.BusinessType) *response.BusinessTypeInfo {
	for _, bt := range businessTypes {
		if bt.ID == businessType || bt.Code == businessType {
			return &response.BusinessTypeInfo{
				ID:          bt.ID,
				Name:        bt.Name,
				Description: bt.Description,
				Icon:        bt.Icon,
			}
		}
	}

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
