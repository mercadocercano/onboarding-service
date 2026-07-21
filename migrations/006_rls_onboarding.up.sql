-- Migration: 006_rls_onboarding.sql
-- Description: RLS fail-closed en onboarding_processes y verification_codes (enable + force +
-- policy tenant_isolation con USING y WITH CHECK, sin caso global, sin break_glass) + función
-- resolver pre-auth onboarding_tenant_for_process (D1).
-- PLAT-E28 T3 — RLS retrofit RULE-09/RULE-10. El GRANT EXECUTE al rol de aplicación vive en el
-- DDL de roles (migrations/roles/), NO acá: las migraciones numeradas quedan role-agnostic.

ALTER TABLE onboarding_processes ENABLE ROW LEVEL SECURITY;
ALTER TABLE onboarding_processes FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON onboarding_processes
  USING (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

ALTER TABLE verification_codes ENABLE ROW LEVEL SECURITY;
ALTER TABLE verification_codes FORCE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON verification_codes
  USING (tenant_id = current_setting('app.tenant_id')::uuid)
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::uuid);

-- Resolver pre-auth (D1): mapea process_id (capability UUID) → tenant_id. Superficie
-- mínima, NUNCA devuelve filas. onboarding_step_definitions queda SIN RLS a propósito:
-- catálogo global sin datos de tenant (GET /steps es pre-auth).
CREATE OR REPLACE FUNCTION onboarding_tenant_for_process(p_process_id uuid)
RETURNS uuid
LANGUAGE sql STABLE SECURITY DEFINER
SET search_path = ''   -- hardening D1 (sign-off owner 2026-07-19): path vacío; pg_catalog
                       -- sigue implícito para operadores/tipos (`=`, `uuid`), y la tabla se
                       -- califica con schema abajo → cero superficie de shadowing
AS $$ SELECT tenant_id FROM public.onboarding_processes WHERE id = p_process_id $$;

REVOKE ALL ON FUNCTION onboarding_tenant_for_process(uuid) FROM PUBLIC;
