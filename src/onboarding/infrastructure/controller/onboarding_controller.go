package controller

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/application/usecase"
	"onboarding/src/onboarding/domain/port"
)

// OnboardingController maneja las operaciones HTTP del onboarding
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

// RegisterUser maneja el registro de usuario (POST /api/v1/onboarding/register-user)
func (c *OnboardingController) RegisterUser(ctx *gin.Context) {
	var req request.RegisterUserRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Datos inválidos en la solicitud",
			"error":   err.Error(),
		})
		return
	}

	response, err := c.registerUserUseCase.Execute(&req)
	if err != nil {
		log.Printf("Error in RegisterUser usecase: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error interno del servidor",
			"error":   err.Error(),
		})
		return
	}

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusBadRequest
	}

	ctx.JSON(statusCode, response)
}

// SetupStore maneja la configuración de la tienda (POST /api/v1/onboarding/setup-store)
func (c *OnboardingController) SetupStore(ctx *gin.Context) {
	var req request.SetupStoreRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Datos inválidos en la solicitud",
			"error":   err.Error(),
		})
		return
	}

	response, err := c.setupStoreUseCase.Execute(&req)
	if err != nil {
		log.Printf("Error in SetupStore usecase: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error interno del servidor",
			"error":   err.Error(),
		})
		return
	}

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusBadRequest
	}

	ctx.JSON(statusCode, response)
}

// SelectPlan maneja la selección de plan (POST /api/v1/onboarding/select-plan)
func (c *OnboardingController) SelectPlan(ctx *gin.Context) {
	var req request.SelectPlanRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Datos inválidos en la solicitud",
			"error":   err.Error(),
		})
		return
	}

	response, err := c.selectPlanUseCase.Execute(&req)
	if err != nil {
		log.Printf("Error in SelectPlan usecase: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error interno del servidor",
			"error":   err.Error(),
		})
		return
	}

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusBadRequest
	}

	ctx.JSON(statusCode, response)
}

// GetBusinessTypes obtiene todos los tipos de negocio disponibles desde PIM
func (c *OnboardingController) GetBusinessTypes(ctx *gin.Context) {
	businessTypes, err := c.pimClient.GetBusinessTypes()
	if err != nil {
		log.Printf("Error obtaining business types from PIM: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Error interno del servidor",
		})
		return
	}

	// DEBUG: Verificar qué llega desde el PIM
	log.Printf("DEBUG: Número de business types recibidos: %d", len(businessTypes))
	if len(businessTypes) > 0 {
		log.Printf("DEBUG: Primer business type raw: ID='%s', Name='%s', Icon='%s'",
			businessTypes[0].ID, businessTypes[0].Name, businessTypes[0].Icon)
	}

	// Transformar structure del PIM a estructura de respuesta usando el ID directamente
	transformedBusinessTypes := make([]map[string]interface{}, len(businessTypes))
	for i, bt := range businessTypes {
		// DEBUG: Log cada business type antes de transformar
		log.Printf("DEBUG: BT[%d] - ID='%s', Name='%s', Icon='%s'", i, bt.ID, bt.Name, bt.Icon)

		transformedBusinessTypes[i] = map[string]interface{}{
			"id":          bt.ID, // Usar el campo ID directamente (retail, food-beverage, etc.)
			"name":        bt.Name,
			"description": bt.Description,
			"icon":        bt.Icon,
			"color":       bt.Color,     // Color del PIM
			"is_active":   bt.IsActive,  // Estado del PIM
			"sort_order":  bt.SortOrder, // Orden del PIM
			"created_at":  bt.CreatedAt, // Timestamp del PIM (ahora mapeado como camelCase)
			"updated_at":  bt.UpdatedAt, // Timestamp del PIM (ahora mapeado como camelCase)

			// Campos adicionales vacíos para compatibilidad
			"default_categories": []string{},
			"default_attributes": []string{},
			"default_variants":   []string{},
		}
	}

	log.Printf("Retrieved and transformed %d business types from PIM service", len(transformedBusinessTypes))

	ctx.JSON(http.StatusOK, gin.H{
		"success":        true,
		"business_types": transformedBusinessTypes,
		"message":        "Tipos de negocio obtenidos exitosamente",
	})
}

// GetCategories obtiene las categorías por tipo de negocio (GET /api/v1/onboarding/categories)
func (c *OnboardingController) GetCategories(ctx *gin.Context) {
	businessType := ctx.Query("business_type")
	if businessType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "El parámetro business_type es requerido",
		})
		return
	}

	categories, err := c.pimClient.GetCategoriesByBusinessType(businessType)
	if err != nil {
		log.Printf("Error getting categories for business type %s: %v", businessType, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error al obtener categorías",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Categorías obtenidas exitosamente",
		"business_type": businessType,
		"categories":    categories,
	})
}

// GetProcessStatus obtiene el estado del proceso de onboarding (GET /api/v1/onboarding/process/:id)
func (c *OnboardingController) GetProcessStatus(ctx *gin.Context) {
	processID := ctx.Param("id")
	if processID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "ID de proceso requerido",
		})
		return
	}

	response, err := c.getProcessStatusUseCase.Execute(processID)
	if err != nil {
		log.Printf("Error in GetProcessStatus usecase: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error interno del servidor",
			"error":   err.Error(),
		})
		return
	}

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusNotFound
	}

	ctx.JSON(statusCode, response)
}

// CompleteOnboarding marca el onboarding como completado (POST /api/v1/onboarding/complete)
func (c *OnboardingController) CompleteOnboarding(ctx *gin.Context) {
	var req request.CompleteOnboardingRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Datos inválidos en la solicitud",
			"error":   err.Error(),
		})
		return
	}

	response, err := c.completeOnboardingUseCase.Execute(&req)
	if err != nil {
		log.Printf("Error in CompleteOnboarding usecase: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error interno del servidor",
			"error":   err.Error(),
		})
		return
	}

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusBadRequest
	}

	ctx.JSON(statusCode, response)
}

// GetStepDefinitions obtiene las definiciones de pasos (GET /api/v1/onboarding/steps)
func (c *OnboardingController) GetStepDefinitions(ctx *gin.Context) {
	stepDefinitions, err := c.onboardingRepo.GetStepDefinitions()
	if err != nil {
		log.Printf("Error getting step definitions: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error al obtener definiciones de pasos",
			"error":   err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Definiciones de pasos obtenidas exitosamente",
		"steps":   stepDefinitions,
	})
}

// StartOnboarding inicia un nuevo proceso de onboarding (POST /api/v1/onboarding/start)
func (c *OnboardingController) StartOnboarding(ctx *gin.Context) {
	var req request.StartOnboardingRequest

	// Bind JSON request (campos opcionales)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON (using defaults): %v", err)
		// No retornamos error aquí porque todos los campos son opcionales
	}

	// Capturar información adicional del contexto
	req.IP = ctx.ClientIP()
	req.UserAgent = ctx.GetHeader("User-Agent")

	response, err := c.startOnboardingUseCase.Execute(&req)
	if err != nil {
		log.Printf("Error in StartOnboarding usecase: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error interno del servidor",
			"error":   err.Error(),
		})
		return
	}

	statusCode := http.StatusOK
	if !response.Success {
		statusCode = http.StatusBadRequest
	}

	ctx.JSON(statusCode, response)
}

// VerifyEmail verifica el código de email (POST /api/v1/onboarding/verify-email)
func (c *OnboardingController) VerifyEmail(ctx *gin.Context) {
	var req request.VerifyEmailRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Datos inválidos en la solicitud",
			"error":   err.Error(),
		})
		return
	}

	response, err := c.verifyEmailUseCase.Execute(&req)
	if err != nil {
		log.Printf("Error in VerifyEmail usecase: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error interno del servidor",
			"error":   err.Error(),
		})
		return
	}

	statusCode := http.StatusOK
	if !response.Success {
		if response.Error == "INVALID_VERIFICATION_CODE" {
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusInternalServerError
		}
	}

	ctx.JSON(statusCode, response)
}

// ResendVerificationEmail reenvía el código de verificación de email (POST /api/v1/onboarding/resend-verification)
func (c *OnboardingController) ResendVerificationEmail(ctx *gin.Context) {
	var req request.ResendVerificationRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Datos inválidos en la solicitud",
			"error":   err.Error(),
		})
		return
	}

	response, err := c.resendVerificationUseCase.Execute(&req)
	if err != nil {
		log.Printf("Error in ResendVerification usecase: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Error interno del servidor",
			"error":   err.Error(),
		})
		return
	}

	statusCode := http.StatusOK
	if !response.Success {
		if response.Error == "RATE_LIMITED" {
			statusCode = http.StatusTooManyRequests
		} else {
			statusCode = http.StatusBadRequest
		}
	}

	ctx.JSON(statusCode, response)
}
