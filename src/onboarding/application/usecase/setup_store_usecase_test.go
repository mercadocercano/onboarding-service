package usecase

import (
	"context"
	"errors"
	"testing"

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

func TestSetupStore_WithValidRequest_ReturnsSuccess(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	pimClient := new(mocks.MockPIMClient)
	iamClient := new(mocks.MockIAMClient)
	uc := NewSetupStoreUseCase(repo, pimClient, iamClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		AtStep3Verification().
		Build()

	businessTypes := []*port.BusinessType{
		{ID: "retail", Code: "retail", Name: "Retail", Description: "Retail store", Icon: "store"},
	}

	repo.On("GetProcessByID", mock.Anything, processID).Return(process, nil)
	repo.On("UpdateProcess", mock.Anything, mock.AnythingOfType("*entity.OnboardingProcess")).Return(nil)
	pimClient.On("GetBusinessTypes").Return(businessTypes, nil)
	pimClient.On("GetBusinessType", "retail").Return(businessTypes[0], nil)
	pimClient.On("GetCategoriesByBusinessType", "retail").Return([]*port.Category{}, nil)
	pimClient.On("ApplyQuickstartTemplate", mock.AnythingOfType("string"), mock.AnythingOfType("*port.QuickstartConfig")).Return(&port.QuickstartResponse{Success: true}, nil)
	pimClient.On("ImportProductsFromBusinessType", mock.AnythingOfType("string"), "retail").Return(&port.ImportProductsResponse{}, nil).Maybe()
	iamClient.On("UpdateTenant", mock.AnythingOfType("string"), mock.AnythingOfType("*port.UpdateTenantRequest")).Return(&port.TenantResponse{}, nil)

	req := &request.SetupStoreRequest{
		ProcessID:    processID.String(),
		StoreName:    "Mi Tienda Test",
		BusinessType: "retail",
		StoreSize:    "micro",
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, processID.String(), resp.ProcessID)
	assert.Equal(t, 5, resp.NextStep)
	assert.Equal(t, "Mi Tienda Test", resp.StoreData.Name)
}

func TestSetupStore_WithInvalidBusinessType_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	pimClient := new(mocks.MockPIMClient)
	iamClient := new(mocks.MockIAMClient)
	uc := NewSetupStoreUseCase(repo, pimClient, iamClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		Build()

	businessTypes := []*port.BusinessType{
		{ID: "retail", Code: "retail", Name: "Retail"},
	}

	repo.On("GetProcessByID", mock.Anything, processID).Return(process, nil)
	pimClient.On("GetBusinessTypes").Return(businessTypes, nil)

	req := &request.SetupStoreRequest{
		ProcessID:    processID.String(),
		StoreName:    "Mi Tienda",
		BusinessType: "inexistente",
		StoreSize:    "micro",
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestSetupStore_WithEmptyStoreName_ReturnsValidationError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	pimClient := new(mocks.MockPIMClient)
	iamClient := new(mocks.MockIAMClient)
	uc := NewSetupStoreUseCase(repo, pimClient, iamClient)

	req := &request.SetupStoreRequest{
		ProcessID:    uuid.New().String(),
		StoreName:    "",
		BusinessType: "retail",
		StoreSize:    "micro",
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestSetupStore_WithInvalidStoreSize_ReturnsValidationError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	pimClient := new(mocks.MockPIMClient)
	iamClient := new(mocks.MockIAMClient)
	uc := NewSetupStoreUseCase(repo, pimClient, iamClient)

	req := &request.SetupStoreRequest{
		ProcessID:    uuid.New().String(),
		StoreName:    "Mi Tienda",
		BusinessType: "retail",
		StoreSize:    "invalid_size",
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestSetupStore_WhenProcessNotFound_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	pimClient := new(mocks.MockPIMClient)
	iamClient := new(mocks.MockIAMClient)
	uc := NewSetupStoreUseCase(repo, pimClient, iamClient)

	processID := uuid.New()
	repo.On("GetProcessByID", mock.Anything, processID).Return((*entity.OnboardingProcess)(nil), errors.New("not found"))

	req := &request.SetupStoreRequest{
		ProcessID:    processID.String(),
		StoreName:    "Mi Tienda",
		BusinessType: "retail",
		StoreSize:    "micro",
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestSetupStore_WhenPIMGetBusinessTypesFails_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	pimClient := new(mocks.MockPIMClient)
	iamClient := new(mocks.MockIAMClient)
	uc := NewSetupStoreUseCase(repo, pimClient, iamClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		Build()

	repo.On("GetProcessByID", mock.Anything, processID).Return(process, nil)
	pimClient.On("GetBusinessTypes").Return(nil, errors.New("pim error"))

	req := &request.SetupStoreRequest{
		ProcessID:    processID.String(),
		StoreName:    "Mi Tienda",
		BusinessType: "retail",
		StoreSize:    "micro",
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestSetupStore_WhenRepoUpdateFails_ReturnsError(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	pimClient := new(mocks.MockPIMClient)
	iamClient := new(mocks.MockIAMClient)
	uc := NewSetupStoreUseCase(repo, pimClient, iamClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		Build()

	businessTypes := []*port.BusinessType{
		{ID: "retail", Code: "retail", Name: "Retail"},
	}

	repo.On("GetProcessByID", mock.Anything, processID).Return(process, nil)
	repo.On("UpdateProcess", mock.Anything, mock.AnythingOfType("*entity.OnboardingProcess")).Return(errors.New("db error"))
	pimClient.On("GetBusinessTypes").Return(businessTypes, nil)

	req := &request.SetupStoreRequest{
		ProcessID:    processID.String(),
		StoreName:    "Mi Tienda",
		BusinessType: "retail",
		StoreSize:    "micro",
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.Error(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
}

func TestSetupStore_WithCategories_IncludesThemInResponse(t *testing.T) {
	// Arrange
	repo := new(mocks.MockOnboardingRepository)
	pimClient := new(mocks.MockPIMClient)
	iamClient := new(mocks.MockIAMClient)
	uc := NewSetupStoreUseCase(repo, pimClient, iamClient)

	processID := uuid.New()
	process := testutil.NewOnboardingProcessMother().
		WithID(processID).
		Build()

	businessTypes := []*port.BusinessType{
		{ID: "retail", Code: "retail", Name: "Retail", Icon: "store"},
	}

	repo.On("GetProcessByID", mock.Anything, processID).Return(process, nil)
	repo.On("UpdateProcess", mock.Anything, mock.AnythingOfType("*entity.OnboardingProcess")).Return(nil)
	pimClient.On("GetBusinessTypes").Return(businessTypes, nil)
	pimClient.On("GetBusinessType", "retail").Return(businessTypes[0], nil)
	pimClient.On("GetCategoriesByBusinessType", "retail").Return([]*port.Category{}, nil)
	pimClient.On("ApplyQuickstartTemplate", mock.AnythingOfType("string"), mock.AnythingOfType("*port.QuickstartConfig")).Return(&port.QuickstartResponse{Success: true}, nil)
	pimClient.On("ImportProductsFromBusinessType", mock.AnythingOfType("string"), "retail").Return(&port.ImportProductsResponse{}, nil).Maybe()
	iamClient.On("UpdateTenant", mock.AnythingOfType("string"), mock.AnythingOfType("*port.UpdateTenantRequest")).Return(&port.TenantResponse{}, nil)

	req := &request.SetupStoreRequest{
		ProcessID:          processID.String(),
		StoreName:          "Mi Tienda",
		BusinessType:       "retail",
		StoreSize:          "pyme",
		SelectedCategories: []string{"cat1", "cat2"},
	}

	// Act
	resp, err := uc.Execute(context.Background(), req)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.Equal(t, "premium", resp.RecommendedPlan)
}
