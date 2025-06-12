package config

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"

	"onboarding/src/onboarding/application/usecase"
	"onboarding/src/onboarding/infrastructure/client"
	"onboarding/src/onboarding/infrastructure/controller"
	"onboarding/src/onboarding/infrastructure/persistence"
)

// SetupOnboardingModule configura e inicializa el módulo de onboarding
func SetupOnboardingModule(router *gin.RouterGroup, db *sql.DB) {
	log.Println("Setting up Onboarding module...")

	// 1. Inicializar repositorios
	onboardingRepo := persistence.NewPostgresOnboardingRepository(db)

	// 2. Inicializar clientes externos
	iamClient := client.NewIAMClient()
	pimClient := client.NewPIMClient()

	// 3. Inicializar casos de uso
	registerUserUseCase := usecase.NewRegisterUserUseCase(onboardingRepo, iamClient)
	setupStoreUseCase := usecase.NewSetupStoreUseCase(onboardingRepo, pimClient, iamClient)

	// 4. Inicializar controladores
	onboardingController := controller.NewOnboardingController(
		registerUserUseCase,
		setupStoreUseCase,
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
		onboardingGroup.POST("/register-user", controller.RegisterUser)
		onboardingGroup.POST("/setup-store", controller.SetupStore)
		onboardingGroup.POST("/complete", controller.CompleteOnboarding)

		// Rutas de información y configuración
		onboardingGroup.GET("/business-types", controller.GetBusinessTypes)
		onboardingGroup.GET("/categories", controller.GetCategories)
		onboardingGroup.GET("/steps", controller.GetStepDefinitions)

		// Rutas de estado y seguimiento
		onboardingGroup.GET("/process/:id", controller.GetProcessStatus)
	}

	log.Println("Onboarding routes configured:")
	log.Println("  POST /api/v1/onboarding/register-user")
	log.Println("  POST /api/v1/onboarding/setup-store")
	log.Println("  POST /api/v1/onboarding/complete")
	log.Println("  GET  /api/v1/onboarding/business-types")
	log.Println("  GET  /api/v1/onboarding/categories")
	log.Println("  GET  /api/v1/onboarding/steps")
	log.Println("  GET  /api/v1/onboarding/process/:id")
}
