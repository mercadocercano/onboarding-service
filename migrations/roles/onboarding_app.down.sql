-- Rollback de migrations/roles/onboarding_app.sql — PLAT-E28 T4 (best-effort, dev local).
-- Revoca los grants y elimina el rol. DROP ROLE falla si el rol es dueño de objetos o tiene
-- privilegios pendientes en otras DBs — best-effort, mismo patrón que customer_app/E27.
REVOKE EXECUTE ON FUNCTION onboarding_tenant_for_process(uuid) FROM onboarding_app;
REVOKE ALL ON schema_migrations           FROM onboarding_app;
REVOKE ALL ON onboarding_step_definitions FROM onboarding_app;
REVOKE ALL ON verification_codes          FROM onboarding_app;
REVOKE ALL ON onboarding_processes        FROM onboarding_app;
REVOKE USAGE ON SCHEMA public             FROM onboarding_app;
REVOKE CONNECT ON DATABASE onboarding_db  FROM onboarding_app;
DROP ROLE IF EXISTS onboarding_app;
