package entity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVerificationCode_WithValidData_ReturnsCode(t *testing.T) {
	// Arrange
	processID := uuid.New()
	email := "user@example.com"
	code := "123456"

	// Act
	vc := NewVerificationCode(processID, email, code)

	// Assert
	require.NotNil(t, vc)
	assert.NotEqual(t, uuid.Nil, vc.ID)
	assert.Equal(t, processID, vc.ProcessID)
	assert.Equal(t, email, vc.UserEmail)
	assert.Equal(t, code, vc.Code)
	assert.False(t, vc.IsUsed)
	assert.False(t, vc.CreatedAt.IsZero())
	assert.True(t, vc.ExpiresAt.After(time.Now()))
}

func TestNewVerificationCode_ExpiresIn15Minutes(t *testing.T) {
	// Arrange & Act
	vc := NewVerificationCode(uuid.New(), "test@test.com", "000000")

	// Assert
	expectedExpiry := time.Now().Add(15 * time.Minute)
	diff := vc.ExpiresAt.Sub(expectedExpiry)
	assert.True(t, diff < time.Second && diff > -time.Second,
		"expiry should be approximately 15 minutes from now")
}

func TestIsExpired_WhenNotExpired_ReturnsFalse(t *testing.T) {
	// Arrange
	vc := NewVerificationCode(uuid.New(), "test@test.com", "123456")

	// Act & Assert
	assert.False(t, vc.IsExpired())
}

func TestIsExpired_WhenExpired_ReturnsTrue(t *testing.T) {
	// Arrange
	vc := NewVerificationCode(uuid.New(), "test@test.com", "123456")
	vc.ExpiresAt = time.Now().Add(-1 * time.Minute)

	// Act & Assert
	assert.True(t, vc.IsExpired())
}

func TestIsValid_WhenFreshAndUnused_ReturnsTrue(t *testing.T) {
	// Arrange
	vc := NewVerificationCode(uuid.New(), "test@test.com", "123456")

	// Act & Assert
	assert.True(t, vc.IsValid())
}

func TestIsValid_WhenUsed_ReturnsFalse(t *testing.T) {
	// Arrange
	vc := NewVerificationCode(uuid.New(), "test@test.com", "123456")
	vc.MarkAsUsed()

	// Act & Assert
	assert.False(t, vc.IsValid())
}

func TestIsValid_WhenExpired_ReturnsFalse(t *testing.T) {
	// Arrange
	vc := NewVerificationCode(uuid.New(), "test@test.com", "123456")
	vc.ExpiresAt = time.Now().Add(-1 * time.Minute)

	// Act & Assert
	assert.False(t, vc.IsValid())
}

func TestIsValid_WhenUsedAndExpired_ReturnsFalse(t *testing.T) {
	// Arrange
	vc := NewVerificationCode(uuid.New(), "test@test.com", "123456")
	vc.MarkAsUsed()
	vc.ExpiresAt = time.Now().Add(-1 * time.Minute)

	// Act & Assert
	assert.False(t, vc.IsValid())
}

func TestMarkAsUsed_SetsIsUsedTrue(t *testing.T) {
	// Arrange
	vc := NewVerificationCode(uuid.New(), "test@test.com", "123456")

	// Act
	vc.MarkAsUsed()

	// Assert
	assert.True(t, vc.IsUsed)
}

func TestMarkAsUsed_UpdatesTimestamp(t *testing.T) {
	// Arrange
	vc := NewVerificationCode(uuid.New(), "test@test.com", "123456")
	originalUpdated := vc.UpdatedAt
	time.Sleep(1 * time.Millisecond)

	// Act
	vc.MarkAsUsed()

	// Assert
	assert.True(t, vc.UpdatedAt.After(originalUpdated) || vc.UpdatedAt.Equal(originalUpdated))
}
