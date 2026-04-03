package usecase

import (
	"errors"
	"testing"
	"time"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/domain/entity"
	"onboarding/src/onboarding/domain/entity/testutil"
	"onboarding/src/onboarding/domain/port"
	"onboarding/src/onboarding/domain/port/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCompleteOnboarding_WithValidRequest_ReturnsSuccess(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	notifClient := new(mocks.MockNotificationClient)
	iamClient := new(mocks.MockIAMClient)
	tenantClient := new(mocks.MockTenantClient)
	uc := NewCompleteOnboardingUseCase(repo, notifClient, iamClient, tenantClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep6Complete().
		WithCompanyName("Mi Tienda").
		WithBusinessType("retail").
		WithStoreSize("micro").
		Build()

	userResp := &port.UserResponse{
		ID:    process.UserID.String(),
		Email: "test@test.com",
		Name:  "Test User",
	}

	repo.On("GetProcessByID", processID).Return(process, nil)
	repo.On("UpdateProcess", mock.AnythingOfType("*entity.OnboardingProcess")).Return(nil)
	iamClient.On("GetUser", process.UserID.String()).Return(userResp, nil)
	notifClient.On("SendWelcomeEmail", mock.Anything, "test@test.com", "Test User", "Mi Tienda", "retail").Return(nil)
	tenantClient.On("BootstrapTenantConfig", mock.Anything, process.TenantID.String()).Return(nil)

	req := &request.CompleteOnboardingRequest{
		ProcessID: processID.String(),
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.True(t, resp.Completed)
	assert.Equal(t, processID.String(), resp.ProcessID)
	assert.Equal(t, "Mi Tienda", resp.Summary.CompanyName)
	assert.Equal(t, "retail", resp.Summary.BusinessType)
	assert.NotEmpty(t, resp.BackofficeURL)
	assert.NotEmpty(t, resp.NextSteps)

	// Esperar a que las goroutines async se ejecuten
	time.Sleep(50 * time.Millisecond)
	repo.AssertCalled(t, "UpdateProcess", mock.AnythingOfType("*entity.OnboardingProcess"))
}

func TestCompleteOnboarding_WithEmptyProcessID_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewCompleteOnboardingUseCase(repo, nil, nil, nil)

	req := &request.CompleteOnboardingRequest{
		ProcessID: "",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestCompleteOnboarding_WithInvalidProcessID_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewCompleteOnboardingUseCase(repo, nil, nil, nil)

	req := &request.CompleteOnboardingRequest{
		ProcessID: "not-a-uuid",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestCompleteOnboarding_WhenProcessNotFound_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewCompleteOnboardingUseCase(repo, nil, nil, nil)

	processID := uuid.New()
	repo.On("GetProcessByID", processID).Return((*entity.OnboardingProcess)(nil), errors.New("not found"))

	req := &request.CompleteOnboardingRequest{
		ProcessID: processID.String(),
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestCompleteOnboarding_WhenProcessNotAtStep6_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewCompleteOnboardingUseCase(repo, nil, nil, nil)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		WithCurrentStep(4). // Not at step 6
		Build()

	repo.On("GetProcessByID", processID).Return(process, nil)

	req := &request.CompleteOnboardingRequest{
		ProcessID: processID.String(),
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestCompleteOnboarding_WhenAlreadyCompleted_ReturnsSuccessAndSendsEmail(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	notifClient := new(mocks.MockNotificationClient)
	iamClient := new(mocks.MockIAMClient)
	uc := NewCompleteOnboardingUseCase(repo, notifClient, iamClient, nil)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep6Complete().
		WithCompleted(true).
		WithCompanyName("Tienda Ya").
		WithBusinessType("retail").
		Build()

	userResp := &port.UserResponse{
		ID:    process.UserID.String(),
		Email: "user@test.com",
		Name:  "User",
	}

	repo.On("GetProcessByID", processID).Return(process, nil)
	iamClient.On("GetUser", process.UserID.String()).Return(userResp, nil)
	notifClient.On("SendWelcomeEmail", mock.Anything, "user@test.com", "User", "Tienda Ya", "retail").Return(nil)

	req := &request.CompleteOnboardingRequest{
		ProcessID: processID.String(),
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)

	// Esperar a que la goroutine async se ejecute
	time.Sleep(50 * time.Millisecond)
}

func TestCompleteOnboarding_WhenRepoUpdateFails_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	uc := NewCompleteOnboardingUseCase(repo, nil, nil, nil)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep6Complete().
		Build()

	repo.On("GetProcessByID", processID).Return(process, nil)
	repo.On("UpdateProcess", mock.AnythingOfType("*entity.OnboardingProcess")).Return(errors.New("db error"))

	req := &request.CompleteOnboardingRequest{
		ProcessID: processID.String(),
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestCompleteOnboarding_WithNilTenantClient_DoesNotPanic(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	notifClient := new(mocks.MockNotificationClient)
	iamClient := new(mocks.MockIAMClient)
	uc := NewCompleteOnboardingUseCase(repo, notifClient, iamClient, nil) // nil tenant client

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep6Complete().
		Build()

	userResp := &port.UserResponse{
		ID:    process.UserID.String(),
		Email: "test@test.com",
		Name:  "Test",
	}

	repo.On("GetProcessByID", processID).Return(process, nil)
	repo.On("UpdateProcess", mock.AnythingOfType("*entity.OnboardingProcess")).Return(nil)
	iamClient.On("GetUser", process.UserID.String()).Return(userResp, nil)
	notifClient.On("SendWelcomeEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	req := &request.CompleteOnboardingRequest{
		ProcessID: processID.String(),
	}

	// Act - should not panic
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	time.Sleep(50 * time.Millisecond)
}
