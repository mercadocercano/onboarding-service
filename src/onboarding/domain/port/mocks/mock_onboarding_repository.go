package mocks

import (
	"onboarding/src/onboarding/domain/entity"
	"onboarding/src/onboarding/domain/port"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockOnboardingRepository es un mock del repositorio de onboarding.
type MockOnboardingRepository struct {
	mock.Mock
}

func (m *MockOnboardingRepository) SaveProcess(process *entity.OnboardingProcess) error {
	args := m.Called(process)
	return args.Error(0)
}

func (m *MockOnboardingRepository) GetProcessByID(id uuid.UUID) (*entity.OnboardingProcess, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OnboardingProcess), args.Error(1)
}

func (m *MockOnboardingRepository) GetProcessByTenantID(tenantID uuid.UUID) (*entity.OnboardingProcess, error) {
	args := m.Called(tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OnboardingProcess), args.Error(1)
}

func (m *MockOnboardingRepository) GetProcessByUserID(userID uuid.UUID) (*entity.OnboardingProcess, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OnboardingProcess), args.Error(1)
}

func (m *MockOnboardingRepository) UpdateProcess(process *entity.OnboardingProcess) error {
	args := m.Called(process)
	return args.Error(0)
}

func (m *MockOnboardingRepository) DeleteProcess(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockOnboardingRepository) GetStepDefinitions() ([]*entity.StepDefinition, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.StepDefinition), args.Error(1)
}

func (m *MockOnboardingRepository) GetStepDefinitionByNumber(stepNumber int) (*entity.StepDefinition, error) {
	args := m.Called(stepNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.StepDefinition), args.Error(1)
}

func (m *MockOnboardingRepository) SaveStepDefinition(stepDef *entity.StepDefinition) error {
	args := m.Called(stepDef)
	return args.Error(0)
}

func (m *MockOnboardingRepository) SaveVerificationCode(code *entity.VerificationCode) error {
	args := m.Called(code)
	return args.Error(0)
}

func (m *MockOnboardingRepository) GetVerificationCodeByProcessID(processID uuid.UUID) (*entity.VerificationCode, error) {
	args := m.Called(processID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.VerificationCode), args.Error(1)
}

func (m *MockOnboardingRepository) GetVerificationCodeByCode(code string) (*entity.VerificationCode, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.VerificationCode), args.Error(1)
}

func (m *MockOnboardingRepository) UpdateVerificationCode(code *entity.VerificationCode) error {
	args := m.Called(code)
	return args.Error(0)
}

func (m *MockOnboardingRepository) DeleteExpiredVerificationCodes() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockOnboardingRepository) GetActiveProcesses() ([]*entity.OnboardingProcess, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.OnboardingProcess), args.Error(1)
}

func (m *MockOnboardingRepository) GetCompletedProcesses() ([]*entity.OnboardingProcess, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.OnboardingProcess), args.Error(1)
}

func (m *MockOnboardingRepository) GetProcessesByDateRange(startDate, endDate string) ([]*entity.OnboardingProcess, error) {
	args := m.Called(startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.OnboardingProcess), args.Error(1)
}

func (m *MockOnboardingRepository) GetProcessStats() (*port.ProcessStats, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.ProcessStats), args.Error(1)
}
