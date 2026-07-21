-- Migration: 005_add_tenant_id_verification_codes (down)
-- Description: Revert tenant_id on verification_codes (orden inverso al up)

DROP INDEX IF EXISTS idx_verification_codes_tenant_id;
ALTER TABLE verification_codes DROP COLUMN IF EXISTS tenant_id;
