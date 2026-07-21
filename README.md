# 🚀 SaaS Multi-Tenant Onboarding Service

Servicio de onboarding simplificado para la plataforma SaaS multi-tenant que maneja el proceso completo de incorporación de nuevos socios comerciales al marketplace MercadoCercano.

## 📚 Documentación

Ver [`docs/`](docs/README.md) para ADRs, arquitectura de integración, setup, runbooks y guías. El contrato de API está en [`api-docs/openapi.yaml`](api-docs/openapi.yaml).

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
│   ├── port/                            # Driven ports declarados por application
│   │   └── verification.go              # Generación de códigos de verificación
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
    │   ├── notification_client.go       # Cliente HTTP para Notifications
    │   └── verification_generator.go    # Adaptador del driven port de códigos
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
  "phone": "+549****4567"
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
  "recommended_plan": "premium"
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

### 🔐 IAM Service

URL por defecto: `http://lab-kong:8000/iam-service` (Kong). Puede sobreescribirse con `IAM_SERVICE_URL`.

#### **Funcionalidades utilizadas:**
- **POST /api/v1/tenants** - Crear tenant con información del negocio
- **GET /api/v1/tenants/{id}** - Obtener tenant
- **PUT /api/v1/tenants/{id}** - Actualizar tenant
- **DELETE /api/v1/tenants/{id}** - Rollback de tenant
- **POST /api/v1/users** - Crear usuario con rol `TENANT_ADMIN`
- **GET /api/v1/users/{id}** - Obtener información del usuario
- **GET /api/v1/roles** - Obtener roles
- **POST /api/v1/auth/login** - Login automático post-registro

#### **Flujo de registro:**
1. Crear tenant con datos del negocio
2. Crear usuario administrador
3. Crear proceso de onboarding local
4. Enviar email de verificación
5. Rollback automático en caso de error

#### **Autenticación S2S**
- Se usa `X-API-Key` con la scoped key `S2S_KEY_ONBOARDING`.
- Fallback legacy a la god-key `S2S_API_KEY` si la scoped key no está presente.
- El header `X-Tenant-ID` viaja con el `SYSTEM_TENANT_ID` para operaciones S2S.

### 📦 PIM Service

URL por defecto: `http://lab-kong:8000/pim-service`.

#### **Endpoints utilizados:**
- **GET /api/v1/quickstart/business-types** - Tipos de negocio dinámicos
- **GET /api/v1/quickstart/categories** - Categorías por tipo de negocio

#### **Mapeo de Business Types:**
```json
{
  "PIM Response": {
    "id": "retail",
    "name": "Comercio Minorista",
    "description": "Tiendas de venta al por menor...",
    "icon": "store",
    "createdAt": "2025-06-20T18:40:28Z"
  },
  "Onboarding Mapping": {
    "id": "retail",
    "name": "Comercio Minorista",
    "is_active": true,
    "color": "",
    "sort_order": 0
  }
}
```

### 📧 Notification Service

URL por defecto: `http://lab-kong:8000/notifications/api/v1`.

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
- **IAM Service** funcionando
- **PIM Service** funcionando
- **Notification Service** funcionando

### 📦 Instalación Local

```bash
# 1. Entrar al servicio
cd services/onboarding-service

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

# 6. Ejecutar servicio
go run src/main.go
```

### 📄 Variables de Entorno

Ver `.env.example` para la lista completa. Las más relevantes:

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

# Servicios externos (Kong es el punto de entrada por defecto)
IAM_SERVICE_URL=http://lab-kong:8000/iam-service
PIM_SERVICE_URL=http://lab-kong:8000/pim-service
TENANT_SERVICE_URL=http://lab-kong:8000/tenant-service
NOTIFICATION_SERVICE_URL=http://lab-kong:8000/notifications/api/v1

# Identidad de sistema
SERVICE_NAMESPACE=mc
SYSTEM_TENANT_ID=123e4567-e89b-12d3-a456-426614174003

# Autenticación IAM (fallback legacy; preferir S2S scoped keys)
JWT_SECRET=your-jwt-secret
IAM_SUPER_ADMIN_TOKEN=your-super-admin-token

# S2S scoped keys
S2S_KEY_ONBOARDING=your-32-byte-hex-key-minimum
# Fallback legacy god-key (se usa solo si S2S_KEY_ONBOARDING no está)
S2S_API_KEY=your-legacy-god-key

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
docker build -t onboarding-service .

# Ejecución
docker run -p 8110:8110 \
  -e DB_HOST=host.docker.internal \
  -e IAM_SERVICE_URL=http://host.docker.internal:8080 \
  -e PIM_SERVICE_URL=http://host.docker.internal:8090 \
  onboarding-service

# Con Docker Compose
docker-compose up -d onboarding-service
```

## 🧪 Testing y Validación

```bash
# Ejecutar todos los tests
go test ./...

# Prueba completa del flujo de onboarding (si existe script local)
./test-simplified-setup-store.sh
```

## 🔒 Notas de Seguridad

- No commitear secrets. `S2S_KEY_ONBOARDING`, `S2S_API_KEY`, `JWT_SECRET` e `IAM_SUPER_ADMIN_TOKEN` deben venir de variables de entorno.
- La scoped key `S2S_KEY_ONBOARDING` debe tener al menos 16 bytes (recomendado 32+ bytes hex) y estar registrada en IAM con los scopes `tenant:provision` y `tenant:admin`.
- Operaciones cross-tenant o de admin global ya no requieren `system:admin` para onboarding; IAM expone GET/PUT/DELETE `/tenants/:id` bajo el grupo tenant-scoped.
