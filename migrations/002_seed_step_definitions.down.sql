-- Down: irreversible (seed de datos de definición de pasos). ADR-001 Param 1.
DO $$ BEGIN RAISE EXCEPTION 'Migration 002 is a data seed (irreversible). Restore from backup if a rollback is required.'; END $$;
