package config

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"onboarding/src/onboarding/application/usecase"
	"onboarding/src/onboarding/infrastructure/client"
	"onboarding/src/onboarding/infrastructure/controller"
	"onboarding/src/onboarding/infrastructure/persistence"
)

// SetupOnboardingModule configura e inicializa el módulo de onboarding
func SetupOnboardingModule(router *gin.RouterGroup, db *sql.DB) {
	log.Println("Setting up Onboarding module...")

	// 1. Inicializar clientes externos
	iamClient := client.NewIAMClient()
	pimClient := client.NewPIMClient()

	// Obtener URL del servicio de notificaciones desde variable de entorno
	notificationServiceURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	if notificationServiceURL == "" {
		notificationServiceURL = "http://localhost:8282" // fallback para desarrollo local
	}
	notificationServiceURL += "/api/v1" // agregar path de la API

	log.Printf("Using notification service URL: %s", notificationServiceURL)
	notificationClient := client.NewNotificationClient(notificationServiceURL)

	// 2. Inicializar repositorios
	onboardingRepo := persistence.NewPostgresOnboardingRepository(db)

	// 3. Inicializar casos de uso
	startOnboardingUseCase := usecase.NewStartOnboardingUseCase(onboardingRepo)
	registerUserUseCase := usecase.NewRegisterUserUseCase(onboardingRepo, iamClient, notificationClient)
	verifyEmailUseCase := usecase.NewVerifyEmailUseCase(onboardingRepo, iamClient)
	resendVerificationUseCase := usecase.NewResendVerificationUseCase(onboardingRepo, notificationClient)
	setupStoreUseCase := usecase.NewSetupStoreUseCase(onboardingRepo, pimClient, iamClient)
	selectPlanUseCase := usecase.NewSelectPlanUseCase(onboardingRepo)
	completeOnboardingUseCase := usecase.NewCompleteOnboardingUseCase(onboardingRepo, notificationClient, iamClient)
	getProcessStatusUseCase := usecase.NewGetProcessStatusUseCase(onboardingRepo)

	// 4. Inicializar controladores
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

	// 5. Configurar rutas
	setupRoutes(router, onboardingController)

	log.Println("Onboarding module setup completed successfully")
}

// setupRoutes configura todas las rutas del módulo onboarding
func setupRoutes(router *gin.RouterGroup, controller *controller.OnboardingController) {
	// Crear grupo de rutas para onboarding
	onboardingGroup := router.Group("/onboarding")
	{
		// Rutas principales del proceso de onboarding
		onboardingGroup.POST("/start", controller.StartOnboarding)
		onboardingGroup.POST("/register-user", controller.RegisterUser)
		onboardingGroup.POST("/verify-email", controller.VerifyEmail)
		onboardingGroup.POST("/resend-verification", controller.ResendVerificationEmail)
		onboardingGroup.POST("/setup-store", controller.SetupStore)
		onboardingGroup.POST("/select-plan", controller.SelectPlan)
		onboardingGroup.POST("/complete", controller.CompleteOnboarding)

		// Rutas de información y configuración
		onboardingGroup.GET("/business-types", controller.GetBusinessTypes)
		onboardingGroup.GET("/categories", controller.GetCategories)
		onboardingGroup.GET("/steps", controller.GetStepDefinitions)

		// Rutas de estado y seguimiento
		onboardingGroup.GET("/process/:id", controller.GetProcessStatus)
	}

	log.Println("Onboarding routes configured:")
	log.Println("  POST /api/v1/onboarding/start")
	log.Println("  POST /api/v1/onboarding/register-user")
	log.Println("  POST /api/v1/onboarding/verify-email")
	log.Println("  POST /api/v1/onboarding/resend-verification")
	log.Println("  POST /api/v1/onboarding/setup-store")
	log.Println("  POST /api/v1/onboarding/select-plan")
	log.Println("  POST /api/v1/onboarding/complete")
	log.Println("  GET  /api/v1/onboarding/business-types")
	log.Println("  GET  /api/v1/onboarding/categories")
	log.Println("  GET  /api/v1/onboarding/steps")
	log.Println("  GET  /api/v1/onboarding/process/:id")
}
