# Instrucciones para Iniciar Onboarding Service

## Estado Actual
- ✅ Repositorio clonado en `services/onboarding-service/`
- ✅ Configuración agregada a `docker-compose.services.yml`
- ✅ Ruta configurada en Kong Gateway (`/onboarding/*`)
- ⏳ Esperando que Docker Desktop esté corriendo

## Pasos para Iniciar

### 1. Iniciar Docker Desktop
```bash
# Abrir Docker Desktop manualmente o:
open -a Docker
```

### 2. Esperar a que Docker esté listo
Verificar que aparezca el ícono de Docker en la barra de menú y esté en verde.

### 3. Iniciar los servicios
```bash
cd /Users/hornosg/MyProjects/saas-mt
make dev-start
```

Esto iniciará:
- PostgreSQL (con la base de datos `onboarding_db`)
- Migraciones automáticas de onboarding
- IAM Service
- PIM Service
- Onboarding Service
- Kong Gateway

### 4. Verificar que el servicio está corriendo
```bash
# Health check del onboarding service
curl http://localhost:8110/health

# A través de Kong Gateway
curl http://localhost:8001/onboarding/health
```

### 5. Probar el flujo completo

#### Iniciar proceso de onboarding
```bash
curl -X POST http://localhost:8001/onboarding/api/v1/onboarding/start
```

#### Registrar usuario
```bash
curl -X POST http://localhost:8001/onboarding/api/v1/onboarding/register-user \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@test.com",
    "password": "password123",
    "phone": "+5491112345678"
  }'
```

## Troubleshooting

### Docker no inicia
```bash
# Verificar que Docker Desktop está instalado
which docker

# Reiniciar Docker Desktop
pkill Docker && open -a Docker
```

### Puerto 8110 ya en uso
```bash
# Ver qué proceso está usando el puerto
lsof -i :8110

# Matar el proceso si es necesario
kill -9 <PID>
```

### Migraciones fallan
```bash
# Entrar al contenedor de PostgreSQL
docker exec -it mc-postgres psql -U postgres

# Verificar que la BD existe
\l

# Conectar a la BD
\c onboarding_db

# Ver tablas
\dt
```

## Próximos Pasos

Una vez que Docker esté corriendo y los servicios levanten correctamente, continuar con los items del HITO 1:
- Item 3: Probar /start
- Item 4: Probar /register-user
- Item 5: Probar /verify-email
- Item 6: Probar /setup-store
- Item 7: Probar /select-plan
- Item 8: Probar /complete
- Item 9: Probar login
- Item 10: Verificar acceso a Quickstart

