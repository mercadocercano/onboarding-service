package usecase

import (
	"strings"
	"testing"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/domain/port/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartOnboarding_WithValidRequest_ReturnsSuccess(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewStartOnboardingUseCase(repo)
	req := &request.StartOnboardingRequest{
		Source: "landing_page",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.ProcessID)
	assert.Equal(t, 1, resp.CurrentStep)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, resp.StepsPending)
}

func TestStartOnboarding_WithEmptySource_SetsDefault(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewStartOnboardingUseCase(repo)
	req := &request.StartOnboardingRequest{}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
}

func TestStartOnboarding_WithSourceTooLong_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewStartOnboardingUseCase(repo)
	req := &request.StartOnboardingRequest{
		Source: strings.Repeat("a", 101),
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestStartOnboarding_WithUTMTooLong_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewStartOnboardingUseCase(repo)
	req := &request.StartOnboardingRequest{
		UTMCampaign: strings.Repeat("x", 201),
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestStartOnboarding_WithReferrerTooLong_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewStartOnboardingUseCase(repo)
	req := &request.StartOnboardingRequest{
		Referrer: strings.Repeat("r", 501),
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}
