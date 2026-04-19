package mocks

import (
	"onboarding/src/onboarding/domain/port"

	"github.com/stretchr/testify/mock"
)

// MockPIMClient es un mock del cliente PIM.
type MockPIMClient struct {
	mock.Mock
}

func (m *MockPIMClient) GetBusinessTypes() ([]*port.BusinessType, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*port.BusinessType), args.Error(1)
}

func (m *MockPIMClient) GetBusinessType(businessTypeID string) (*port.BusinessType, error) {
	args := m.Called(businessTypeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.BusinessType), args.Error(1)
}

func (m *MockPIMClient) GetCategoriesByBusinessType(businessType string) ([]*port.Category, error) {
	args := m.Called(businessType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*port.Category), args.Error(1)
}

func (m *MockPIMClient) GetCategories() ([]*port.Category, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*port.Category), args.Error(1)
}

func (m *MockPIMClient) GetCategory(categoryID string) (*port.Category, error) {
	args := m.Called(categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.Category), args.Error(1)
}

func (m *MockPIMClient) ApplyQuickstartTemplate(tenantID string, config *port.QuickstartConfig) (*port.QuickstartResponse, error) {
	args := m.Called(tenantID, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.QuickstartResponse), args.Error(1)
}

func (m *MockPIMClient) GetQuickstartTemplate(businessType string) (*port.QuickstartTemplate, error) {
	args := m.Called(businessType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.QuickstartTemplate), args.Error(1)
}

func (m *MockPIMClient) CreateInitialProducts(tenantID string, config *port.InitialProductsConfig) (*port.ProductSetupResponse, error) {
	args := m.Called(tenantID, config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.ProductSetupResponse), args.Error(1)
}

func (m *MockPIMClient) ImportProductsFromBusinessType(tenantID, businessTypeCode string) (*port.ImportProductsResponse, error) {
	args := m.Called(tenantID, businessTypeCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.ImportProductsResponse), args.Error(1)
}
