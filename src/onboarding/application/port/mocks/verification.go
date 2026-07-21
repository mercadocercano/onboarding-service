package mocks

import "github.com/stretchr/testify/mock"

// MockVerificationCodeGenerator es un mock para el driven port
// VerificationCodeGenerator de application.
type MockVerificationCodeGenerator struct {
	mock.Mock
}

func (m *MockVerificationCodeGenerator) Generate() string {
	args := m.Called()
	return args.String(0)
}
