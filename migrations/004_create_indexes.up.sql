-- Migration: 003_create_indexes.sql
-- Description: Create indexes for better performance

-- Indexes for onboarding_processes table
CREATE INDEX IF NOT EXISTS idx_onboarding_processes_tenant_id ON onboarding_processes(tenant_id);
CREATE INDEX IF NOT EXISTS idx_onboarding_processes_user_id ON onboarding_processes(user_id);
CREATE INDEX IF NOT EXISTS idx_onboarding_processes_is_completed ON onboarding_processes(is_completed);
CREATE INDEX IF NOT EXISTS idx_onboarding_processes_current_step ON onboarding_processes(current_step_number);

-- Indexes for onboarding_step_definitions table
CREATE INDEX IF NOT EXISTS idx_onboarding_step_definitions_step_number ON onboarding_step_definitions(step_number);
CREATE INDEX IF NOT EXISTS idx_onboarding_step_definitions_active ON onboarding_step_definitions(is_active); 