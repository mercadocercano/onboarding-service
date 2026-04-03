package usecase

import (
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

func TestVerifyEmail_WithValidCode_ReturnsSuccess(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	uc := NewVerifyEmailUseCase(repo, iamClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep3Verification().
		Build()

	code := testutil.NewVerificationCodeMother().
		WithProcessID(processID).
		WithCode("123456").
		Build()

	repo.On("GetProcessByID", processID).Return(process, nil)
	repo.On("GetVerificationCodeByProcessID", processID).Return(code, nil)
	repo.On("UpdateVerificationCode", mock.AnythingOfType("*entity.VerificationCode")).Return(nil)
	repo.On("UpdateProcess", mock.AnythingOfType("*entity.OnboardingProcess")).Return(nil)

	req := &request.VerifyEmailRequest{
		ProcessID:        processID.String(),
		VerificationCode: "123456",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.True(t, resp.Verified)
	repo.AssertCalled(t, "UpdateProcess", mock.AnythingOfType("*entity.OnboardingProcess"))
}

func TestVerifyEmail_WithInvalidCode_ReturnsInvalidCodeResponse(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	uc := NewVerifyEmailUseCase(repo, iamClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep3Verification().
		Build()

	code := testutil.NewVerificationCodeMother().
		WithProcessID(processID).
		WithCode("654321").
		Build()

	repo.On("GetProcessByID", processID).Return(process, nil)
	repo.On("GetVerificationCodeByProcessID", processID).Return(code, nil)

	req := &request.VerifyEmailRequest{
		ProcessID:        processID.String(),
		VerificationCode: "123456",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.False(t, resp.Verified)
}

func TestVerifyEmail_WithExpiredCode_ReturnsFalse(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	uc := NewVerifyEmailUseCase(repo, iamClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep3Verification().
		Build()

	code := testutil.NewVerificationCodeMother().
		WithProcessID(processID).
		WithCode("123456").
		WithExpired().
		Build()

	repo.On("GetProcessByID", processID).Return(process, nil)
	repo.On("GetVerificationCodeByProcessID", processID).Return(code, nil)

	req := &request.VerifyEmailRequest{
		ProcessID:        processID.String(),
		VerificationCode: "123456",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.False(t, resp.Verified)
}

func TestVerifyEmail_WithUsedCode_ReturnsFalse(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	uc := NewVerifyEmailUseCase(repo, iamClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep3Verification().
		Build()

	code := testutil.NewVerificationCodeMother().
		WithProcessID(processID).
		WithCode("123456").
		WithUsed(true).
		Build()

	repo.On("GetProcessByID", processID).Return(process, nil)
	repo.On("GetVerificationCodeByProcessID", processID).Return(code, nil)

	req := &request.VerifyEmailRequest{
		ProcessID:        processID.String(),
		VerificationCode: "123456",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestVerifyEmail_WithEmptyProcessID_ReturnsValidationError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	uc := NewVerifyEmailUseCase(repo, iamClient)

	req := &request.VerifyEmailRequest{
		ProcessID:        "",
		VerificationCode: "123456",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestVerifyEmail_WithInvalidProcessID_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	uc := NewVerifyEmailUseCase(repo, iamClient)

	req := &request.VerifyEmailRequest{
		ProcessID:        "not-a-uuid",
		VerificationCode: "123456",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestVerifyEmail_WhenProcessNotFound_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	uc := NewVerifyEmailUseCase(repo, iamClient)

	processID := uuid.New()
	repo.On("GetProcessByID", processID).Return((*entity.OnboardingProcess)(nil), errors.New("not found"))

	req := &request.VerifyEmailRequest{
		ProcessID:        processID.String(),
		VerificationCode: "123456",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestVerifyEmail_WhenRepoUpdateFails_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	uc := NewVerifyEmailUseCase(repo, iamClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep3Verification().
		Build()

	code := testutil.NewVerificationCodeMother().
		WithProcessID(processID).
		WithCode("123456").
		Build()

	repo.On("GetProcessByID", processID).Return(process, nil)
	repo.On("GetVerificationCodeByProcessID", processID).Return(code, nil)
	repo.On("UpdateVerificationCode", mock.AnythingOfType("*entity.VerificationCode")).Return(nil)
	repo.On("UpdateProcess", mock.AnythingOfType("*entity.OnboardingProcess")).Return(errors.New("db error"))

	req := &request.VerifyEmailRequest{
		ProcessID:        processID.String(),
		VerificationCode: "123456",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestVerifyEmail_WhenNoCodeInDB_Accepts6DigitCode(t *testing.T) {
	// Arrange - development mode fallback
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	uc := NewVerifyEmailUseCase(repo, iamClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep3Verification().
		Build()

	repo.On("GetProcessByID", processID).Return(process, nil)
	repo.On("GetVerificationCodeByProcessID", processID).Return((*entity.VerificationCode)(nil), nil)
	repo.On("UpdateProcess", mock.AnythingOfType("*entity.OnboardingProcess")).Return(nil)

	req := &request.VerifyEmailRequest{
		ProcessID:        processID.String(),
		VerificationCode: "999999",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
}

func TestVerifyEmail_WithEmptyCode_ReturnsValidationError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	uc := NewVerifyEmailUseCase(repo, iamClient)

	req := &request.VerifyEmailRequest{
		ProcessID:        uuid.New().String(),
		VerificationCode: "",
	}

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}
