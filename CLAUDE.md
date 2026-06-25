# onboarding-service

Guía breve para asistentes de código en el repositorio **onboarding-service** (Mercado Cercano).

## Identidad

Microservicio de **onboarding guiado** para nuevos socios: flujo por pasos (inicio, registro, verificación, tienda, plan, cierre) con persistencia en PostgreSQL e integración con IAM, PIM, notificaciones y tenant.

Todos los servicios de plataforma (IAM, PIM, notification) se consumen vía Kong
(`/iam-service`, `/pim-service`, `/notification-service`).

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

## Memoria persistente (Engram)

Tenés acceso a memoria persistente entre sesiones vía las herramientas MCP de Engram (`mem_save`, `mem_search`, `mem_context`, etc.). Proyecto: **`mercado-cercano`** (memoria unificada con el resto del ecosistema; este service es un polyrepo con su propio `.git`, por eso `.engram/config.json` fija el nombre).

**Cuándo guardar** — sin esperar que te lo pidan:
- Al resolver un bug no trivial: síntoma, causa raíz, fix aplicado.
- Al tomar una decisión de diseño: qué se decidió y por qué.
- Al descubrir un patrón o convención del proyecto que no está documentada.
- Al completar una feature o refactor significativo: qué cambió y dónde.

**Cuándo buscar** — antes de empezar cualquier tarea:
- `mem_context` al inicio de sesión o tras una compaction para recuperar el estado anterior.
- `mem_search` cuando el usuario menciona algo que puede tener historial ("el bug de autenticación", "la migración de la semana pasada").

**Al cerrar sesión**: llamar `mem_session_summary` para dejar un resumen recuperable.
