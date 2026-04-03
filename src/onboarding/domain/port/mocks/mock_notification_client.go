package mocks

import (
	"context"

	"onboarding/src/onboarding/domain/port"

	"github.com/stretchr/testify/mock"
)

// MockNotificationClient es un mock del cliente de notificaciones.
type MockNotificationClient struct {
	mock.Mock
}

func (m *MockNotificationClient) SendEmailVerification(ctx context.Context, email, name, verificationCode string) error {
	args := m.Called(ctx, email, name, verificationCode)
	return args.Error(0)
}

func (m *MockNotificationClient) SendWelcomeEmail(ctx context.Context, email, userName, companyName, businessType string) error {
	args := m.Called(ctx, email, userName, companyName, businessType)
	return args.Error(0)
}

func (m *MockNotificationClient) SendNotification(ctx context.Context, req *port.NotificationRequest) (*port.NotificationResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*port.NotificationResponse), args.Error(1)
}
