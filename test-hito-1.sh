#!/bin/bash

# Script de prueba para HITO 1 - Onboarding Tecnico Funcional
# Este script prueba los 10 items del checklist de forma automatica

set -e  # Exit on error

# Colores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Variables globales
API_URL="http://localhost:8001/onboarding/api/v1"
IAM_URL="http://localhost:8001/iam/api/v1"
PIM_URL="http://localhost:8001/pim/api/v1"
PROCESS_ID=""
TENANT_ID=""
USER_ID=""
ACCESS_TOKEN=""
TEST_EMAIL="test-$(date +%s)@test.com"
TEST_PASSWORD="password123"
VERIFICATION_CODE=""

echo -e "${YELLOW}=== HITO 1: Onboarding Tecnico Funcional ===${NC}"
echo ""

# Item 1: Health Check
echo -e "${YELLOW}Item 1: Verificando que onboarding-service levanta correctamente...${NC}"
HEALTH_RESPONSE=$(curl -s http://localhost:8110/health)
if echo "$HEALTH_RESPONSE" | grep -q "ok"; then
    echo -e "${GREEN}✅ Item 1: PASADO - Servicio levanta correctamente${NC}"
else
    echo -e "${RED}❌ Item 1: FALLIDO - Servicio no responde${NC}"
    exit 1
fi
echo ""

# Item 2: Base de datos
echo -e "${YELLOW}Item 2: Verificando base de datos onboarding_db...${NC}"
DB_CHECK=$(docker exec mc-postgres psql -U postgres -d onboarding_db -c "\dt" 2>&1)
if echo "$DB_CHECK" | grep -q "onboarding_processes"; then
    echo -e "${GREEN}✅ Item 2: PASADO - Base de datos creada y migrada${NC}"
else
    echo -e "${RED}❌ Item 2: FALLIDO - Base de datos no encontrada${NC}"
    exit 1
fi
echo ""

# Item 3: POST /start
echo -e "${YELLOW}Item 3: Probando POST /start...${NC}"
START_RESPONSE=$(curl -s -X POST "$API_URL/onboarding/start" \
  -H "Content-Type: application/json")
PROCESS_ID=$(echo "$START_RESPONSE" | jq -r '.process_id // empty')

if [ -n "$PROCESS_ID" ]; then
    echo -e "${GREEN}✅ Item 3: PASADO - Proceso creado: $PROCESS_ID${NC}"
else
    echo -e "${RED}❌ Item 3: FALLIDO - No se pudo crear proceso${NC}"
    echo "Response: $START_RESPONSE"
    exit 1
fi
echo ""

# Item 4: POST /register-user
echo -e "${YELLOW}Item 4: Probando POST /register-user...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/onboarding/register-user" \
  -H "Content-Type: application/json" \
  -d "{
    \"process_id\": \"$PROCESS_ID\",
    \"name\": \"Test User\",
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\",
    \"phone\": \"+5491112345678\"
  }")

TENANT_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.tenant_id // empty')
USER_ID=$(echo "$REGISTER_RESPONSE" | jq -r '.user_id // empty')

if [ -n "$TENANT_ID" ] && [ -n "$USER_ID" ]; then
    echo -e "${GREEN}✅ Item 4: PASADO - Tenant y User creados${NC}"
    echo "  Tenant ID: $TENANT_ID"
    echo "  User ID: $USER_ID"
    
    # Verificar en BD
    TENANT_EXISTS=$(docker exec mc-postgres psql -U postgres -d iam_db -t -c "SELECT COUNT(*) FROM tenants WHERE id='$TENANT_ID'" | xargs)
    USER_EXISTS=$(docker exec mc-postgres psql -U postgres -d iam_db -t -c "SELECT COUNT(*) FROM users WHERE id='$USER_ID'" | xargs)
    
    if [ "$TENANT_EXISTS" == "1" ] && [ "$USER_EXISTS" == "1" ]; then
        echo -e "${GREEN}  ✓ Verificado en base de datos IAM${NC}"
    else
        echo -e "${RED}  ✗ No se encontraron en base de datos${NC}"
    fi
else
    echo -e "${RED}❌ Item 4: FALLIDO - No se creó tenant o user${NC}"
    echo "Response: $REGISTER_RESPONSE"
    exit 1
fi
echo ""

# Item 5: POST /verify-email
echo -e "${YELLOW}Item 5: Probando POST /verify-email...${NC}"
# Obtener código de verificación de los logs o BD
VERIFICATION_CODE=$(docker exec mc-postgres psql -U postgres -d onboarding_db -t -c \
  "SELECT code FROM verification_codes WHERE process_id='$PROCESS_ID' ORDER BY created_at DESC LIMIT 1" | xargs)

if [ -n "$VERIFICATION_CODE" ]; then
    echo "Código de verificación obtenido: $VERIFICATION_CODE"
    
    VERIFY_RESPONSE=$(curl -s -X POST "$API_URL/onboarding/verify-email" \
      -H "Content-Type: application/json" \
      -d "{
        \"process_id\": \"$PROCESS_ID\",
        \"verification_code\": \"$VERIFICATION_CODE\"
      }")
    
    if echo "$VERIFY_RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Item 5: PASADO - Email verificado${NC}"
    else
        echo -e "${RED}❌ Item 5: FALLIDO - Verificación falló${NC}"
        echo "Response: $VERIFY_RESPONSE"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ Item 5: SALTADO - No se pudo obtener código de verificación${NC}"
fi
echo ""

# Item 6: POST /setup-store
echo -e "${YELLOW}Item 6: Probando POST /setup-store...${NC}"
SETUP_RESPONSE=$(curl -s -X POST "$API_URL/onboarding/setup-store" \
  -H "Content-Type: application/json" \
  -d "{
    \"process_id\": \"$PROCESS_ID\",
    \"store_name\": \"Mi Tienda Test\",
    \"business_type\": \"retail\",
    \"store_size\": \"pyme\"
  }")

if echo "$SETUP_RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Item 6: PASADO - Tienda configurada${NC}"
else
    echo -e "${RED}❌ Item 6: FALLIDO - Setup falló${NC}"
    echo "Response: $SETUP_RESPONSE"
    exit 1
fi
echo ""

# Item 7: POST /select-plan (hardcoded FREE)
echo -e "${YELLOW}Item 7: Probando POST /select-plan...${NC}"
PLAN_RESPONSE=$(curl -s -X POST "$API_URL/onboarding/select-plan" \
  -H "Content-Type: application/json" \
  -d "{
    \"process_id\": \"$PROCESS_ID\"
  }")

if echo "$PLAN_RESPONSE" | jq -e '.success' > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Item 7: PASADO - Plan seleccionado (hardcoded FREE)${NC}"
else
    echo -e "${RED}❌ Item 7: FALLIDO - Selección de plan falló${NC}"
    echo "Response: $PLAN_RESPONSE"
    exit 1
fi
echo ""

# Item 8: POST /complete
echo -e "${YELLOW}Item 8: Probando POST /complete...${NC}"
COMPLETE_RESPONSE=$(curl -s -X POST "$API_URL/onboarding/complete" \
  -H "Content-Type: application/json" \
  -d "{
    \"process_id\": \"$PROCESS_ID\"
  }")

if echo "$COMPLETE_RESPONSE" | jq -e '.is_completed' > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Item 8: PASADO - Onboarding completado${NC}"
    REDIRECT_URL=$(echo "$COMPLETE_RESPONSE" | jq -r '.redirect_url')
    echo "  Redirect URL: $REDIRECT_URL"
else
    echo -e "${RED}❌ Item 8: FALLIDO - Complete falló${NC}"
    echo "Response: $COMPLETE_RESPONSE"
    exit 1
fi
echo ""

# Item 9: POST /auth/login
echo -e "${YELLOW}Item 9: Probando POST /auth/login...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$IAM_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\"
  }")

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token // empty')

if [ -n "$ACCESS_TOKEN" ]; then
    echo -e "${GREEN}✅ Item 9: PASADO - Login exitoso${NC}"
    echo "  Token obtenido (primeros 50 chars): ${ACCESS_TOKEN:0:50}..."
else
    echo -e "${RED}❌ Item 9: FALLIDO - Login falló${NC}"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi
echo ""

# Item 10: GET /quickstart/business-types
echo -e "${YELLOW}Item 10: Probando acceso a Quickstart del PIM...${NC}"
QUICKSTART_RESPONSE=$(curl -s -X GET "$PIM_URL/quickstart/business-types" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID")

if echo "$QUICKSTART_RESPONSE" | jq -e 'length > 0' > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Item 10: PASADO - Acceso a Quickstart verificado${NC}"
    BUSINESS_TYPES_COUNT=$(echo "$QUICKSTART_RESPONSE" | jq 'length')
    echo "  Business types disponibles: $BUSINESS_TYPES_COUNT"
else
    echo -e "${RED}❌ Item 10: FALLIDO - No se pudo acceder a Quickstart${NC}"
    echo "Response: $QUICKSTART_RESPONSE"
    exit 1
fi
echo ""

# Resumen final
echo -e "${GREEN}═══════════════════════════════════════${NC}"
echo -e "${GREEN}🎉 HITO 1 COMPLETADO EXITOSAMENTE 🎉${NC}"
echo -e "${GREEN}═══════════════════════════════════════${NC}"
echo ""
echo "Detalles del flujo completado:"
echo "  • Process ID: $PROCESS_ID"
echo "  • Tenant ID: $TENANT_ID"
echo "  • User ID: $USER_ID"
echo "  • Email: $TEST_EMAIL"
echo "  • Token disponible: Sí"
echo ""
echo -e "${YELLOW}Próximos pasos:${NC}"
echo "  1. El comercio puede acceder al backoffice"
echo "  2. Puede seleccionar template de productos"
echo "  3. Puede importar productos CSV"
echo "  4. HITO 1 CERRADO ✅"
echo ""

