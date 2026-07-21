package port

// VerificationCodeGenerator define la operación de generación de códigos de
// verificación. Es un driven port (output) declarado por application, no un
// detalle de infraestructura. Esto evita que application importe el paquete
// `infrastructure/client` solo para generar códigos.
type VerificationCodeGenerator interface {
	Generate() string
}
