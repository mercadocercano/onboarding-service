# onboarding-service

Guía breve para asistentes de código en el repositorio **onboarding-service** (Mercado Cercano).

## Identidad

Microservicio de **onboarding guiado** para nuevos socios: flujo por pasos (inicio, registro, verificación, tienda, plan, cierre) con persistencia en PostgreSQL e integración con IAM, PIM, notificaciones y tenant.

- **Puerto por defecto**: `8110`
- **Stack**: Go, Gin, PostgreSQL
- **Prefijo HTTP**: `/api/v1` + grupo `onboarding` (vía Kong: `/onboarding/api/v1` según OpenAPI del repo)

## Comandos esenciales

| Acción | Comando |
|--------|---------|
| Ejecutar local | `go run src/main.go` |
| Tests | `go test ./...` |

## Contexto on-demand (cargar según necesidad)

| Archivo | Cuándo cargar |
|---------|---------------|
| `onboarding-service-management/api-endpoints.md` | Al trabajar con endpoints |
| `onboarding-service-management/architecture.md` | Al crear módulos, entidades, puertos |
| `onboarding-service-management/config.md` | Variables de entorno, Docker, migraciones |

## Reglas compartidas (cargar según contexto)

| Regla | Archivo |
|-------|---------|
| Arquitectura hexagonal | `ai-tools/rules/architecture.md` |
| Multi-tenancy | `ai-tools/rules/multi-tenant.md` |

Wizard público: rutas `/api/v1/onboarding/*` excluidas de JWT en `src/main.go`. Contrato: `api-docs/openapi.yaml`.
