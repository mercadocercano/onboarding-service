package usecase

import (
	"errors"
	"testing"
	"time"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/domain/entity"
	"onboarding/src/onboarding/domain/entity/testutil"
	"onboarding/src/onboarding/domain/port/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestResendVerification_WithValidRequest_ReturnsSuccess(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewResendVerificationUseCase(repo, notifClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep3Verification().
		Build()

	// Existing code created more than 1 minute ago
	oldCode := testutil.NewVerificationCodeMother().
		WithProcessID(processID).
		WithCreatedAt(time.Now().Add(-2 * time.Minute)).
		Build()

	repo.On("GetProcessByID", processID).Return(process, nil)
	repo.On("GetVerificationCodeByProcessID", processID).Return(oldCode, nil)
	repo.On("UpdateVerificationCode", mock.AnythingOfType("*entity.VerificationCode")).Return(nil)
	repo.On("SaveVerificationCode", mock.AnythingOfType("*entity.VerificationCode")).Return(nil)
	notifClient.On("SendEmailVerification", mock.Anything, "test@example.com", "Usuario", mock.AnythingOfType("string")).Return(nil)

	req := &request.ResendVerificationRequest{
		ProcessID: processID.String(),
		Email:     "test@example.com",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.True(t, resp.EmailSent)
	assert.True(t, resp.NewCodeSent)
}

func TestResendVerification_WithEmptyProcessID_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewResendVerificationUseCase(repo, notifClient)

	req := &request.ResendVerificationRequest{
		ProcessID: "",
		Email:     "test@example.com",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestResendVerification_WithEmptyEmail_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewResendVerificationUseCase(repo, notifClient)

	req := &request.ResendVerificationRequest{
		ProcessID: uuid.New().String(),
		Email:     "",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestResendVerification_WithInvalidEmail_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewResendVerificationUseCase(repo, notifClient)

	req := &request.ResendVerificationRequest{
		ProcessID: uuid.New().String(),
		Email:     "not-an-email",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestResendVerification_WhenProcessNotFound_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewResendVerificationUseCase(repo, notifClient)

	processID := uuid.New()
	repo.On("GetProcessByID", processID).Return((*entity.OnboardingProcess)(nil), errors.New("not found"))

	req := &request.ResendVerificationRequest{
		ProcessID: processID.String(),
		Email:     "test@example.com",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestResendVerification_WhenProcessNil_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewResendVerificationUseCase(repo, notifClient)

	processID := uuid.New()
	repo.On("GetProcessByID", processID).Return((*entity.OnboardingProcess)(nil), nil)

	req := &request.ResendVerificationRequest{
		ProcessID: processID.String(),
		Email:     "test@example.com",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestResendVerification_WhenNotAtStep3_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewResendVerificationUseCase(repo, notifClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		WithCurrentStep(4). // Not at step 3
		Build()

	repo.On("GetProcessByID", processID).Return(process, nil)

	req := &request.ResendVerificationRequest{
		ProcessID: processID.String(),
		Email:     "test@example.com",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestResendVerification_WhenThrottled_ReturnsThrottleResponse(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewResendVerificationUseCase(repo, notifClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep3Verification().
		Build()

	// Code created just now (within 1 minute throttle window)
	recentCode := testutil.NewVerificationCodeMother().
		WithProcessID(processID).
		WithCreatedAt(time.Now().Add(-10 * time.Second)).
		Build()

	repo.On("GetProcessByID", processID).Return(process, nil)
	repo.On("GetVerificationCodeByProcessID", processID).Return(recentCode, nil)

	req := &request.ResendVerificationRequest{
		ProcessID: processID.String(),
		Email:     "test@example.com",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Equal(t, "RATE_LIMITED", resp.Error)
	assert.Greater(t, resp.NextResendTime, 0)
}

func TestResendVerification_WhenSaveCodeFails_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewResendVerificationUseCase(repo, notifClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep3Verification().
		Build()

	repo.On("GetProcessByID", processID).Return(process, nil)
	repo.On("GetVerificationCodeByProcessID", processID).Return((*entity.VerificationCode)(nil), nil)
	repo.On("SaveVerificationCode", mock.AnythingOfType("*entity.VerificationCode")).Return(errors.New("db error"))

	req := &request.ResendVerificationRequest{
		ProcessID: processID.String(),
		Email:     "test@example.com",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestResendVerification_WhenNotificationFails_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewResendVerificationUseCase(repo, notifClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep3Verification().
		Build()

	repo.On("GetProcessByID", processID).Return(process, nil)
	repo.On("GetVerificationCodeByProcessID", processID).Return((*entity.VerificationCode)(nil), nil)
	repo.On("SaveVerificationCode", mock.AnythingOfType("*entity.VerificationCode")).Return(nil)
	notifClient.On("SendEmailVerification", mock.Anything, "test@example.com", "Usuario", mock.AnythingOfType("string")).Return(errors.New("email error"))

	req := &request.ResendVerificationRequest{
		ProcessID: processID.String(),
		Email:     "test@example.com",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestResendVerification_InvalidatesOldCode(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewResendVerificationUseCase(repo, notifClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep3Verification().
		Build()

	oldCode := testutil.NewVerificationCodeMother().
		WithProcessID(processID).
		WithCreatedAt(time.Now().Add(-2 * time.Minute)).
		Build()

	repo.On("GetProcessByID", processID).Return(process, nil)
	repo.On("GetVerificationCodeByProcessID", processID).Return(oldCode, nil)
	repo.On("UpdateVerificationCode", mock.AnythingOfType("*entity.VerificationCode")).Return(nil)
	repo.On("SaveVerificationCode", mock.AnythingOfType("*entity.VerificationCode")).Return(nil)
	notifClient.On("SendEmailVerification", mock.Anything, "test@example.com", "Usuario", mock.AnythingOfType("string")).Return(nil)

	req := &request.ResendVerificationRequest{
		ProcessID: processID.String(),
		Email:     "test@example.com",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	repo.AssertCalled(t, "UpdateVerificationCode", mock.AnythingOfType("*entity.VerificationCode"))
}
