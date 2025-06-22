package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/application/usecase"
	"onboarding/src/onboarding/domain/port"
	"onboarding/src/shared/middleware"
)

// OnboardingController maneja las operaciones HTTP del onboarding - versión refactorizada
type OnboardingController struct {
	startOnboardingUseCase    *usecase.StartOnboardingUseCase
	registerUserUseCase       *usecase.RegisterUserUseCase
	verifyEmailUseCase        *usecase.VerifyEmailUseCase
	resendVerificationUseCase *usecase.ResendVerificationUseCase
	setupStoreUseCase         *usecase.SetupStoreUseCase
	selectPlanUseCase         *usecase.SelectPlanUseCase
	completeOnboardingUseCase *usecase.CompleteOnboardingUseCase
	getProcessStatusUseCase   *usecase.GetProcessStatusUseCase
	pimClient                 port.PIMClient
	onboardingRepo            port.OnboardingRepository
}

// NewOnboardingController crea una nueva instancia del controlador
func NewOnboardingController(
	startOnboardingUseCase *usecase.StartOnboardingUseCase,
	registerUserUseCase *usecase.RegisterUserUseCase,
	verifyEmailUseCase *usecase.VerifyEmailUseCase,
	resendVerificationUseCase *usecase.ResendVerificationUseCase,
	setupStoreUseCase *usecase.SetupStoreUseCase,
	selectPlanUseCase *usecase.SelectPlanUseCase,
	completeOnboardingUseCase *usecase.CompleteOnboardingUseCase,
	getProcessStatusUseCase *usecase.GetProcessStatusUseCase,
	pimClient port.PIMClient,
	onboardingRepo port.OnboardingRepository,
) *OnboardingController {
	return &OnboardingController{
		startOnboardingUseCase:    startOnboardingUseCase,
		registerUserUseCase:       registerUserUseCase,
		verifyEmailUseCase:        verifyEmailUseCase,
		resendVerificationUseCase: resendVerificationUseCase,
		setupStoreUseCase:         setupStoreUseCase,
		selectPlanUseCase:         selectPlanUseCase,
		completeOnboardingUseCase: completeOnboardingUseCase,
		getProcessStatusUseCase:   getProcessStatusUseCase,
		pimClient:                 pimClient,
		onboardingRepo:            onboardingRepo,
	}
}

// RegisterUser maneja el registro de usuario - versión refactorizada
func (c *OnboardingController) RegisterUser(ctx *gin.Context) {
	var req request.RegisterUserRequest

	if middleware.ShouldBindJSONWithError(ctx, &req) != nil {
		return // Error ya manejado por el middleware
	}

	response, err := c.registerUserUseCase.Execute(&req)
	if err != nil {
		middleware.AbortWithError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// SetupStore maneja la configuración de la tienda - versión refactorizada
func (c *OnboardingController) SetupStore(ctx *gin.Context) {
	var req request.SetupStoreRequest

	if middleware.ShouldBindJSONWithError(ctx, &req) != nil {
		return
	}

	response, err := c.setupStoreUseCase.Execute(&req)
	if err != nil {
		middleware.AbortWithError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// SelectPlan maneja la selección de plan - versión refactorizada
func (c *OnboardingController) SelectPlan(ctx *gin.Context) {
	var req request.SelectPlanRequest

	if middleware.ShouldBindJSONWithError(ctx, &req) != nil {
		return
	}

	response, err := c.selectPlanUseCase.Execute(&req)
	if err != nil {
		middleware.AbortWithError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// GetBusinessTypes obtiene tipos de negocio - versión refactorizada
func (c *OnboardingController) GetBusinessTypes(ctx *gin.Context) {
	businessTypes, err := c.pimClient.GetBusinessTypes()
	if err != nil {
		log.Printf("Error obtaining business types from PIM: %v", err)
		middleware.AbortWithError(ctx, err)
		return
	}

	// Transformar para response
	response := make([]map[string]interface{}, len(businessTypes))
	for i, bt := range businessTypes {
		response[i] = map[string]interface{}{
			"id":                 bt.ID,
			"name":               bt.Name,
			"description":        bt.Description,
			"icon":               bt.Icon,
			"color":              bt.Color,
			"is_active":          bt.IsActive,
			"sort_order":         bt.SortOrder,
			"created_at":         bt.CreatedAt,
			"updated_at":         bt.UpdatedAt,
			"default_categories": []string{},
			"default_attributes": []string{},
			"default_variants":   []string{},
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":        true,
		"business_types": response,
		"message":        "Tipos de negocio obtenidos exitosamente",
	})
}

// GetCategories obtiene categorías por tipo de negocio - versión refactorizada
func (c *OnboardingController) GetCategories(ctx *gin.Context) {
	businessType := ctx.Query("business_type")
	if businessType == "" {
		middleware.AbortWithBusinessError(ctx, middleware.BusinessError{
			Code:       "MISSING_BUSINESS_TYPE",
			Message:    "El parámetro business_type es requerido",
			HTTPStatus: http.StatusBadRequest,
		})
		return
	}

	categories, err := c.pimClient.GetCategoriesByBusinessType(businessType)
	if err != nil {
		middleware.AbortWithError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Categorías obtenidas exitosamente",
		"business_type": businessType,
		"categories":    categories,
	})
}

// GetProcessStatus obtiene el estado del proceso - versión refactorizada
func (c *OnboardingController) GetProcessStatus(ctx *gin.Context) {
	processID := ctx.Param("id")
	if processID == "" {
		middleware.AbortWithBusinessError(ctx, middleware.BusinessError{
			Code:       "MISSING_PROCESS_ID",
			Message:    "ID de proceso requerido",
			HTTPStatus: http.StatusBadRequest,
		})
		return
	}

	response, err := c.getProcessStatusUseCase.Execute(processID)
	if err != nil {
		middleware.AbortWithError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// CompleteOnboarding completa el proceso - versión refactorizada
func (c *OnboardingController) CompleteOnboarding(ctx *gin.Context) {
	var req request.CompleteOnboardingRequest

	if middleware.ShouldBindJSONWithError(ctx, &req) != nil {
		return
	}

	response, err := c.completeOnboardingUseCase.Execute(&req)
	if err != nil {
		middleware.AbortWithError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// GetStepDefinitions obtiene definiciones de pasos - versión refactorizada
func (c *OnboardingController) GetStepDefinitions(ctx *gin.Context) {
	stepDefinitions, err := c.onboardingRepo.GetStepDefinitions()
	if err != nil {
		middleware.AbortWithError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Definiciones de pasos obtenidas exitosamente",
		"steps":   stepDefinitions,
	})
}

// StartOnboarding inicia proceso - versión refactorizada
func (c *OnboardingController) StartOnboarding(ctx *gin.Context) {
	var req request.StartOnboardingRequest

	// Bind JSON (campos opcionales, no abortamos en error)
	ctx.ShouldBindJSON(&req)

	req.IP = ctx.ClientIP()
	req.UserAgent = ctx.GetHeader("User-Agent")

	response, err := c.startOnboardingUseCase.Execute(&req)
	if err != nil {
		middleware.AbortWithError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// VerifyEmail verifica código de email - versión refactorizada
func (c *OnboardingController) VerifyEmail(ctx *gin.Context) {
	var req request.VerifyEmailRequest

	if middleware.ShouldBindJSONWithError(ctx, &req) != nil {
		return
	}

	response, err := c.verifyEmailUseCase.Execute(&req)
	if err != nil {
		middleware.AbortWithError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// ResendVerificationEmail reenvía código - versión refactorizada
func (c *OnboardingController) ResendVerificationEmail(ctx *gin.Context) {
	var req request.ResendVerificationRequest

	if middleware.ShouldBindJSONWithError(ctx, &req) != nil {
		return
	}

	response, err := c.resendVerificationUseCase.Execute(&req)
	if err != nil {
		middleware.AbortWithError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, response)
}
