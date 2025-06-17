-- Migration: 001_initial_schema.sql
-- Description: Create initial schema for onboarding service

-- Create onboarding_step_definitions table
CREATE TABLE IF NOT EXISTS onboarding_step_definitions (
    id UUID PRIMARY KEY,
    step_number INTEGER UNIQUE NOT NULL,
    step_name VARCHAR(100) NOT NULL,
    step_title VARCHAR(200) NOT NULL,
    description TEXT,
    is_required BOOLEAN DEFAULT true,
    has_ui BOOLEAN DEFAULT true,
    requires_user_input BOOLEAN DEFAULT true,
    can_be_skipped BOOLEAN DEFAULT false,
    display_order INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create onboarding_processes table
CREATE TABLE IF NOT EXISTS onboarding_processes (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    current_step_number INTEGER DEFAULT 1,
    is_completed BOOLEAN DEFAULT false,
    company_name VARCHAR(255),
    business_type VARCHAR(100),
    store_size VARCHAR(50),
    steps_completed JSONB DEFAULT '[]'::jsonb,
    steps_pending JSONB DEFAULT '[1,2,3,4,5]'::jsonb,
    steps_skipped JSONB DEFAULT '[]'::jsonb,
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Note: Indexes are created in migration 003_create_indexes.sql 