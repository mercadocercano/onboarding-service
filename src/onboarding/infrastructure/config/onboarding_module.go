package config

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"onboarding/src/onboarding/infrastructure/controller"
	"onboarding/src/onboarding/infrastructure/persistence/repository"
)

// SetupOnboardingModule configura e inicializa el módulo de onboarding
func SetupOnboardingModule(apiGroup *gin.RouterGroup, db *sql.DB) {
	// Crear repositorios PostgreSQL
	tenantOnboardingRepo := repository.NewPostgresTenantOnboardingRepository(db)
	businessTypeRepo := repository.NewPostgresBusinessTypeRepository(db)
	onboardingStepRepo := repository.NewPostgresOnboardingStepRepository(db)

	// Configurar controlador HTTP
	onboardingHandler := controller.NewOnboardingHandler(
		tenantOnboardingRepo,
		businessTypeRepo,
		onboardingStepRepo,
	)

	// Registrar rutas HTTP
	onboardingHandler.RegisterRoutes(apiGroup)
}
