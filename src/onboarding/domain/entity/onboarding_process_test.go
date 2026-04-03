package entity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOnboardingProcess_WithValidIDs_ReturnsProcessAtStep1(t *testing.T) {
	// Arrange
	tenantID := uuid.New()
	userID := uuid.New()

	// Act
	process := NewOnboardingProcess(tenantID, userID)

	// Assert
	require.NotNil(t, process)
	assert.NotEqual(t, uuid.Nil, process.ID)
	assert.Equal(t, tenantID, process.TenantID)
	assert.Equal(t, userID, process.UserID)
	assert.Equal(t, 1, process.CurrentStepNumber)
	assert.False(t, process.IsCompleted)
	assert.Empty(t, process.StepsCompleted)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, process.StepsPending)
	assert.Empty(t, process.StepsSkipped)
	assert.Nil(t, process.CompletedAt)
	assert.False(t, process.StartedAt.IsZero())
	assert.False(t, process.CreatedAt.IsZero())
}

func TestAdvanceToStep_FromStep1ToStep2_MarksStep1Completed(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())

	// Act
	process.AdvanceToStep(2)

	// Assert
	assert.Equal(t, 2, process.CurrentStepNumber)
	assert.Contains(t, process.StepsCompleted, 1)
	assert.NotContains(t, process.StepsPending, 1)
}

func TestAdvanceToStep_WhenStepAlreadyCompleted_DoesNotDuplicate(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())
	process.CompleteStep(1)

	// Act
	process.AdvanceToStep(2)

	// Assert
	count := 0
	for _, s := range process.StepsCompleted {
		if s == 1 {
			count++
		}
	}
	assert.Equal(t, 1, count, "step 1 should appear only once in completed")
}

func TestCompleteStep_WithNewStep_AddsToCompleted(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())

	// Act
	process.CompleteStep(1)

	// Assert
	assert.Contains(t, process.StepsCompleted, 1)
	assert.NotContains(t, process.StepsPending, 1)
}

func TestCompleteStep_WithAlreadyCompletedStep_DoesNotDuplicate(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())
	process.CompleteStep(1)

	// Act
	process.CompleteStep(1)

	// Assert
	count := 0
	for _, s := range process.StepsCompleted {
		if s == 1 {
			count++
		}
	}
	assert.Equal(t, 1, count)
}

func TestCompleteOnboarding_WhenInProgress_MarksCompleted(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())

	// Act
	process.CompleteOnboarding()

	// Assert
	assert.True(t, process.IsCompleted)
	require.NotNil(t, process.CompletedAt)
	assert.Empty(t, process.StepsPending)
	assert.Len(t, process.StepsCompleted, 5)
}

func TestSetStoreConfiguration_WithValidData_UpdatesFields(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())
	categories := []string{"cat1", "cat2"}

	// Act
	process.SetStoreConfiguration("Mi Tienda", "retail", "pyme", categories)

	// Assert
	assert.Equal(t, "Mi Tienda", process.CompanyName)
	assert.Equal(t, "retail", process.BusinessType)
	assert.Equal(t, "pyme", process.StoreSize)
	assert.Equal(t, categories, process.SelectedCategories)
}

func TestGetProgress_WithNoStepsCompleted_ReturnsZero(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())

	// Act
	progress := process.GetProgress()

	// Assert
	assert.Equal(t, 0.0, progress)
}

func TestGetProgress_WithAllStepsCompleted_Returns100(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())
	process.CompleteOnboarding()

	// Act
	progress := process.GetProgress()

	// Assert
	assert.Equal(t, 100.0, progress)
}

func TestGetProgress_With2Of5Completed_Returns40(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())
	process.CompleteStep(1)
	process.CompleteStep(2)

	// Act
	progress := process.GetProgress()

	// Assert
	assert.Equal(t, 40.0, progress)
}

func TestIsStepCompleted_WhenCompleted_ReturnsTrue(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())
	process.CompleteStep(1)

	// Act & Assert
	assert.True(t, process.IsStepCompleted(1))
}

func TestIsStepCompleted_WhenNotCompleted_ReturnsFalse(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())

	// Act & Assert
	assert.False(t, process.IsStepCompleted(1))
}

func TestGetNextStep_WithPendingSteps_ReturnsFirst(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())

	// Act
	next := process.GetNextStep()

	// Assert
	assert.Equal(t, 1, next)
}

func TestGetNextStep_WhenCompleted_ReturnsZero(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())
	process.CompleteOnboarding()

	// Act
	next := process.GetNextStep()

	// Assert
	assert.Equal(t, 0, next)
}

func TestGetNextStep_WithNoPendingButNotCompleted_ReturnsCurrentStep(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())
	process.StepsPending = []int{}
	process.CurrentStepNumber = 3

	// Act
	next := process.GetNextStep()

	// Assert
	assert.Equal(t, 3, next)
}

func TestToJSON_WithValidProcess_ReturnsValidJSON(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())

	// Act
	data, err := process.ToJSON()

	// Assert
	require.NoError(t, err)
	assert.NotEmpty(t, data)
	assert.Contains(t, string(data), "id")
	assert.Contains(t, string(data), "tenant_id")
}

func TestIsValid_WithValidProcess_ReturnsTrue(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())

	// Act & Assert
	assert.True(t, process.IsValid())
}

func TestIsValid_WithNilTenantID_ReturnsFalse(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.Nil, uuid.New())

	// Act & Assert
	assert.False(t, process.IsValid())
}

func TestIsValid_WithNilUserID_ReturnsFalse(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.Nil)

	// Act & Assert
	assert.False(t, process.IsValid())
}

func TestIsValid_WithStepOutOfRange_ReturnsFalse(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())
	process.CurrentStepNumber = 6

	// Act & Assert
	assert.False(t, process.IsValid())
}

func TestIsValid_WithStepZero_ReturnsFalse(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())
	process.CurrentStepNumber = 0

	// Act & Assert
	assert.False(t, process.IsValid())
}

func TestHasValidStoreConfig_WithAllFields_ReturnsTrue(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())
	process.SetStoreConfiguration("Tienda", "retail", "micro", nil)

	// Act & Assert
	assert.True(t, process.HasValidStoreConfig())
}

func TestHasValidStoreConfig_WithEmptyCompanyName_ReturnsFalse(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())
	process.BusinessType = "retail"
	process.StoreSize = "micro"

	// Act & Assert
	assert.False(t, process.HasValidStoreConfig())
}

func TestHasValidStoreConfig_WithEmptyBusinessType_ReturnsFalse(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())
	process.CompanyName = "Tienda"
	process.StoreSize = "micro"

	// Act & Assert
	assert.False(t, process.HasValidStoreConfig())
}

func TestHasValidStoreConfig_WithEmptyStoreSize_ReturnsFalse(t *testing.T) {
	// Arrange
	process := NewOnboardingProcess(uuid.New(), uuid.New())
	process.CompanyName = "Tienda"
	process.BusinessType = "retail"

	// Act & Assert
	assert.False(t, process.HasValidStoreConfig())
}
