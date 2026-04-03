package usecase

import (
	"errors"
	"testing"

	"onboarding/src/onboarding/application/request"
	"onboarding/src/onboarding/domain/port"
	"onboarding/src/onboarding/domain/port/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newValidRegisterRequest() *request.RegisterUserRequest {
	return &request.RegisterUserRequest{
		Name:            "Test User",
		Email:           "test@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}
}

func TestRegisterUser_WithValidRequest_ReturnsSuccess(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewRegisterUserUseCase(repo, iamClient, notifClient)

	roleID := uuid.New().String()
	tenantID := uuid.New().String()
	userID := uuid.New().String()

	iamClient.On("GetRoleByType", "TENANT_ADMIN").Return(&port.RoleResponse{
		ID: roleID, Type: "TENANT_ADMIN",
	}, nil)
	iamClient.On("CreateTenant", mock.AnythingOfType("*port.CreateTenantRequest")).Return(&port.TenantResponse{
		ID: tenantID, Name: "Tienda de Test User", Slug: "testuser", Type: "STARTUP",
	}, nil)
	iamClient.On("CreateUser", mock.AnythingOfType("*port.CreateUserRequest")).Return(&port.UserResponse{
		ID: userID, Email: "test@example.com", Name: "Test User",
	}, nil)
	repo.On("SaveProcess", mock.AnythingOfType("*entity.OnboardingProcess")).Return(nil)
	repo.On("SaveVerificationCode", mock.AnythingOfType("*entity.VerificationCode")).Return(nil)
	notifClient.On("SendEmailVerification", mock.Anything, "test@example.com", "Test User", mock.AnythingOfType("string")).Return(nil)
	iamClient.On("Login", mock.AnythingOfType("*port.LoginRequest")).Return(&port.LoginResponse{
		AccessToken: "jwt-token", RefreshToken: "refresh-token",
	}, nil)

	req := newValidRegisterRequest()

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, tenantID, resp.TenantID)
	assert.Equal(t, userID, resp.UserID)
	assert.Equal(t, 3, resp.NextStep)
	assert.Equal(t, "jwt-token", resp.AccessToken)
	assert.Equal(t, "refresh-token", resp.RefreshToken)
}

func TestRegisterUser_WithEmptyName_ReturnsValidationError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewRegisterUserUseCase(repo, iamClient, notifClient)

	req := newValidRegisterRequest()
	req.Name = ""

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestRegisterUser_WithInvalidEmail_ReturnsValidationError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewRegisterUserUseCase(repo, iamClient, notifClient)

	req := newValidRegisterRequest()
	req.Email = "not-an-email"

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestRegisterUser_WithShortPassword_ReturnsValidationError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewRegisterUserUseCase(repo, iamClient, notifClient)

	req := newValidRegisterRequest()
	req.Password = "short1"
	req.ConfirmPassword = "short1"

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestRegisterUser_WithMismatchedPasswords_ReturnsValidationError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewRegisterUserUseCase(repo, iamClient, notifClient)

	req := newValidRegisterRequest()
	req.ConfirmPassword = "different123"

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestRegisterUser_WithPasswordNoNumbers_ReturnsValidationError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewRegisterUserUseCase(repo, iamClient, notifClient)

	req := newValidRegisterRequest()
	req.Password = "passwordonly"
	req.ConfirmPassword = "passwordonly"

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	assert.False(t, resp.Success)
}

func TestRegisterUser_WhenGetRoleFails_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewRegisterUserUseCase(repo, iamClient, notifClient)

	iamClient.On("GetRoleByType", "TENANT_ADMIN").Return(nil, errors.New("iam error"))

	req := newValidRegisterRequest()

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestRegisterUser_WhenCreateTenantFails_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewRegisterUserUseCase(repo, iamClient, notifClient)

	iamClient.On("GetRoleByType", "TENANT_ADMIN").Return(&port.RoleResponse{
		ID: uuid.New().String(), Type: "TENANT_ADMIN",
	}, nil)
	iamClient.On("CreateTenant", mock.AnythingOfType("*port.CreateTenantRequest")).Return(nil, errors.New("tenant error"))

	req := newValidRegisterRequest()

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestRegisterUser_WhenCreateUserFails_DeletesTenantAndReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewRegisterUserUseCase(repo, iamClient, notifClient)

	tenantID := uuid.New().String()

	iamClient.On("GetRoleByType", "TENANT_ADMIN").Return(&port.RoleResponse{
		ID: uuid.New().String(), Type: "TENANT_ADMIN",
	}, nil)
	iamClient.On("CreateTenant", mock.AnythingOfType("*port.CreateTenantRequest")).Return(&port.TenantResponse{
		ID: tenantID, Name: "Tienda",
	}, nil)
	iamClient.On("CreateUser", mock.AnythingOfType("*port.CreateUserRequest")).Return(nil, errors.New("user error"))
	iamClient.On("DeleteTenant", tenantID).Return(nil)

	req := newValidRegisterRequest()

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
	iamClient.AssertCalled(t, "DeleteTenant", tenantID) // Rollback
}

func TestRegisterUser_WhenLoginFails_StillReturnsSuccess(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	iamClient := new(mocks.MockIAMClient)
	notifClient := new(mocks.MockNotificationClient)
	uc := NewRegisterUserUseCase(repo, iamClient, notifClient)

	iamClient.On("GetRoleByType", "TENANT_ADMIN").Return(&port.RoleResponse{
		ID: uuid.New().String(), Type: "TENANT_ADMIN",
	}, nil)
	iamClient.On("CreateTenant", mock.AnythingOfType("*port.CreateTenantRequest")).Return(&port.TenantResponse{
		ID: uuid.New().String(), Name: "Tienda",
	}, nil)
	iamClient.On("CreateUser", mock.AnythingOfType("*port.CreateUserRequest")).Return(&port.UserResponse{
		ID: uuid.New().String(), Email: "test@example.com", Name: "Test User",
	}, nil)
	repo.On("SaveProcess", mock.AnythingOfType("*entity.OnboardingProcess")).Return(nil)
	repo.On("SaveVerificationCode", mock.AnythingOfType("*entity.VerificationCode")).Return(nil)
	notifClient.On("SendEmailVerification", mock.Anything, "test@example.com", "Test User", mock.AnythingOfType("string")).Return(nil)
	iamClient.On("Login", mock.AnythingOfType("*port.LoginRequest")).Return(nil, errors.New("login error"))

	req := newValidRegisterRequest()

	// Act
	resp, err := uc.Execute(req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Empty(t, resp.AccessToken) // No token because login failed
}
