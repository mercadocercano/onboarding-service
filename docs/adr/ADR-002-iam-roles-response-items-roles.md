# ADR-002: Compatibilidad con el formato de respuesta de roles del IAM Service

**Estado**: Aceptado
**Fecha**: 2026-01-31
**Contexto**: Al integrar el servicio al monorepo de mercado-cercano, el cliente HTTP de IAM (`src/onboarding/infrastructure/client/iam_client.go`) deserializaba la respuesta de roles esperando un campo `roles`. El IAM Service del monorepo devuelve paginación estándar con el campo `items`, lo que rompía la obtención del rol `TENANT_ADMIN` durante el registro.

## Decisión

El cliente de IAM deserializa ambos formatos, priorizando `items` (formato actual del monorepo) y manteniendo `roles` por compatibilidad hacia atrás con el repo original:

```go
var result struct {
    Items []*port.RoleResponse `json:"items"` // formato monorepo (paginación estándar)
    Roles []*port.RoleResponse `json:"roles"` // backward compatibility (repo original)
}
```

## Alternativas consideradas

| Opción | Por qué no |
|--------|-----------|
| Cambiar solo a `items` | Rompería la compatibilidad si se vuelve a apuntar al IAM del repo original; cambio innecesariamente destructivo durante la migración. |
| Modificar el IAM Service para devolver `roles` | Toca un servicio compartido y rompe a otros consumidores que ya esperan paginación estándar con `items`. |

## Consecuencias

**Positivas**: La integración con el IAM del monorepo funciona sin tocar el servicio compartido; se conserva compatibilidad con el formato anterior.
**Negativas / trade-offs**: El cliente acarrea dos campos para el mismo dato; si el formato `roles` se abandona definitivamente, queda código muerto a limpiar.
**Neutral**: Refuerza la convención del ecosistema de que los listados del IAM usan paginación con `items`.

> Contexto de integración completo: ver [../architecture/integracion-monorepo.md](../architecture/integracion-monorepo.md).
