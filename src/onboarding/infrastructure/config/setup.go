package config

import (
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"onboarding/src/onboarding/application/usecase"
	"onboarding/src/onboarding/infrastructure/auth"
	"onboarding/src/onboarding/infrastructure/client"
	"onboarding/src/onboarding/infrastructure/controller"
	"onboarding/src/onboarding/infrastructure/logging"
	"onboarding/src/onboarding/infrastructure/persistence"
)

// SetupOnboardingModule configura e inicializa el módulo de onboarding
func SetupOnboardingModule(router *gin.RouterGroup, db *sql.DB) {
	log.Println("Setting up Onboarding module...")

	// 0. Inicializar logger canónico (ADR-001)
	eventLogger := logging.NewOnboardingLogger("onboarding")

	// 1. Inicializar token provider para service-to-service auth
	jwtSecret := os.Getenv("JWT_SECRET")
	staticToken := os.Getenv("IAM_SUPER_ADMIN_TOKEN")
	tokenProvider := auth.NewServiceTokenProvider(jwtSecret, staticToken)

	// 2. Inicializar clientes externos
	iamClient := client.NewIAMClientWithProvider(tokenProvider)
	pimClient := client.NewPIMClient()

	// Obtener URL del servicio de notificaciones desde variable de entorno
	notificationServiceURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	if notificationServiceURL == "" {
		notificationServiceURL = "http://localhost:8282/api/v1"
	}
	if !strings.Contains(notificationServiceURL, "/api/v1") {
		notificationServiceURL += "/api/v1"
	}

	log.Printf("Using notification service URL: %s", notificationServiceURL)
	notificationClient := client.NewNotificationClient(notificationServiceURL)

	// Obtener URL del servicio de tenant desde variable de entorno
	tenantServiceURL := os.Getenv("TENANT_SERVICE_URL")
	if tenantServiceURL == "" {
		tenantServiceURL = "http://localhost:8120"
	}

	log.Printf("Using tenant service URL: %s", tenantServiceURL)
	tenantClient := client.NewTenantClient(tenantServiceURL)

	// 3. Inicializar repositorios
	onboardingRepo := persistence.NewPostgresOnboardingRepository(db)

	// 4. Inicializar casos de uso con logger canónico inyectado
	startOnboardingUseCase := usecase.NewStartOnboardingUseCase(onboardingRepo, eventLogger)
	registerUserUseCase := usecase.NewRegisterUserUseCase(onboardingRepo, iamClient, notificationClient, eventLogger)
	verifyEmailUseCase := usecase.NewVerifyEmailUseCase(onboardingRepo, iamClient, eventLogger)
	resendVerificationUseCase := usecase.NewResendVerificationUseCase(onboardingRepo, notificationClient, eventLogger)
	setupStoreUseCase := usecase.NewSetupStoreUseCase(onboardingRepo, pimClient, iamClient, eventLogger)
	selectPlanUseCase := usecase.NewSelectPlanUseCase(onboardingRepo, eventLogger)
	completeOnboardingUseCase := usecase.NewCompleteOnboardingUseCase(onboardingRepo, notificationClient, iamClient, tenantClient, eventLogger)
	getProcessStatusUseCase := usecase.NewGetProcessStatusUseCase(onboardingRepo, eventLogger)

	// 5. Inicializar controladores
	onboardingController := controller.NewOnboardingController(
		startOnboardingUseCase,
		registerUserUseCase,
		verifyEmailUseCase,
		resendVerificationUseCase,
		setupStoreUseCase,
		selectPlanUseCase,
		completeOnboardingUseCase,
		getProcessStatusUseCase,
		pimClient,
		onboardingRepo,
	)

	// 6. Configurar rutas
	setupRoutes(router, onboardingController)

	log.Println("Onboarding module setup completed successfully")
}

// setupRoutes configura todas las rutas del módulo onboarding
func setupRoutes(router *gin.RouterGroup, controller *controller.OnboardingController) {
	onboardingGroup := router.Group("/onboarding")
	{
		onboardingGroup.POST("/start", controller.StartOnboarding)
		onboardingGroup.POST("/register-user", controller.RegisterUser)
		onboardingGroup.POST("/verify-email", controller.VerifyEmail)
		onboardingGroup.POST("/resend-verification", controller.ResendVerificationEmail)
		onboardingGroup.POST("/setup-store", controller.SetupStore)
		onboardingGroup.POST("/select-plan", controller.SelectPlan)
		onboardingGroup.POST("/complete", controller.CompleteOnboarding)

		onboardingGroup.GET("/business-types", controller.GetBusinessTypes)
		onboardingGroup.GET("/categories", controller.GetCategories)
		onboardingGroup.GET("/steps", controller.GetStepDefinitions)

		onboardingGroup.GET("/process/:id", controller.GetProcessStatus)
	}

	log.Println("Onboarding routes configured")
}
