-- Down: irreversible (solo índices de performance; no se revierten). ADR-001 Param 1.
DO $$ BEGIN RAISE EXCEPTION 'Migration 004 (perf indexes) is not reverted. Restore from backup if a rollback is required.'; END $$;
