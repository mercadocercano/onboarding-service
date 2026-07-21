-- Migration: 005_add_tenant_id_verification_codes.sql
-- Description: Add tenant_id to verification_codes (backfill from onboarding_processes via FK)
-- PLAT-E28 T2 (D3) — RLS retrofit RULE-09/RULE-10: verification_codes necesita tenant_id propio
-- para una policy RLS directa (scope por join en policy = costo por fila + acoplamiento).

ALTER TABLE verification_codes ADD COLUMN tenant_id UUID;

UPDATE verification_codes vc
   SET tenant_id = op.tenant_id
  FROM onboarding_processes op
 WHERE vc.process_id = op.id;

ALTER TABLE verification_codes ALTER COLUMN tenant_id SET NOT NULL;
CREATE INDEX IF NOT EXISTS idx_verification_codes_tenant_id ON verification_codes(tenant_id);

COMMENT ON COLUMN verification_codes.tenant_id IS 'Tenant dueño del código (copiado del proceso de onboarding) — scope de la policy RLS tenant_isolation';
