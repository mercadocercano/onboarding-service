# Comandos Útiles - Onboarding Service

## Inicio Rápido

### Iniciar servicio localmente (desarrollo)
```bash
cd services/onboarding-service

PORT=8110 \
DB_HOST=localhost \
DB_PORT=5432 \
DB_USER=postgres \
DB_PASSWORD=postgres \
DB_NAME=onboarding_db \
DB_SSLMODE=disable \
IAM_SERVICE_URL=http://localhost:8080 \
PIM_SERVICE_URL=http://localhost:8090 \
NOTIFICATION_SERVICE_URL=http://localhost:8282 \
go run src/main.go
```

### Compilar binario
```bash
cd services/onboarding-service
go build -o onboarding-binary src/main.go
./onboarding-binary
```

### Detener servicio
```bash
pkill -f "go run src/main.go"
# o
pkill -f "onboarding-binary"
```

---

## Testing

### Test automático completo (HITO 1)
```bash
/tmp/test-hito-1-completo.sh
```

### Test manual paso a paso

#### 1. Health Check
```bash
curl http://localhost:8110/health
```

#### 2. Iniciar proceso
```bash
curl -X POST http://localhost:8110/api/v1/onboarding/start
```

#### 3. Registrar usuario
```bash
PROCESS_ID="<tu-process-id>"
TEST_EMAIL="test@example.com"

curl -X POST http://localhost:8110/api/v1/onboarding/register-user \
  -H "Content-Type: application/json" \
  -d "{
    \"process_id\": \"$PROCESS_ID\",
    \"name\": \"Usuario Test\",
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"password123\",
    \"confirm_password\": \"password123\"
  }"
```

#### 4. Obtener código de verificación
```bash
docker exec mc-postgres psql -U postgres -d onboarding_db \
  -c "SELECT code FROM verification_codes WHERE email='$TEST_EMAIL' ORDER BY created_at DESC LIMIT 1;"
```

#### 5. Verificar email
```bash
VERIFICATION_CODE="123456"

curl -X POST http://localhost:8110/api/v1/onboarding/verify-email \
  -H "Content-Type: application/json" \
  -d "{
    \"process_id\": \"$PROCESS_ID\",
    \"verification_code\": \"$VERIFICATION_CODE\"
  }"
```

#### 6. Setup Store
```bash
curl -X POST http://localhost:8110/api/v1/onboarding/setup-store \
  -H "Content-Type: application/json" \
  -d "{
    \"process_id\": \"$PROCESS_ID\",
    \"store_name\": \"Mi Tienda\",
    \"business_type\": \"hardware_store\",
    \"store_size\": \"pyme\",
    \"selected_categories\": []
  }"
```

#### 7. Select Plan
```bash
curl -X POST http://localhost:8110/api/v1/onboarding/select-plan \
  -H "Content-Type: application/json" \
  -d "{
    \"process_id\": \"$PROCESS_ID\"
  }"
```

#### 8. Complete Onboarding
```bash
curl -X POST http://localhost:8110/api/v1/onboarding/complete \
  -H "Content-Type: application/json" \
  -d "{
    \"process_id\": \"$PROCESS_ID\"
  }"
```

#### 9. Login IAM
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"password123\",
    \"provider\": \"LOCAL\"
  }"
```

#### 10. Access Quickstart
```bash
TENANT_ID="<tenant-id-from-register>"
ACCESS_TOKEN="<token-from-login>"

curl -X GET http://localhost:8090/api/v1/business-types \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID"
```

---

## Base de Datos

### Conectar a onboarding_db
```bash
docker exec -it mc-postgres psql -U postgres -d onboarding_db
```

### Queries útiles

#### Ver procesos de onboarding
```sql
SELECT id, current_step_number, is_completed, company_name, business_type, created_at 
FROM onboarding_processes 
ORDER BY created_at DESC 
LIMIT 10;
```

#### Ver códigos de verificación
```sql
SELECT id, email, code, is_used, expires_at, created_at 
FROM verification_codes 
ORDER BY created_at DESC 
LIMIT 10;
```

#### Ver definiciones de pasos
```sql
SELECT step_number, step_name, step_title, is_required 
FROM onboarding_step_definitions 
ORDER BY step_number;
```

#### Ver migraciones ejecutadas
```sql
SELECT * FROM schema_migrations ORDER BY id;
```

#### Limpiar datos de test
```sql
-- Eliminar procesos de test
DELETE FROM onboarding_processes WHERE company_name LIKE '%Test%';

-- Eliminar códigos de verificación expirados
DELETE FROM verification_codes WHERE expires_at < NOW();
```

---

## Troubleshooting

### Puerto 8110 ocupado
```bash
# Ver qué proceso está usando el puerto
lsof -i :8110

# Matar proceso
kill -9 <PID>

# O matar todos los go run
pkill -f "go run"
```

### Servicio no conecta a base de datos
```bash
# Verificar que PostgreSQL esté corriendo
docker ps | grep postgres

# Verificar que la BD existe
docker exec mc-postgres psql -U postgres -l | grep onboarding
```

### Servicio no conecta a IAM
```bash
# Verificar que IAM esté corriendo
curl http://localhost:8080/health

# Verificar conectividad
nc -zv localhost 8080
```

### Ver logs en tiempo real
```bash
tail -f /tmp/onboarding.log
```

### Rebuild completo
```bash
cd services/onboarding-service
go clean -cache
go build -o onboarding-binary src/main.go
./onboarding-binary
```

---

## Integración con Kong (a través de gateway)

### A través de Kong Gateway (producción)
```bash
# Health check
curl http://localhost:8001/onboarding/health

# Start proceso
curl -X POST http://localhost:8001/onboarding/api/v1/onboarding/start

# Registro
curl -X POST http://localhost:8001/onboarding/api/v1/onboarding/register-user \
  -H "Content-Type: application/json" \
  -d '{...}'
```

---

## Información Técnica

### Endpoints disponibles
- `GET /health` - Health check
- `POST /api/v1/onboarding/start` - Iniciar proceso
- `POST /api/v1/onboarding/register-user` - Registrar usuario
- `POST /api/v1/onboarding/verify-email` - Verificar email
- `POST /api/v1/onboarding/resend-verification` - Reenviar código
- `POST /api/v1/onboarding/setup-store` - Configurar tienda
- `POST /api/v1/onboarding/select-plan` - Seleccionar plan
- `POST /api/v1/onboarding/complete` - Completar onboarding
- `GET /api/v1/onboarding/business-types` - Tipos de negocio
- `GET /api/v1/onboarding/steps` - Definiciones de pasos
- `GET /api/v1/onboarding/process/:id` - Estado del proceso

### Dependencias externas
- IAM Service (http://localhost:8080) - REQUERIDO
- PIM Service (http://localhost:8090) - REQUERIDO
- Notification Service (http://localhost:8282) - OPCIONAL
- PostgreSQL (localhost:5432) - REQUERIDO

---

## Migración a Producción

### 1. Fix Docker Build
Actualmente el servicio corre con `go run`. Para producción:
- Verificar Dockerfile
- O usar Dockerfile.dev alternativo

### 2. Configurar Notification Service
- Agregar SMTP credentials
- O usar servicio de email (SendGrid, SES, etc.)

### 3. Variables de Entorno Producción
```bash
ENVIRONMENT=production
IAM_SERVICE_URL=https://api.mercadocercano.com/iam
PIM_SERVICE_URL=https://api.mercadocercano.com/pim
NOTIFICATION_SERVICE_URL=https://api.mercadocercano.com/notifications
```

---

**Última actualización:** 31 Enero 2026  
**Mantenido por:** Equipo SaaS Multi-Tenant

