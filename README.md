# 🚀 SaaS Multi-Tenant Onboarding Service

Servicio de onboarding para la plataforma SaaS multi-tenant que maneja el proceso de incorporación de nuevos tenants.

## 🎯 Descripción

Este servicio gestiona el proceso completo de onboarding para nuevos tenants, incluyendo:

- **Configuración inicial del tenant**: Setup básico y configuración de preferencias
- **Tipos de negocio**: Gestión de categorías de negocio y mapeo de atributos
- **Pasos de onboarding**: Flujo guiado para la configuración del tenant
- **Validaciones**: Verificación de completitud del proceso de onboarding

## 🏗️ Arquitectura

Construido siguiendo los principios de **Arquitectura Hexagonal**:

```
src/onboarding/
├── domain/
│   ├── entity/           # Entidades de negocio
│   ├── port/            # Interfaces (contratos)
│   └── exception/       # Excepciones del dominio
├── application/
│   ├── usecase/         # Casos de uso
│   ├── request/         # DTOs de entrada
│   └── response/        # DTOs de salida
└── infrastructure/
    ├── persistence/     # Repositorios PostgreSQL
    ├── controller/      # Controladores HTTP
    └── config/         # Configuración del módulo
```

## 📦 Entidades

### TenantOnboarding
- Gestión del proceso de onboarding por tenant
- Estados: `pending`, `in_progress`, `completed`, `cancelled`
- Tracking de progreso y metadatos

### BusinessType
- Tipos de negocio disponibles (e-commerce, marketplace, B2B, etc.)
- Configuraciones específicas por tipo
- Mapeo de atributos y categorías

### OnboardingStep
- Pasos individuales del proceso de onboarding
- Dependencias entre pasos
- Validaciones y requisitos

## 🌐 API Endpoints

### TenantOnboarding
```
GET    /api/v1/onboarding               # Listar procesos de onboarding
GET    /api/v1/onboarding/{id}          # Obtener proceso específico
POST   /api/v1/onboarding               # Iniciar nuevo onboarding
PUT    /api/v1/onboarding/{id}          # Actualizar proceso
DELETE /api/v1/onboarding/{id}          # Cancelar proceso
```

### BusinessType
```
GET    /api/v1/business-types           # Listar tipos de negocio
GET    /api/v1/business-types/{id}      # Obtener tipo específico
POST   /api/v1/business-types           # Crear nuevo tipo
PUT    /api/v1/business-types/{id}      # Actualizar tipo
```

### OnboardingStep
```
GET    /api/v1/onboarding-steps         # Listar pasos de onboarding
GET    /api/v1/onboarding-steps/{id}    # Obtener paso específico
POST   /api/v1/onboarding-steps         # Crear nuevo paso
PUT    /api/v1/onboarding-steps/{id}    # Actualizar paso
```

## 🔧 Configuración

### Variables de Entorno

```bash
# Base de datos
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=onboarding_db
DB_SSLMODE=disable

# Servidor
PORT=8080

# Métricas
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=2115

# Modo
GIN_MODE=debug
```

### Docker

```bash
# Desarrollo
docker-compose up -d

# Construcción
docker build -t saas-mt-onboarding-service .

# Ejecución
docker run -p 8110:8080 saas-mt-onboarding-service
```

## 🚀 Desarrollo

### Prerequisitos

- Go 1.22+
- PostgreSQL 15+
- Docker (opcional)

### Instalación

```bash
# Clonar dependencias
go mod download

# Ejecutar migraciones
./scripts/run-migrations.sh

# Ejecutar el servicio
go run src/main.go
```

### Testing

```bash
# Tests unitarios
go test ./...

# Tests con cobertura
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📊 Métricas

El servicio expone métricas en `/metrics` (puerto 2115) para Prometheus:

- Requests HTTP por endpoint
- Latencia de respuesta
- Errores por tipo
- Conexiones de base de datos

## 🔄 Integración

### Con API Gateway (Kong)

```yaml
services:
  - name: onboarding-service
    url: http://onboarding-service:8080
    routes:
      - name: onboarding-route
        paths:
          - /onboarding/api/v1
```

### Con otros servicios

- **IAM Service**: Validación de tenants y usuarios
- **PIM Service**: Configuración inicial de productos y categorías

## 📝 TODO

- [ ] Implementar casos de uso específicos
- [ ] Agregar validaciones de negocio
- [ ] Crear migraciones de base de datos
- [ ] Implementar tests unitarios
- [ ] Documentación API con Swagger
- [ ] Métricas personalizadas
- [ ] Health checks avanzados

---

**Puerto**: `8110` (HTTP) | `2115` (Métricas)  
**Base de datos**: `onboarding_db`  
**Tecnología**: Go + Gin + PostgreSQL  
**Arquitectura**: Hexagonal 