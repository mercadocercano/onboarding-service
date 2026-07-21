package usecase

import (
	"context"
	"errors"
	"github.com/stretchr/testify/mock"
	"testing"

	"onboarding/src/onboarding/domain/entity/testutil"
	"onboarding/src/onboarding/domain/port/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetProcessStatus_WithValidID_ReturnsStatus(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewGetProcessStatusUseCase(repo)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		WithCurrentStep(3).
		WithStepsCompleted([]int{1, 2}).
		Build()

	repo.On("GetProcessByID", mock.Anything, processID).Return(process, nil)

	// Act
	resp, err := uc.Execute(context.Background(), processID.String())

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, processID.String(), resp.ProcessID)
	assert.Equal(t, "in_progress", resp.Status)
	assert.Equal(t, 3, resp.CurrentStep)
	assert.Equal(t, 40.0, resp.ProgressPercent)
}

func TestGetProcessStatus_WhenCompleted_ReturnsCompleted(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewGetProcessStatusUseCase(repo)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		WithCompleted(true).
		WithStepsCompleted([]int{1, 2, 3, 4, 5}).
		Build()

	repo.On("GetProcessByID", mock.Anything, processID).Return(process, nil)

	// Act
	resp, err := uc.Execute(context.Background(), processID.String())

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "completed", resp.Status)
	assert.Equal(t, 100.0, resp.ProgressPercent)
}

func TestGetProcessStatus_WithInvalidID_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewGetProcessStatusUseCase(repo)

	// Act
	resp, err := uc.Execute(context.Background(), "invalid-uuid")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestGetProcessStatus_WhenRepoFails_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewGetProcessStatusUseCase(repo)

	processID := uuid.New()
	repo.On("GetProcessByID", mock.Anything, processID).Return(nil, errors.New("db error"))

	// Act
	resp, err := uc.Execute(context.Background(), processID.String())

	// Assert
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}
