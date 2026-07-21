-- Migration: 006_rls_onboarding (down)
-- Description: Revert RLS + policies + función resolver (orden inverso al up)

DROP FUNCTION IF EXISTS onboarding_tenant_for_process(uuid);

DROP POLICY IF EXISTS tenant_isolation ON verification_codes;
ALTER TABLE verification_codes NO FORCE ROW LEVEL SECURITY;
ALTER TABLE verification_codes DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS tenant_isolation ON onboarding_processes;
ALTER TABLE onboarding_processes NO FORCE ROW LEVEL SECURITY;
ALTER TABLE onboarding_processes DISABLE ROW LEVEL SECURITY;
