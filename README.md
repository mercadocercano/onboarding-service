# 🚀 SaaS Multi-Tenant Onboarding Service

Servicio de onboarding simplificado para la plataforma SaaS multi-tenant que maneja el proceso de incorporación de nuevos tenants en **5 pasos optimizados**.

## 🎯 Descripción

Este servicio gestiona el **Onboarding Simplificado** dividido en dos fases:

### 📱 **FASE 1: Onboarding Inicial (3-5 min)**
Registro rápido para generar engagement inmediato:
1. **🏠 Bienvenida** - Landing con expectativas
2. **👤 Registro de Usuario** - Creación de tenant + usuario TENANT_ADMIN  
3. **✉️ Verificación Email** - Código de 6 dígitos
4. **🏪 Configuración Tienda** - Negocio + categorías (sincronizado con PIM)
5. **🎉 Finalización** - Redirección al backoffice

### ⚙️ **FASE 2: Configuración Completa (Backoffice)**
Setup detallado paso a paso con gamificación (futuro).

**Filosofía**: Conseguir el registro rápido, luego guiar la configuración completa.

## 🏗️ Arquitectura Hexagonal

```
src/onboarding/
├── domain/
│   ├── entity/
│   │   ├── onboarding_process.go        # 🎯 Proceso principal de onboarding
│   │   └── step_definition.go           # 📋 Definiciones de pasos
│   └── port/
│       ├── onboarding_repository.go     # 🗄️ Puerto de persistencia
│       ├── iam_client.go                # 🔐 Puerto cliente IAM
│       └── pim_client.go                # 📦 Puerto cliente PIM
├── application/
│   ├── request/
│   │   ├── register_user_request.go     # 📝 DTO registro usuario
│   │   └── setup_store_request.go       # 🏪 DTO configuración tienda
│   ├── response/
│   │   ├── register_user_response.go    # ✅ Respuesta registro
│   │   └── setup_store_response.go      # 🏪 Respuesta configuración
│   └── usecase/
│       ├── register_user_usecase.go     # 🧑‍💼 Caso uso registro completo
│       └── setup_store_usecase.go       # 🏪 Caso uso configuración
└── infrastructure/
    ├── client/
    │   ├── iam_client.go                # 🌐 Cliente HTTP para IAM Service
    │   └── pim_client.go                # 🌐 Cliente HTTP para PIM Service
    ├── persistence/
    │   └── postgres_onboarding_repository.go # 🐘 Repositorio PostgreSQL
    ├── controller/
    │   └── onboarding_controller.go     # 🎮 Controlador REST
    └── config/
        └── setup.go                     # ⚙️ Configuración del módulo
```

## 📊 Base de Datos

### Tablas Principales

```sql
-- Master data: Definiciones de pasos
onboarding_step_definitions (
    id, step_number, step_name, step_title, description,
    has_ui, requires_user_input, can_be_skipped, is_active
)

-- Procesos de onboarding por tenant
onboarding_processes (
    id, tenant_id, user_id, current_step_number, is_completed,
    company_name, business_type, store_size,
    steps_completed, steps_pending, steps_skipped,
    started_at, completed_at
)
```

## 🌐 API Endpoints

### 🎯 Flujo Principal de Onboarding
```http
POST   /api/v1/onboarding/register-user     # 👤 Registro usuario + tenant
POST   /api/v1/onboarding/setup-store       # 🏪 Configuración tienda
POST   /api/v1/onboarding/complete          # 🎉 Completar onboarding
```

### 📊 Información y Configuración
```http
GET    /api/v1/onboarding/business-types    # 📋 Tipos de negocio (desde PIM)
GET    /api/v1/onboarding/categories        # 🏷️ Categorías por tipo de negocio
GET    /api/v1/onboarding/steps             # 📝 Definiciones de pasos
```

### 📈 Estado y Seguimiento
```http
GET    /api/v1/onboarding/process/{id}      # 📊 Estado del proceso
```

## 🔗 Integraciones

### 🔐 IAM Service (Puerto 8080)
- **Crear Tenant** con información del negocio
- **Crear Usuario** con rol `TENANT_ADMIN` automático
- **Actualizar Owner** del tenant creado
- **Rollback automático** en caso de errores

### 📦 PIM Service (Puerto 8090)  
- **Business Types dinámicos** desde `/api/v1/quickstart/business-types`
- **Categorías por negocio** desde `/api/v1/quickstart/categories`
- **Validación en tiempo real** contra datos PIM
- **Templates de quickstart** preparados para backoffice

### 🗃️ Mapeo Business Types

| UI Onboarding | PIM business_type | Nombre PIM |
|---------------|------------------|------------|
| Pinturería | `home-construction` | Hogar y Construcción |
| Ferretería | `home-construction` | Hogar y Construcción |
| Ropa y accesorios | `fashion` | Moda y Vestimenta |
| Electrónicos | `electronics` | Electrónicos y Tecnología |
| Repuestos automotriz | `automotive` | Automotriz y Repuestos |
| ... | ... | ... |
| Polirubro | `polirubro` | Polirubro |

## 🔧 Configuración

### 📄 Variables de Entorno (.env)

```bash
# Server Configuration
PORT=8110

# Database Configuration  
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=onboarding_db
DB_SSLMODE=disable

# External Services
IAM_SERVICE_URL=http://localhost:8080
PIM_SERVICE_URL=http://localhost:8090

# IAM Authentication (for super admin operations)
IAM_SUPER_ADMIN_TOKEN=

# Monitoring
PROMETHEUS_ENABLED=true

# Environment
ENVIRONMENT=development
```

### 🐳 Docker

```bash
# Desarrollo
docker-compose up -d

# Construcción
docker build -t saas-mt-onboarding-service .

# Ejecución
docker run -p 8110:8110 saas-mt-onboarding-service
```

## 🚀 Desarrollo

### ✅ Prerequisitos

- Go 1.22+
- PostgreSQL 15+
- IAM Service (puerto 8080)
- PIM Service (puerto 8090)

### 📦 Instalación

```bash
# 1. Instalar dependencias
go mod tidy

# 2. Configurar variables de entorno
cp .env.example .env
# Editar .env con tus configuraciones

# 3. Crear base de datos
createdb onboarding_db

# 4. Ejecutar el servicio
go run src/main.go
```

### 🧪 Testing

```bash
# Tests unitarios
go test ./...

# Tests con cobertura
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Test de integración
./scripts/test-integration.sh
```

## 📊 Métricas

El servicio expone métricas en `/metrics` para Prometheus:

- ✅ Requests HTTP por endpoint  
- ⏱️ Latencia de respuesta
- ❌ Errores por tipo
- 🗄️ Conexiones de base de datos
- 📈 Progreso de onboarding por paso

## 🔄 Flujo Completo

### 1. 👤 Registro de Usuario
```json
POST /api/v1/onboarding/register-user
{
  "name": "Juan Pérez",
  "email": "juan@empresa.com", 
  "password": "password123",
  "confirm_password": "password123"
}
```
**✅ Resultado**: Tenant creado + Usuario TENANT_ADMIN + Proceso iniciado

### 2. 🏪 Configuración de Tienda  
```json
POST /api/v1/onboarding/setup-store
{
  "process_id": "uuid-del-proceso",
  "store_name": "Ferretería El Martillo",
  "business_type": "home-construction",
  "store_size": "pyme", 
  "selected_categories": ["herramientas", "pinturas", "electricidad"]
}
```
**✅ Resultado**: Tenant actualizado + Configuración PIM preparada

### 3. 🎉 Completar Onboarding
```json
POST /api/v1/onboarding/complete
{
  "process_id": "uuid-del-proceso"
}
```
**✅ Resultado**: Proceso marcado como completado + Redirección a backoffice

## 📈 Objetivos de Performance

- ⏱️ **Tiempo total**: < 10 minutos
- 📈 **Tasa de completación**: > 85%  
- 🎯 **Abandono por paso**: < 15%
- ✅ **Time-to-first-product**: < 30 minutos

## 🔮 Roadmap

### ✅ **Completado**
- [x] Arquitectura hexagonal implementada
- [x] Integración completa con IAM Service
- [x] Integración completa con PIM Service  
- [x] Flujo de 5 pasos optimizado
- [x] Persistencia PostgreSQL con tracking de estado
- [x] Validaciones de negocio y rollback automático
- [x] API REST completa con documentación

### 🚧 **En Desarrollo**
- [ ] Tests unitarios y de integración
- [ ] Documentación API con Swagger
- [ ] Migraciones de base de datos
- [ ] Health checks avanzados

### 🔮 **Futuro (Fase 2)**
- [ ] Verificación de email real con SendGrid
- [ ] Configuración avanzada en backoffice
- [ ] Gamificación del proceso
- [ ] Analytics y métricas de conversión
- [ ] Onboarding personalizado por business type

---

**🌐 Puerto**: `8110` (HTTP) | **🗄️ Base de datos**: `onboarding_db`  
**🛠️ Tecnología**: Go + Gin + PostgreSQL | **🏗️ Arquitectura**: Hexagonal  
**🔗 Dependencias**: IAM Service (8080) + PIM Service (8090) 