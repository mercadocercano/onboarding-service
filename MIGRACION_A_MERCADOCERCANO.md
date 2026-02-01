# Instrucciones para Migración a mercadocercano

## Estado Actual
El repositorio está clonado desde `trinityweb/saas-mt-onboarding-service` y funcionando localmente en el monorepo.

## Migración Oficial (cuando tengas permisos)

```bash
# 1. Crear repo vacío en GitHub
# Ir a https://github.com/orgs/mercadocercano/repositories
# Click "New repository"
# Nombre: saas-mt-onboarding-service

# 2. Actualizar remote
cd /Users/hornosg/MyProjects/saas-mt/services/onboarding-service
git remote set-url origin https://github.com/mercadocercano/saas-mt-onboarding-service.git

# 3. Push al nuevo repo
git push -u origin main

# 4. Actualizar submodulo en monorepo (si aplica)
cd /Users/hornosg/MyProjects/saas-mt
git submodule update --init --recursive
```

## Mientras tanto
El servicio funciona correctamente desde el clon local. La migración a mercadocercano es solo para organización, no afecta funcionalidad.

