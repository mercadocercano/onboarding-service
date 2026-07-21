package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStepDefinition_WithValidData_ReturnsStep(t *testing.T) {
	// Arrange & Act
	step := NewStepDefinition(1, "bienvenida", "Bienvenida", "Pantalla de bienvenida")

	// Assert
	require.NotNil(t, step)
	assert.NotEqual(t, uuid.Nil, step.ID)
	assert.Equal(t, 1, step.StepNumber)
	assert.Equal(t, "bienvenida", step.StepName)
	assert.Equal(t, "Bienvenida", step.StepTitle)
	assert.Equal(t, "Pantalla de bienvenida", step.Description)
	assert.True(t, step.IsRequired)
	assert.True(t, step.HasUI)
	assert.True(t, step.RequiresUserInput)
	assert.False(t, step.CanBeSkipped)
	assert.Equal(t, 1, step.DisplayOrder)
	assert.True(t, step.IsActive)
}

func TestIsValid_WithValidStep_ReturnsTrue(t *testing.T) {
	// Arrange
	step := NewStepDefinition(1, "test", "Test", "Test step")

	// Act & Assert
	assert.True(t, step.IsValid())
}

func TestIsValid_WithZeroStepNumber_ReturnsFalse(t *testing.T) {
	// Arrange
	step := NewStepDefinition(0, "test", "Test", "Test step")
	// NewStepDefinition sets DisplayOrder = stepNumber = 0

	// Act & Assert
	assert.False(t, step.IsValid())
}

func TestIsValid_WithEmptyStepName_ReturnsFalse(t *testing.T) {
	// Arrange
	step := NewStepDefinition(1, "", "Test", "Test step")

	// Act & Assert
	assert.False(t, step.IsValid())
}

func TestIsValid_WithEmptyStepTitle_ReturnsFalse(t *testing.T) {
	// Arrange
	step := NewStepDefinition(1, "test", "", "Test step")

	// Act & Assert
	assert.False(t, step.IsValid())
}

func TestIsAutomaticStep_WhenNoUserInput_ReturnsTrue(t *testing.T) {
	// Arrange
	step := NewStepDefinition(1, "test", "Test", "Test")
	step.RequiresUserInput = false

	// Act & Assert
	assert.True(t, step.IsAutomaticStep())
}

func TestIsAutomaticStep_WhenRequiresInput_ReturnsFalse(t *testing.T) {
	// Arrange
	step := NewStepDefinition(1, "test", "Test", "Test")

	// Act & Assert
	assert.False(t, step.IsAutomaticStep())
}

func TestIsInteractiveStep_WhenRequiresInputAndHasUI_ReturnsTrue(t *testing.T) {
	// Arrange
	step := NewStepDefinition(1, "test", "Test", "Test")

	// Act & Assert
	assert.True(t, step.IsInteractiveStep())
}

func TestIsInteractiveStep_WhenNoUI_ReturnsFalse(t *testing.T) {
	// Arrange
	step := NewStepDefinition(1, "test", "Test", "Test")
	step.HasUI = false

	// Act & Assert
	assert.False(t, step.IsInteractiveStep())
}

func TestIsInteractiveStep_WhenNoUserInput_ReturnsFalse(t *testing.T) {
	// Arrange
	step := NewStepDefinition(1, "test", "Test", "Test")
	step.RequiresUserInput = false

	// Act & Assert
	assert.False(t, step.IsInteractiveStep())
}
