-- Migration: 003_create_verification_codes.sql
-- Description: Create verification_codes table for email verification

CREATE TABLE IF NOT EXISTS verification_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    process_id UUID NOT NULL,
    email VARCHAR(255) NOT NULL,
    code VARCHAR(6) NOT NULL,
    is_used BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Foreign key constraint to onboarding_processes
    CONSTRAINT fk_verification_codes_process 
        FOREIGN KEY (process_id) 
        REFERENCES onboarding_processes(id) 
        ON DELETE CASCADE
);

-- Índices separados para mejorar rendimiento
CREATE INDEX IF NOT EXISTS idx_verification_codes_process_id ON verification_codes(process_id);
CREATE INDEX IF NOT EXISTS idx_verification_codes_code ON verification_codes(code);
CREATE INDEX IF NOT EXISTS idx_verification_codes_email ON verification_codes(email);
CREATE INDEX IF NOT EXISTS idx_verification_codes_expires_at ON verification_codes(expires_at);

-- Comentarios en la tabla y columnas
COMMENT ON TABLE verification_codes IS 'Códigos de verificación de email para el proceso de onboarding';
COMMENT ON COLUMN verification_codes.id IS 'Identificador único del código de verificación';
COMMENT ON COLUMN verification_codes.process_id IS 'ID del proceso de onboarding asociado';
COMMENT ON COLUMN verification_codes.email IS 'Email del usuario al que se envió el código';
COMMENT ON COLUMN verification_codes.code IS 'Código de verificación de 6 dígitos';
COMMENT ON COLUMN verification_codes.is_used IS 'Indica si el código ya fue utilizado';
COMMENT ON COLUMN verification_codes.expires_at IS 'Fecha y hora de expiración del código';
COMMENT ON COLUMN verification_codes.created_at IS 'Fecha y hora de creación del código';
COMMENT ON COLUMN verification_codes.updated_at IS 'Fecha y hora de última actualización';

-- Trigger para actualizar updated_at automáticamente
CREATE OR REPLACE FUNCTION update_verification_codes_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_verification_codes_updated_at
    BEFORE UPDATE ON verification_codes
    FOR EACH ROW
    EXECUTE FUNCTION update_verification_codes_updated_at(); 