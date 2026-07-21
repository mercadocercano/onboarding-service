package mocks

import (
	"context"

	"onboarding/src/onboarding/domain/entity"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockOnboardingRepository es un mock del repositorio de onboarding.
type MockOnboardingRepository struct {
	mock.Mock
}

func (m *MockOnboardingRepository) SaveProcess(ctx context.Context, process *entity.OnboardingProcess) error {
	args := m.Called(ctx, process)
	return args.Error(0)
}

func (m *MockOnboardingRepository) GetProcessByID(ctx context.Context, id uuid.UUID) (*entity.OnboardingProcess, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.OnboardingProcess), args.Error(1)
}

func (m *MockOnboardingRepository) UpdateProcess(ctx context.Context, process *entity.OnboardingProcess) error {
	args := m.Called(ctx, process)
	return args.Error(0)
}

func (m *MockOnboardingRepository) GetStepDefinitions(ctx context.Context) ([]*entity.StepDefinition, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.StepDefinition), args.Error(1)
}

func (m *MockOnboardingRepository) GetStepDefinitionByNumber(ctx context.Context, stepNumber int) (*entity.StepDefinition, error) {
	args := m.Called(ctx, stepNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.StepDefinition), args.Error(1)
}

func (m *MockOnboardingRepository) SaveVerificationCode(ctx context.Context, code *entity.VerificationCode) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}

func (m *MockOnboardingRepository) GetVerificationCodeByProcessID(ctx context.Context, tenantID, processID uuid.UUID) (*entity.VerificationCode, error) {
	args := m.Called(ctx, tenantID, processID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.VerificationCode), args.Error(1)
}

func (m *MockOnboardingRepository) UpdateVerificationCode(ctx context.Context, code *entity.VerificationCode) error {
	args := m.Called(ctx, code)
	return args.Error(0)
}
