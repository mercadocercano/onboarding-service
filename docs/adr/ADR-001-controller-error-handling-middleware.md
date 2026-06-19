---
adr: ADR-001
status: accepted
skills:
  implement:
    - dev/hexagonal-go
  verify:
    - dev/go-hex-audit
    - dev/code-reviewer
  pending:   # skills aún inexistentes — ver épica EPIC roadmap
    - dev/api-error-contract
---
# ADR-001: Manejo centralizado de errores en el controller vía middleware

**Estado**: Aceptado
**Fecha**: 2026-01-31
**Contexto**: El `OnboardingController` repetía en cada handler (~10 métodos) el mismo bloque de binding de JSON, logging con `log.Printf` y construcción manual de respuestas de error con `gin.H`. Esto generaba ~425 líneas con ~50 líneas duplicadas, respuestas inconsistentes entre endpoints y manejo de status codes manual y propenso a errores de copy-paste.

## Decisión

Centralizamos el manejo de errores y el binding en un middleware compartido. Los handlers delegan en helpers:

- `middleware.ShouldBindJSONWithError(ctx, &req)` — binding de JSON con manejo de error automático.
- `middleware.AbortWithError(ctx, err)` / `AbortWithBusinessError(ctx, err)` — abortan con respuesta estandarizada.
- `ErrorHandlerMiddleware()` registrado globalmente en `main.go`.

Cada handler queda reducido a binding + ejecución del use case + `ctx.JSON(http.StatusOK, response)`. El controller pasó de ~425 a ~240 líneas (-44%).

Componentes asociados:
- `shared/middleware/error_handler.go` — manejo centralizado y logging estructurado.
- `shared/response/result.go` — estructura `Result` consistente y mapeo de errores.

## Alternativas consideradas

| Opción | Por qué no |
|--------|-----------|
| Mantener el patrón manual en cada handler | Duplicación, respuestas inconsistentes, alta superficie de testing y errores de copy-paste. |
| Wrapper genérico por handler (closures) | Más implícito y difícil de leer que un middleware explícito; no resuelve el logging estructurado central. |

## Consecuencias

**Positivas**: -44% de código en el controller, respuestas y logging consistentes, agregar endpoints nuevos es trivial, menor superficie de testing (~-90% de código repetitivo).
**Negativas / trade-offs**: El flujo de error es menos explícito en el sitio de llamada — hay que conocer el contrato del middleware para entender qué pasa ante un error.
**Neutral**: El patrón `Result` (`shared/response/result.go`) queda disponible para que los use cases lo adopten gradualmente; aún no es obligatorio.

> Detalle de la transformación (código antes/después, métricas): ver [../guides/controller-refactoring.md](../guides/controller-refactoring.md).
