package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"onboarding/src/onboarding/application/usecase"
	"pim/src/shared/domain/criteria"
	sharedCriteria "pim/src/shared/infrastructure/criteria"
)

// TenantonboardingHandler maneja las peticiones HTTP para tenant_onboardings
type TenantonboardingHandler struct {
	// TODO: Agregar casos de uso cuando estén implementados
	criteriaHelper *sharedCriteria.EntityCriteriaHelper
}

// NewTenantonboardingHandler crea una nueva instancia del manejador
func NewTenantonboardingHandler() *TenantonboardingHandler {
	return &TenantonboardingHandler{
		criteriaHelper: sharedCriteria.NewEntityCriteriaHelper(),
	}
}

// RegisterRoutes registra las rutas del API
func (h *TenantonboardingHandler) RegisterRoutes(router *gin.RouterGroup) {
	tenant_onboardings := router.Group("/tenant_onboardings")
	{
		tenant_onboardings.POST("", h.Create)
		tenant_onboardings.GET("", h.List)
		tenant_onboardings.GET("/:id", h.GetByID)
		tenant_onboardings.PUT("/:id", h.Update)
		tenant_onboardings.DELETE("/:id", h.Delete)
	}
}

// Create maneja la solicitud para crear un nuevo tenant_onboarding
func (h *TenantonboardingHandler) Create(c *gin.Context) {
	// Obtener el tenantID del header
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el header X-Tenant-ID es obligatorio"})
		return
	}

	// TODO: Implementar binding de request y llamada al use case
	c.JSON(http.StatusNotImplemented, gin.H{"error": "no implementado - falta implementar casos de uso"})
}

// List maneja la solicitud para listar tenant_onboardings
func (h *TenantonboardingHandler) List(c *gin.Context) {
	// Obtener el tenantID del header
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el header X-Tenant-ID es obligatorio"})
		return
	}

	// TODO: Implementar listado con criterios
	c.JSON(http.StatusNotImplemented, gin.H{"error": "no implementado - falta implementar casos de uso"})
}

// GetByID maneja la solicitud para obtener un tenant_onboarding por ID
func (h *TenantonboardingHandler) GetByID(c *gin.Context) {
	// Obtener el tenantID del header
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el header X-Tenant-ID es obligatorio"})
		return
	}

	id := c.Param("id")
	// TODO: Implementar obtención por ID
	c.JSON(http.StatusNotImplemented, gin.H{"error": "no implementado - falta implementar casos de uso"})
}

// Update maneja la solicitud para actualizar un tenant_onboarding
func (h *TenantonboardingHandler) Update(c *gin.Context) {
	// Obtener el tenantID del header
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el header X-Tenant-ID es obligatorio"})
		return
	}

	id := c.Param("id")
	// TODO: Implementar actualización
	c.JSON(http.StatusNotImplemented, gin.H{"error": "no implementado - falta implementar casos de uso"})
}

// Delete maneja la solicitud para eliminar un tenant_onboarding
func (h *TenantonboardingHandler) Delete(c *gin.Context) {
	// Obtener el tenantID del header
	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "el header X-Tenant-ID es obligatorio"})
		return
	}

	id := c.Param("id")
	// TODO: Implementar eliminación
	c.JSON(http.StatusNotImplemented, gin.H{"error": "no implementado - falta implementar casos de uso"})
}
