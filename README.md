# 🚀 SaaS Multi-Tenant Onboarding Service

Servicio de onboarding simplificado para la plataforma SaaS multi-tenant que maneja el proceso completo de incorporación de nuevos socios comerciales al marketplace MercadoCercano.

## 📋 Descripción del Servicio

El **Onboarding Service** gestiona el flujo optimizado de registro de nuevos partners/socios en **6 pasos estratégicos**, diseñado bajo la filosofía: **"conseguir el registro rápido, luego guiar la configuración completa"**.

### ✨ Funcionalidades Principales

- **Registro Multi-Tenant**: Creación automática de tenant + usuario administrador
- **Verificación de Email**: Sistema de códigos de 6 dígitos con expiración
- **Configuración de Tienda**: Setup básico con integración al PIM Service
- **Gestión de Estado Persistente**: Seguimiento completo del progreso del usuario
- **Integración con IAM**: Autenticación y autorización automatizada
- **Notificaciones**: Emails de verificación y bienvenida automáticos
- **Rollback Automático**: Recuperación ante errores en el proceso

### 🔄 Flujo de Onboarding (6 Pasos)

1. **🏠 Bienvenida** - Landing page con información del proceso
2. **👤 Registro de Usuario** - Creación de tenant + usuario `TENANT_ADMIN`
3. **✉️ Verificación Email** - Código de 6 dígitos con validación
4. **🏪 Configuración Tienda** - Datos del negocio + sincronización PIM
5. **📋 Selección de Plan** - Planes disponibles y características
6. **🎉 Finalización** - Proceso completado + email de bienvenida

## 🏗️ Arquitectura Hexagonal

```
src/onboarding/
├── domain/                              # 🎯 Lógica de negocio
│   ├── entity/
│   │   ├── onboarding_process.go        # Proceso principal de onboarding
│   │   ├── step_definition.go           # Definiciones de pasos
│   │   └── verification_code.go         # Códigos de verificación
│   └── port/
│       ├── onboarding_repository.go     # Puerto de persistencia
│       ├── iam_client.go                # Puerto cliente IAM
│       ├── pim_client.go                # Puerto cliente PIM
│       └── notification_client.go       # Puerto cliente Notifications
├── application/                         # 📝 Casos de uso
│   ├── request/
│   │   ├── register_user_request.go     # DTO registro usuario
│   │   ├── setup_store_request.go       # DTO configuración tienda
│   │   ├── verify_email_request.go      # DTO verificación email
│   │   └── complete_onboarding_request.go # DTO completar proceso
│   ├── response/
│   │   ├── register_user_response.go    # Respuesta registro
│   │   ├── setup_store_response.go      # Respuesta configuración
│   │   ├── verify_email_response.go     # Respuesta verificación
│   │   └── complete_onboarding_response.go # Respuesta finalización
│   └── usecase/
│       ├── register_user_usecase.go     # Registro completo usuario
│       ├── setup_store_usecase.go       # Configuración tienda
│       ├── verify_email_usecase.go      # Verificación email
│       └── complete_onboarding_usecase.go # Finalización proceso
└── infrastructure/                      # 🌐 Adaptadores externos
    ├── client/
    │   ├── iam_client.go                # Cliente HTTP para IAM Service
    │   ├── pim_client.go                # Cliente HTTP para PIM Service
    │   └── notification_client.go       # Cliente HTTP para Notifications
    ├── persistence/
    │   └── postgres_onboarding_repository.go # Repositorio PostgreSQL
    ├── controller/
    │   └── onboarding_controller.go     # Controlador REST API
    └── config/
        └── setup.go                     # Configuración del módulo
```

## 🚀 Tecnologías Utilizadas

### Backend Framework
- **Go 1.22+** - Lenguaje principal
- **Gin v1.9** - Framework HTTP/REST
- **PostgreSQL 15+** - Base de datos principal

### Librerías Principales
- **google/uuid v1.6** - Generación de UUIDs
- **lib/pq v1.10** - Driver PostgreSQL
- **prometheus/client_golang v1.22** - Métricas y monitoring
- **go-playground/validator v10.16** - Validación de datos

### Integraciones
- **IAM Service** (Puerto 8080) - Gestión de usuarios y tenants
- **PIM Service** (Puerto 8090) - Catálogo y categorías de productos
- **Notification Service** (Puerto 8282) - Envío de emails

## 📊 Base de Datos

### Esquema Principal

#### **onboarding_step_definitions** - Definiciones maestras de pasos
```sql
CREATE TABLE onboarding_step_definitions (
    id UUID PRIMARY KEY,
    step_number INTEGER UNIQUE NOT NULL,
    step_name VARCHAR(100) NOT NULL,
    step_title VARCHAR(200) NOT NULL,
    description TEXT,
    is_required BOOLEAN DEFAULT true,
    has_ui BOOLEAN DEFAULT true,
    requires_user_input BOOLEAN DEFAULT true,
    can_be_skipped BOOLEAN DEFAULT false,
    display_order INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### **onboarding_processes** - Procesos de onboarding por tenant
```sql
CREATE TABLE onboarding_processes (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    current_step_number INTEGER DEFAULT 1,
    is_completed BOOLEAN DEFAULT false,
    company_name VARCHAR(255),
    business_type VARCHAR(100),
    store_size VARCHAR(50),
    steps_completed JSONB DEFAULT '[]'::jsonb,
    steps_pending JSONB DEFAULT '[1,2,3,4,5,6]'::jsonb,
    steps_skipped JSONB DEFAULT '[]'::jsonb,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

#### **verification_codes** - Códigos de verificación de email
```sql
CREATE TABLE verification_codes (
    id UUID PRIMARY KEY,
    process_id UUID NOT NULL REFERENCES onboarding_processes(id),
    user_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    code VARCHAR(6) NOT NULL,
    is_used BOOLEAN DEFAULT false,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    used_at TIMESTAMP WITH TIME ZONE
);
```

## 🌐 API Endpoints

### 🎯 Flujo Principal de Onboarding

#### **POST /api/v1/onboarding/start**
Inicializar proceso de onboarding
```json
Request: {}
Response: {
  "success": true,
  "process_id": "uuid",
  "current_step": 1,
  "next_step_url": "/onboarding/registro"
}
```

#### **POST /api/v1/onboarding/register-user**
Registro de usuario y creación de tenant
```json
Request: {
  "process_id": "uuid",
  "name": "Juan Pérez",
  "email": "juan@tienda.com",
  "password": "password123",
  "phone": "+54911234567"
}
Response: {
  "success": true,
  "message": "Usuario registrado exitosamente",
  "process_id": "uuid",
  "tenant_id": "uuid",
  "user_id": "uuid",
  "next_step": 3,
  "verification_code_sent": true
}
```

#### **POST /api/v1/onboarding/verify-email**
Verificación de código de email
```json
Request: {
  "process_id": "uuid",
  "verification_code": "123456"
}
Response: {
  "success": true,
  "message": "Email verificado exitosamente",
  "next_step": 4
}
```

#### **POST /api/v1/onboarding/setup-store**
Configuración de tienda (simplificado - sin categorías obligatorias)
```json
Request: {
  "process_id": "uuid",
  "store_name": "Ferretería El Martillo",
  "business_type": "home-construction",
  "store_size": "pyme",
  "selected_categories": [] // Opcional
}
Response: {
  "success": true,
  "message": "Tienda configurada exitosamente. Las categorías se configurarán en el backoffice.",
  "process_id": "uuid",
  "tenant_id": "uuid",
  "next_step": 5,
  "store_data": {
    "name": "Ferretería El Martillo",
    "business_type": {
      "id": "home-construction",
      "name": "Hogar y Construcción"
    },
    "store_size": "pyme"
  },
  "recommended_plan": "professional"
}
```

#### **POST /api/v1/onboarding/complete**
Finalizar proceso de onboarding
```json
Request: {
  "process_id": "uuid"
}
Response: {
  "success": true,
  "message": "Onboarding completado exitosamente",
  "process_id": "uuid",
  "tenant_id": "uuid",
  "completed_at": "2025-06-22T19:34:27Z",
  "dashboard_url": "https://backoffice.mercadocercano.com/dashboard",
  "welcome_email_sent": true
}
```

### 📊 Endpoints de Información

#### **GET /api/v1/onboarding/business-types**
Tipos de negocio disponibles (desde PIM Service)
```json
Response: {
  "success": true,
  "business_types": [
    {
      "id": "retail",
      "name": "Comercio Minorista",
      "description": "Tiendas de venta al por menor...",
      "icon": "store",
      "created_at": "2025-06-20T18:40:28Z"
    }
  ]
}
```

#### **GET /api/v1/onboarding/categories**
Categorías por tipo de negocio
```json
Query: ?business_type=retail
Response: {
  "success": true,
  "business_type": "retail",
  "categories": [
    {
      "id": "alimentacion",
      "name": "Alimentación",
      "description": "Productos alimenticios"
    }
  ]
}
```

#### **GET /api/v1/onboarding/steps**
Definiciones de pasos del onboarding
```json
Response: {
  "success": true,
  "steps": [
    {
      "step_number": 1,
      "step_name": "welcome",
      "step_title": "Bienvenida",
      "description": "Introducción al proceso",
      "is_required": true,
      "has_ui": true
    }
  ]
}
```

#### **GET /api/v1/onboarding/process/{id}/status**
Estado actual del proceso
```json
Response: {
  "success": true,
  "process_id": "uuid",
  "current_step": 4,
  "steps_completed": [1, 2, 3],
  "steps_pending": [4, 5, 6],
  "is_completed": false,
  "next_step_url": "/onboarding/configurar-tienda",
  "company_name": "Mi Tienda",
  "business_type": "retail"
}
```

## 🔗 Integraciones con Servicios

### 🔐 IAM Service (Puerto 8080)

#### **Funcionalidades utilizadas:**
- **POST /api/v1/tenants** - Crear tenant con información del negocio
- **POST /api/v1/users** - Crear usuario con rol `TENANT_ADMIN`
- **PUT /api/v1/tenants/{id}/owner** - Asignar owner al tenant
- **GET /api/v1/users/{id}** - Obtener información del usuario

#### **Flujo de registro:**
1. Crear tenant con datos del negocio
2. Crear usuario administrador
3. Asignar usuario como owner del tenant
4. Rollback automático en caso de error

### 📦 PIM Service (Puerto 8090)

#### **Endpoints utilizados:**
- **GET /api/v1/quickstart/business-types** - Tipos de negocio dinámicos
- **GET /api/v1/quickstart/categories** - Categorías por tipo de negocio

#### **Mapeo de Business Types:**
```json
{
  "PIM Response": {
    "id": "retail",                    // ✅ Identificador funcional
    "name": "Comercio Minorista",      // ✅ Nombre descriptivo
    "description": "Tiendas de venta...", // ✅ Descripción completa
    "icon": "store",                   // ✅ Icono UI
    "createdAt": "2025-06-20T18:40:28Z" // ✅ camelCase
  },
  "Onboarding Mapping": {
    "id": "retail",                    // ✅ Mapeo directo
    "name": "Comercio Minorista",      // ✅ Sin transformación
    "is_active": true,                 // ✅ Default si no existe
    "color": "",                       // ✅ Opcional con omitempty
    "sort_order": 0                    // ✅ Default
  }
}
```

### 📧 Notification Service (Puerto 8282)

#### **Emails enviados:**
- **EMAIL_VERIFICATION** - Código de verificación de 6 dígitos
- **WELCOME** - Email de bienvenida al completar onboarding

#### **Estructura de envío:**
```json
{
  "type": "email",
  "action": "EMAIL_VERIFICATION",
  "recipient": "user@email.com",
  "data": {
    "name": "Usuario",
    "token": "123456",
    "company": "MercadoCercano",
    "expiry_time": "15 minutos"
  },
  "async": false
}
```

## 🛠️ Instalación y Configuración

### ✅ Prerrequisitos

- **Go 1.22+**
- **PostgreSQL 15+**
- **IAM Service** funcionando en puerto 8080
- **PIM Service** funcionando en puerto 8090
- **Notification Service** funcionando en puerto 8282

### 📦 Instalación Local

```bash
# 1. Clonar repositorio
git clone <repository-url>
cd services/saas-mt-onboarding-service

# 2. Instalar dependencias
go mod tidy

# 3. Configurar variables de entorno
cp .env.example .env
# Editar .env con configuraciones específicas

# 4. Crear base de datos
createdb onboarding_db

# 5. Ejecutar migraciones
psql -d onboarding_db -f migrations/001_initial_schema.sql
psql -d onboarding_db -f migrations/002_seed_step_definitions.sql
psql -d onboarding_db -f migrations/003_create_verification_codes.sql
psql -d onboarding_db -f migrations/003_create_indexes.sql

# 6. Ejecutar servicio
go run src/main.go
```

### 📄 Variables de Entorno

```bash
# Configuración del servidor
PORT=8110
ENVIRONMENT=development

# Base de datos PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=onboarding_db
DB_SSLMODE=disable

# Servicios externos
IAM_SERVICE_URL=http://localhost:8080
PIM_SERVICE_URL=http://localhost:8090
NOTIFICATION_SERVICE_URL=http://localhost:8282

# Autenticación IAM (para operaciones super admin)
IAM_SUPER_ADMIN_TOKEN=your-super-admin-token

# Monitoring
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=2116

# Timeouts (segundos)
HTTP_CLIENT_TIMEOUT=30
DB_CONNECTION_TIMEOUT=10
```

## 🐳 Docker

### Dockerfile incluido para desarrollo:

```dockerfile
# Imagen multi-stage optimizada
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o onboarding-service src/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/onboarding-service .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8110
CMD ["./onboarding-service"]
```

### Comandos Docker:

```bash
# Construcción
docker build -t saas-mt-onboarding-service .

# Ejecución
docker run -p 8110:8110 \
  -e DB_HOST=host.docker.internal \
  -e IAM_SERVICE_URL=http://host.docker.internal:8080 \
  -e PIM_SERVICE_URL=http://host.docker.internal:8090 \
  saas-mt-onboarding-service

# Con Docker Compose
docker-compose up -d onboarding-service
```

## 🧪 Testing y Validación

### Scripts de prueba incluidos:

```bash
# Prueba completa del flujo de onboarding
./test-simplified-setup-store.sh

# Prueba específica de business types
./debug-business-types-v2.sh

# Prueba rápida de health check
./quick-test.sh
```

### Escenarios de prueba cubiertos:

1. ✅ **Health check** del servicio
2. ✅ **Business types** desde PIM Service
3. ✅ **Registro completo** de usuario
4. ✅ **Verificación de email** con código
5. ✅ **Setup de tienda** sin categorías (flujo principal)
6. ✅ **Setup de tienda** con categorías (backward compatible)
7. ✅ **Completar onboarding** con email de bienvenida
8. ✅ **Validación de errores** y casos edge

### Tests unitarios:

```bash
# Ejecutar todos los tests
go test ./...

# Tests con cobertura
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Tests específicos por módulo
go test ./src/onboarding/application/usecase/...
go test ./src/onboarding/infrastructure/client/...
```

## 📊 Monitoreo y Métricas

### Métricas de Prometheus expuestas:

```
# Métricas del servicio
onboarding_requests_total{method, endpoint, status}
onboarding_request_duration_seconds{method, endpoint}
onboarding_processes_created_total
onboarding_processes_completed_total
onboarding_verification_codes_sent_total
onboarding_verification_codes_validated_total

# Health checks
onboarding_service_up
onboarding_database_connection_status
onboarding_iam_service_status
onboarding_pim_service_status
```

### Endpoints de salud:

```bash
# Health check básico
curl http://localhost:8110/health

# Métricas de Prometheus
curl http://localhost:2116/metrics

# Estado detallado
curl http://localhost:8110/api/v1/health/detailed
```

## 🚀 Despliegue

### Desarrollo:
```bash
go run src/main.go
```

### Staging:
```bash
go build -o onboarding-service src/main.go
./onboarding-service
```

### Producción:
```bash
docker build -t onboarding-service:latest .
docker run -d -p 8110:8110 \
  --name onboarding-service \
  -e ENVIRONMENT=production \
  onboarding-service:latest
```

## 📈 Mejoras y Evolución

### ✅ Cambios Recientes Implementados:

1. **Setup Store Simplificado** - Categorías opcionales en onboarding inicial
2. **Mapeo Business Types Corregido** - Sincronización real con PIM Service
3. **Email de Bienvenida** - Automático al completar proceso
4. **Verificación de Email** - Sistema robusto con códigos de 6 dígitos
5. **Rollback Automático** - Recuperación ante errores de integración
6. **Logging Mejorado** - Debugging completo del flujo

### 🚧 Roadmap:

- [ ] **Gamificación**: Setup wizard paso a paso en backoffice
- [ ] **A/B Testing**: Diferentes flujos de onboarding  
- [ ] **Analytics**: Métricas de conversión y abandono
- [ ] **PWA**: Aplicación móvil para onboarding
- [ ] **Webhooks**: Notificaciones a sistemas externos
- [ ] **Multi-idioma**: Soporte para i18n
- [ ] **Templates**: Configuraciones pre-definidas por industry

### 🎯 KPIs del Servicio:

- **Tiempo promedio de onboarding**: < 5 minutos
- **Tasa de conversión**: > 85%
- **Tasa de abandono por paso**: < 10%
- **Tiempo de respuesta API**: < 200ms P95
- **Disponibilidad del servicio**: > 99.9%

## 🤝 Contribución

### Flujo de desarrollo:

1. **Fork** del repositorio
2. **Crear rama** para feature (`git checkout -b feature/amazing-feature`)
3. **Commit** con mensajes descriptivos (`git commit -m 'Add amazing feature'`)
4. **Push** a la rama (`git push origin feature/amazing-feature`)
5. **Pull Request** con descripción detallada

### Estándares de código:

- **Go fmt** para formateo
- **golint** para linting
- **Tests** obligatorios para nuevas funcionalidades
- **Documentación** actualizada en README
- **Logs estructurados** con contexto

## 🆘 Soporte y Contacto

### Para dudas técnicas:
- **GitHub Issues**: Reportar bugs y solicitar features
- **Documentación**: Este README como fuente de verdad
- **Logs**: Revisar logs del servicio para debugging

### Contacto del equipo:
- **Desarrollador Principal**: Leonardo Pegorín
- **Email**: desarrollo@mercadocercano.com
- **Slack**: #onboarding-service

## 📄 Licencia

Este proyecto es propiedad de **MercadoCercano**. Todos los derechos reservados.

---

**Versión**: 2.0.0  
**Última actualización**: Junio 2025  
**Mantenido por**: Equipo de Desarrollo MercadoCercano  
**Puerto**: 8110  
**Base de datos**: PostgreSQL 15+