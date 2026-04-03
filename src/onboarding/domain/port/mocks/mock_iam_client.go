package mocks

import (
	"onboarding/src/onboarding/domain/port"

	"github.com/stretchr/testify/mock"
)

// MockIAMClient es un mock del cliente IAM.
type MockIAMClient struct {
	mock.Mock
}

func (m *MockIAMClient) CreateTenant(request *port.CreateTenantRequest) (*port.TenantResponse, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.TenantResponse), args.Error(1)
}

func (m *MockIAMClient) GetTenant(tenantID string) (*port.TenantResponse, error) {
	args := m.Called(tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.TenantResponse), args.Error(1)
}

func (m *MockIAMClient) UpdateTenant(tenantID string, request *port.UpdateTenantRequest) (*port.TenantResponse, error) {
	args := m.Called(tenantID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.TenantResponse), args.Error(1)
}

func (m *MockIAMClient) UpdateTenantOwner(tenantID, userID string) error {
	args := m.Called(tenantID, userID)
	return args.Error(0)
}

func (m *MockIAMClient) DeleteTenant(tenantID string) error {
	args := m.Called(tenantID)
	return args.Error(0)
}

func (m *MockIAMClient) CreateUser(request *port.CreateUserRequest) (*port.UserResponse, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.UserResponse), args.Error(1)
}

func (m *MockIAMClient) GetUser(userID string) (*port.UserResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.UserResponse), args.Error(1)
}

func (m *MockIAMClient) UpdateUser(userID string, request *port.UpdateUserRequest) (*port.UserResponse, error) {
	args := m.Called(userID, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.UserResponse), args.Error(1)
}

func (m *MockIAMClient) DeleteUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockIAMClient) GetRoleByType(roleType string) (*port.RoleResponse, error) {
	args := m.Called(roleType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.RoleResponse), args.Error(1)
}

func (m *MockIAMClient) GetRoles() ([]*port.RoleResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*port.RoleResponse), args.Error(1)
}

func (m *MockIAMClient) Login(request *port.LoginRequest) (*port.LoginResponse, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.LoginResponse), args.Error(1)
}

func (m *MockIAMClient) ValidateToken(token string) (*port.TokenValidationResponse, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.TokenValidationResponse), args.Error(1)
}
