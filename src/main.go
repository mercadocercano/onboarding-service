package main

// FORCE REBUILD: 2025-06-20 19:30 - Debug business types mapping issue

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hornosg/go-shared/infrastructure/env"
	tenantmw "github.com/hornosg/go-shared/infrastructure/middleware"
	"github.com/hornosg/go-shared/infrastructure/postgres"
	sharedmigrate "github.com/hornosg/go-shared/migrate"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	onboardingroot "onboarding"
	onboardingConfig "onboarding/src/onboarding/infrastructure/config"
	"onboarding/src/shared/middleware"
)

func main() {
	// Configuración de la base de datos
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Fail-fast anti-superuser (PLAT-E28 T7, C1 CRÍTICO @dev-security, patrón E27): el runtime
	// NUNCA debe correr como superuser/BYPASSRLS. FORCE ROW LEVEL SECURITY no aplica a superusers →
	// con un rol privilegiado la RLS de `onboarding_processes`/`verification_codes` (identidad en
	// el alta + email PII + códigos de verificación) queda "activa" pero nunca ejercida, sirviendo
	// datos cross-tenant sin error visible. Abortar el arranque es la única forma fail-closed ante
	// misconfig de credenciales. Escape hatch explícito ALLOW_SUPERUSER_DB=true solo para tareas
	// admin locales (nunca en prod).
	if err := assertNoRLSBypass(db); err != nil {
		log.Fatalf("%v", err)
	}

	// Migraciones versionadas in-app (ADR-001) — golang-migrate, fail-fast.
	// Reemplaza el migrador casero (src/onboarding/infrastructure/migration).
	dbName := env.Get("DB_NAME", "onboarding_db")
	if err := sharedmigrate.RunMigrations(db, onboardingroot.MigrationsFS, dbName); err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	// Configuración del router
	router := gin.New() // Usar gin.New() para evitar middlewares duplicados

	// Agregar middlewares básicos necesarios
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(tenantmw.TenantValidation(tenantmw.TenantValidationConfig{
		JWTSecret: os.Getenv("JWT_SECRET"),
		ExcludedRoutes: []string{
			"/health",
			"/api/v1/health",
			"/metrics",
			// Rutas públicas del wizard (usuario aún no autenticado)
			"/api/v1/onboarding/*",
		},
		RejectMissingTenant: true, // cierre de bypass de tenant (rollout verificado 2026-06-19)
	}))

	// Agregar middleware de manejo de errores
	router.Use(middleware.ErrorHandlerMiddleware())

	// Configurar Prometheus metrics si está habilitado
	prometheusEnabled := os.Getenv("PROMETHEUS_ENABLED")
	log.Printf("PROMETHEUS_ENABLED value: '%s'", prometheusEnabled)

	if prometheusEnabled == "true" {
		log.Println("Registering /metrics endpoint")
		// Endpoint de métricas usando la librería oficial de Prometheus
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))
		log.Println("/metrics endpoint registered successfully")
	} else {
		log.Println("Prometheus metrics disabled")
	}

	// Configuración de CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "up",
			"service": "onboarding",
		})
	})

	// API v1 group
	apiV1 := router.Group("/api/v1")

	// Configurar módulo onboarding
	onboardingConfig.SetupOnboardingModule(apiV1, db)

	// Iniciar el servidor
	port := env.Get("PORT", "8110")
	log.Printf("Starting Onboarding server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func setupDatabase() (*sql.DB, error) {
	// Configuración de la base de datos desde variables de entorno
	host := env.Get("DB_HOST", "localhost")
	port := env.Get("DB_PORT", "5432")
	// Sin default inseguro (PLAT-E28 T7, C1 @dev-security): el viejo default "postgres" es
	// superuser → BYPASSRLS. Si el env falta o queda mal, el servicio arrancaba SIN error con la
	// RLS de `onboarding_processes`/`verification_codes` "activa" pero nunca ejercida, sirviendo
	// datos del alta cross-tenant. DB_USER es obligatorio y debe ser un rol NOBYPASSRLS
	// (onboarding_app). El chequeo de rol vivo se hace en assertNoRLSBypass.
	user := env.Get("DB_USER", "")
	if user == "" {
		return nil, fmt.Errorf("DB_USER is required and must be a NOBYPASSRLS application role " +
			"(e.g. onboarding_app), never postgres — RULE-09/RULE-10, PLAT-E28 C1")
	}
	password := env.Get("DB_PASSWORD", "")
	dbname := env.Get("DB_NAME", "onboarding_db")
	sslmode := env.Get("DB_SSLMODE", "disable")

	db, err := postgres.Connect(postgres.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DBName:   dbname,
		SSLMode:  sslmode,
	})
	if err != nil {
		return nil, err
	}

	postgres.StartPoolMonitor(context.Background(), db, postgres.MonitorOptions{
		Service: "onboarding-service",
		DBName:  dbname,
	})

	log.Println("Successfully connected to database")
	return db, nil
}

// assertNoRLSBypass aborta el arranque si el rol de base de datos con el que conectamos es
// superuser o tiene el atributo BYPASSRLS (PLAT-E28 T7, C1 CRÍTICO @dev-security). Con un rol
// así, FORCE ROW LEVEL SECURITY no se aplica y la RLS de `onboarding_processes` y
// `verification_codes` (identidad en el momento del alta, email PII y códigos de verificación)
// queda inerte: el servicio serviría datos cross-tenant sin ningún error visible. Convierte ese
// fail-OPEN silencioso en un fail-CLOSED ruidoso. ALLOW_SUPERUSER_DB=true es un escape hatch
// explícito para tareas admin locales — jamás debe usarse en producción.
func assertNoRLSBypass(db *sql.DB) error {
	if env.Get("ALLOW_SUPERUSER_DB", "false") == "true" {
		log.Println("⚠️  ALLOW_SUPERUSER_DB=true — se omite el chequeo NOBYPASSRLS (solo admin/local, NUNCA prod)")
		return nil
	}

	var privileged bool
	if err := db.QueryRow(
		`SELECT rolsuper OR rolbypassrls FROM pg_roles WHERE rolname = current_user`,
	).Scan(&privileged); err != nil {
		return fmt.Errorf("no se pudo verificar los privilegios del rol de DB (current_user): %w", err)
	}
	if privileged {
		return fmt.Errorf("negativa a arrancar: el rol de DB actual es SUPERUSER o BYPASSRLS y "+
			"eludiría la row-level security de `onboarding_processes`/`verification_codes` (email "+
			"PII + códigos de verificación — RULE-09/RULE-10, PLAT-E28 C1). Usá un rol NOBYPASSRLS "+
			"como onboarding_app, o exportá ALLOW_SUPERUSER_DB=true solo para tareas admin locales")
	}

	log.Println("RLS guard OK: el rol de DB es NOBYPASSRLS (onboarding_processes y verification_codes protegidas por FORCE ROW LEVEL SECURITY)")
	return nil
}

