package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockTenantClient es un mock del cliente de tenant.
type MockTenantClient struct {
	mock.Mock
}

func (m *MockTenantClient) BootstrapTenantConfig(ctx context.Context, tenantID string) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}
