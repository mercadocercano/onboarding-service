package main

// FORCE REBUILD: 2025-06-20 19:30 - Debug business types mapping issue

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/hornosg/go-shared/infrastructure/env"
	tenantmw "github.com/hornosg/go-shared/infrastructure/middleware"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	onboardingConfig "onboarding/src/onboarding/infrastructure/config"
	"onboarding/src/onboarding/infrastructure/migration"
	"onboarding/src/shared/middleware"
)

func main() {
	// Configuración de la base de datos
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Ejecutar migraciones automáticamente
	if err := runMigrations(db); err != nil {
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
	user := env.Get("DB_USER", "postgres")
	password := env.Get("DB_PASSWORD", "postgres")
	dbname := env.Get("DB_NAME", "onboarding_db")
	sslmode := env.Get("DB_SSLMODE", "disable")

	dsn := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=" + sslmode

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Verificar la conexión
	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to database")
	return db, nil
}

// runMigrations ejecuta las migraciones de base de datos
func runMigrations(db *sql.DB) error {
	migrationsPath := migration.GetMigrationsPath()
	log.Printf("Using migrations path: %s", migrationsPath)

	migrator := migration.NewMigrator(db, migrationsPath)
	return migrator.RunMigrations()
}
