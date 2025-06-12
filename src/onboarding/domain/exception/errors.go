package exception

import "errors"

// Errores de validación
var (
	ErrTenantonboardingInvalidName = errors.New("nombre de tenant_onboarding inválido")
	ErrTenantonboardingNameRequired = errors.New("nombre de tenant_onboarding es requerido")
)

// Errores de negocio
var (
	ErrTenantonboardingNotFound = errors.New("tenant_onboarding no encontrado")
	ErrTenantonboardingAlreadyExists = errors.New("tenant_onboarding ya existe")
)

// Errores de persistencia
var (
	ErrTenantonboardingCreateFailed = errors.New("error al crear tenant_onboarding")
	ErrTenantonboardingUpdateFailed = errors.New("error al actualizar tenant_onboarding")
	ErrTenantonboardingDeleteFailed = errors.New("error al eliminar tenant_onboarding")
)
