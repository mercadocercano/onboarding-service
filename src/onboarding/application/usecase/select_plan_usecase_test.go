package usecase

import (
	"context"
	"errors"
	"testing"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/domain/entity"
	"onboarding/src/onboarding/domain/entity/testutil"
	"onboarding/src/onboarding/domain/port/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSelectPlan_WithValidRequest_ReturnsSuccess(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewSelectPlanUseCase(repo)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep5SelectPlan().
		Build()

	repo.On("GetProcessByID", mock.Anything, processID).Return(process, nil)
	repo.On("UpdateProcess", mock.Anything, mock.AnythingOfType("*entity.OnboardingProcess")).Return(nil)

	req := &request.SelectPlanRequest{
		ProcessID:    processID.String(),
		SelectedPlan: "basic",
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "basic", resp.PlanSelected)
	assert.True(t, resp.PlanConfigured)
}

func TestSelectPlan_WithInvalidPlan_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewSelectPlanUseCase(repo)

	req := &request.SelectPlanRequest{
		ProcessID:    uuid.New().String(),
		SelectedPlan: "invalid_plan",
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestSelectPlan_WithEmptyProcessID_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewSelectPlanUseCase(repo)

	req := &request.SelectPlanRequest{
		ProcessID:    "",
		SelectedPlan: "basic",
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestSelectPlan_WithInvalidProcessID_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewSelectPlanUseCase(repo)

	req := &request.SelectPlanRequest{
		ProcessID:    "not-a-uuid",
		SelectedPlan: "basic",
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestSelectPlan_WhenProcessNotFound_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewSelectPlanUseCase(repo)

	processID := uuid.New()
	repo.On("GetProcessByID", mock.Anything, processID).Return((*entity.OnboardingProcess)(nil), errors.New("not found"))

	req := &request.SelectPlanRequest{
		ProcessID:    processID.String(),
		SelectedPlan: "basic",
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestSelectPlan_WhenProcessNotAtStep5_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewSelectPlanUseCase(repo)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		WithCurrentStep(3). // Not at step 5
		Build()

	repo.On("GetProcessByID", mock.Anything, processID).Return(process, nil)

	req := &request.SelectPlanRequest{
		ProcessID:    processID.String(),
		SelectedPlan: "basic",
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestSelectPlan_WhenRepoUpdateFails_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewSelectPlanUseCase(repo)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep5SelectPlan().
		Build()

	repo.On("GetProcessByID", mock.Anything, processID).Return(process, nil)
	repo.On("UpdateProcess", mock.Anything, mock.AnythingOfType("*entity.OnboardingProcess")).Return(errors.New("db error"))

	req := &request.SelectPlanRequest{
		ProcessID:    processID.String(),
		SelectedPlan: "premium",
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestSelectPlan_WithAllValidPlans_Succeeds(t *testing.T) {
	validPlans := []string{"free", "basic", "premium", "enterprise"}

	for _, plan := range validPlans {
		t.Run("plan_"+plan, func(t *testing.T) {
			// Arrange
			repo := new(mocks.MockOnboardingRepository)
			uc := NewSelectPlanUseCase(repo)

			processID := uuid.New()
			process := testutil.NewOnboardingProcessMother().
				WithID(processID).
				AtStep5SelectPlan().
				Build()

			repo.On("GetProcessByID", mock.Anything, processID).Return(process, nil)
			repo.On("UpdateProcess", mock.Anything, mock.AnythingOfType("*entity.OnboardingProcess")).Return(nil)

			req := &request.SelectPlanRequest{
				ProcessID:    processID.String(),
				SelectedPlan: plan,
			}

			// Act
			resp, err := uc.Execute(context.Background(), req)

			// Assert
			require.NoError(t, err)
			assert.True(t, resp.Success)
			assert.Equal(t, plan, resp.PlanSelected)
		})
	}
}
