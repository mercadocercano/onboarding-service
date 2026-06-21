package config

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hornosg/go-shared/infrastructure/postgres"
	"github.com/mercadocercano/eventbus"

	"onboarding/src/onboarding/application/usecase"
	"onboarding/src/onboarding/domain/port"
	"onboarding/src/onboarding/infrastructure/auth"
	"onboarding/src/onboarding/infrastructure/client"
	"onboarding/src/onboarding/infrastructure/controller"
	onboardingevent "onboarding/src/onboarding/infrastructure/event"
	"onboarding/src/onboarding/infrastructure/logging"
	"onboarding/src/onboarding/infrastructure/persistence"
)

// SetupOnboardingModule configura e inicializa el módulo de onboarding
func SetupOnboardingModule(router *gin.RouterGroup, db *sql.DB) {
	log.Println("Setting up Onboarding module...")

	// 0. Inicializar logger canónico (ADR-001)
	eventLogger := logging.NewOnboardingLogger("onboarding")

	// 1. Inicializar token provider para service-to-service auth.
	//    Se firma tenant_id (system tenant, mismo valor que viaja en X-Tenant-ID) y
	//    namespace para que el token S2S no dependa del bypass de tenant.
	jwtSecret := os.Getenv("JWT_SECRET")
	staticToken := os.Getenv("IAM_SUPER_ADMIN_TOKEN")
	systemTenantID := os.Getenv("SYSTEM_TENANT_ID")
	if systemTenantID == "" {
		systemTenantID = "123e4567-e89b-12d3-a456-426614174003"
	}
	serviceNamespace := os.Getenv("SERVICE_NAMESPACE")
	if serviceNamespace == "" {
		serviceNamespace = "mc"
	}
	tokenProvider := auth.NewServiceTokenProviderWithIdentity(jwtSecret, staticToken, systemTenantID, serviceNamespace)

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
	// Welcome email event-driven (Plan F1): si EVENTBUS_ENABLED, el welcome se publica al
	// EventBus y lo consume notification-service; si no, sigue por HTTP sincrónico.
	if publisher := setupEventPublisher(serviceNamespace); publisher != nil {
		completeOnboardingUseCase = completeOnboardingUseCase.WithEventPublisher(publisher)
		log.Println("Onboarding welcome email: event-driven (EventBus) ENABLED")
	}
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

// setupEventPublisher construye el publisher del EventBus, gateado por EVENTBUS_ENABLED.
// Best-effort: si la conexión a la DB del EventBus falla, devuelve nil → el welcome cae al
// path HTTP legacy (no rompe el onboarding).
func setupEventPublisher(namespace string) port.EventPublisher {
	if os.Getenv("EVENTBUS_ENABLED") != "true" {
		return nil
	}

	ebDB, err := postgres.Connect(postgres.Config{
		Host:     getEnv("EVENTBUS_DB_HOST", "localhost"),
		Port:     getEnv("EVENTBUS_DB_PORT", "5432"),
		User:     getEnv("EVENTBUS_DB_USER", "postgres"),
		Password: getEnv("EVENTBUS_DB_PASSWORD", "postgres"),
		DBName:   getEnv("EVENTBUS_DB_NAME", "eventbus"),
		SSLMode:  getEnv("EVENTBUS_DB_SSL_MODE", "disable"),
	})
	if err != nil {
		log.Printf("Warning: could not connect to EventBus DB, welcome email falls back to HTTP: %v", err)
		return nil
	}

	postgres.StartPoolMonitor(context.Background(), ebDB, postgres.MonitorOptions{
		Service: "onboarding-service",
		DBName:  getEnv("EVENTBUS_DB_NAME", "eventbus"),
	})

	infraLogger := eventbus.NewLogger(eventbus.LevelInfo)
	store := eventbus.NewSQLEventStore(ebDB, infraLogger)
	publishUC := eventbus.NewPublishEventUseCase(store, infraLogger)
	return onboardingevent.NewEventBusPublisher(publishUC, namespace)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
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
