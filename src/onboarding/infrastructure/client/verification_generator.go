package client

import (
	"onboarding/src/onboarding/application/port"
)

// verificationCodeGenerator adapta GenerateVerificationCode al driven port de
// application. Se inyecta en los use cases que necesiten generar códigos,
// manteniendo la regla de dependencia hexagonal.
type verificationCodeGenerator struct{}

// NewVerificationCodeGenerator expone la función de infraestructura como un
// driven port de application.
func NewVerificationCodeGenerator() port.VerificationCodeGenerator {
	return &verificationCodeGenerator{}
}

func (g *verificationCodeGenerator) Generate() string {
	return GenerateVerificationCode()
}
