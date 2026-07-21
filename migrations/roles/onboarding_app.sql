-- migrations/roles/onboarding_app.sql — PLAT-E28 T4 (RULE-09/RULE-10: rol de aplicación least-privilege, NOBYPASSRLS)
--
-- onboarding-service corre hoy en runtime como `postgres` (superuser → BYPASSRLS: FORCE ROW LEVEL
-- SECURITY no aplica a superusers, así que la RLS de la migración 006 quedaría "activa" pero nunca
-- ejercida). Este rol least-privilege lo reemplaza en runtime (el flip del compose se DIFIERE a T7
-- — flipear antes de T6 rompe toda lectura, el repo aún no fija el GUC):
--   • SELECT/INSERT/UPDATE sobre onboarding_processes y verification_codes — SIN DELETE: los
--     únicos métodos DELETE del port (DeleteProcess, DeleteExpiredVerificationCodes) son código
--     muerto que T5 poda (D2). Si el purge de códigos expirados vuelve algún día, gana DELETE ON
--     verification_codes acá con un caller real con scope (ver plan PLAT-E28).
--   • SELECT sobre onboarding_step_definitions: catálogo GLOBAL read-only, sin RLS por diseño
--     (GET /steps es pre-auth). SIN INSERT: el seed canónico es la migración 002; el seed in-app
--     del constructor se poda en T5 (must-fix A2 de la revisión L4 2026-07-18).
--   • SELECT sobre schema_migrations (el runtime solo chequea la versión al arrancar, no aplica DDL).
--   • EXECUTE sobre onboarding_tenant_for_process(uuid): resolver pre-auth D1 (SECURITY DEFINER,
--     revocada de PUBLIC en la migración 006) — único camino para mapear process_id → tenant_id
--     en el wizard pre-auth. Sin este grant el rol no tiene acceso alguno al lookup.
--   • NOBYPASSRLS: las policies tenant_isolation de la migración 006 se ejercen de verdad.
--
-- NO es una migración numerada del boot: vive en migrations/roles/ (fuera del alcance de
-- `//go:embed migrations/*.up.sql migrations/*.down.sql` en migrations.go — glob NO recursivo →
-- este archivo no entra al FS embebido) para que NO se auto-aplique en el arranque bajo
-- onboarding_app (NOCREATEROLE no podría crear roles). Se aplica UNA sola vez, out-of-band,
-- contra el cluster por un admin (postgres). Idempotente.
--
-- Seguridad (mismo patrón que customer_app/E27, sales_app/E25, payment_method_app/E26): este
-- script NO fija password —un literal quedaría en el git history para siempre—. El rol se crea
-- con LOGIN pero SIN password (login efectivo deshabilitado hasta setearla); la password real se
-- fija out-of-band vía `ALTER ROLE onboarding_app PASSWORD '...'` corrido a mano contra
-- lab-postgres, nunca versionada (dev local: onboarding_app123, ver .env.example /
-- docker-compose ONBOARDING_DB_PASSWORD).

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'onboarding_app') THEN
    CREATE ROLE onboarding_app LOGIN
      NOSUPERUSER NOCREATEDB NOCREATEROLE NOBYPASSRLS;
  END IF;
END
$$;

GRANT CONNECT ON DATABASE onboarding_db TO onboarding_app;
GRANT USAGE ON SCHEMA public TO onboarding_app;

-- Hardening least-privilege (B1, serie E27): revocar privilegios de tabla que PUBLIC pudiera
-- conceder implícitamente; el rol solo obtiene lo que se le concede explícito abajo.
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;

-- Superficie VIVA post-poda T5 (D2): sin DELETE en ninguna tabla.
GRANT SELECT, INSERT, UPDATE ON onboarding_processes TO onboarding_app;
GRANT SELECT, INSERT, UPDATE ON verification_codes TO onboarding_app;

-- Catálogo global read-only (seed canónico = migración 002; sin seed in-app post-T5, A2).
GRANT SELECT ON onboarding_step_definitions TO onboarding_app;

-- El runtime solo chequea la versión al arrancar (no aplica DDL nuevo → SELECT alcanza).
GRANT SELECT ON schema_migrations TO onboarding_app;

-- Resolver pre-auth D1 (migración 006; REVOKE ALL ... FROM PUBLIC ya aplicado ahí).
GRANT EXECUTE ON FUNCTION onboarding_tenant_for_process(uuid) TO onboarding_app;
