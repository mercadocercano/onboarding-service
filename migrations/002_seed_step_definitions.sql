-- Migration: 002_seed_step_definitions.sql
-- Description: Seed initial step definitions for onboarding process

-- Insert default step definitions
INSERT INTO onboarding_step_definitions (
    id, step_number, step_name, step_title, description, is_required, has_ui, 
    requires_user_input, can_be_skipped, display_order, is_active, created_at, updated_at
) VALUES 
(
    gen_random_uuid(), 1, 'bienvenida', 'Bienvenida', 
    'Pantalla de bienvenida al onboarding', 
    true, true, false, false, 1, true, NOW(), NOW()
),
(
    gen_random_uuid(), 2, 'registro', 'Datos de Usuario', 
    'Captura nombre, email y contraseña', 
    true, true, true, false, 2, true, NOW(), NOW()
),
(
    gen_random_uuid(), 3, 'verificacion', 'Verificar Email', 
    'Validación del código de verificación', 
    true, true, true, false, 3, true, NOW(), NOW()
),
(
    gen_random_uuid(), 4, 'configuracion_tienda', 'Detalles de la Tienda', 
    'Configuración básica del negocio', 
    true, true, true, false, 4, true, NOW(), NOW()
),
(
    gen_random_uuid(), 5, 'completar', 'Finalizar', 
    'Completar onboarding y redirección', 
    true, true, false, false, 5, true, NOW(), NOW()
)
ON CONFLICT (step_number) DO UPDATE SET
    step_name = EXCLUDED.step_name,
    step_title = EXCLUDED.step_title,
    description = EXCLUDED.description,
    updated_at = NOW(); 