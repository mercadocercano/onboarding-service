# Documentación — Onboarding Service

Servicio de onboarding guiado de mercado-cercano (puerto 8110). El [README raíz](../README.md) describe el flujo de 6 pasos, el esquema de base de datos y los endpoints. El contrato de API vive en [`api-docs/openapi.yaml`](../api-docs/openapi.yaml).

## Architecture Decision Records

| ADR | Título | Estado | Fecha |
|-----|--------|--------|-------|
| [ADR-001](adr/ADR-001-controller-error-handling-middleware.md) | Manejo centralizado de errores en el controller vía middleware | Aceptado | 2026-01-31 |
| [ADR-002](adr/ADR-002-iam-roles-response-items-roles.md) | Compatibilidad con el formato de respuesta de roles del IAM Service | Aceptado | 2026-01-31 |

## Arquitectura

- [Integración al monorepo SaaS-MT](architecture/integracion-monorepo.md) — cambios de integración, esquema de BD, integraciones verificadas e issues conocidos.

## Setup

- [Instrucciones de inicio](setup/start-instructions.md) — levantar el servicio vía Docker/`make dev-start` y smoke test del flujo.

## Runbooks

- [Comandos útiles](runbooks/comandos-utiles.md) — ejecución local, testing del flujo, queries de BD, troubleshooting y Kong.
- [Migración a mercadocercano](runbooks/migracion-a-mercadocercano.md) — migración oficial del repo cuando haya permisos.

## Guías

- [Refactoring del OnboardingController](guides/controller-refactoring.md) — detalle del patrón de manejo de errores (ver ADR-001).
- [Resumen de implementación HITO 1](guides/resumen-implementacion-hito-1.md) — estado y checklist del HITO 1.
