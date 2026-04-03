package testutil

import (
	"time"

	"onboarding/src/onboarding/domain/entity"

	"github.com/google/uuid"
)

// OnboardingProcessMother construye instancias de OnboardingProcess para tests.
type OnboardingProcessMother struct {
	process *entity.OnboardingProcess
}

// NewOnboardingProcessMother crea un builder con valores por defecto válidos.
func NewOnboardingProcessMother() *OnboardingProcessMother {
	now := time.Now()
	return &OnboardingProcessMother{
		process: &entity.OnboardingProcess{
			ID:                uuid.New(),
			TenantID:          uuid.New(),
			UserID:            uuid.New(),
			CurrentStepNumber: 1,
			IsCompleted:       false,
			CompanyName:       "Tienda Test",
			BusinessType:      "retail",
			StoreSize:         "micro",
			SelectedCategories: []string{},
			StepsCompleted:    []int{},
			StepsPending:      []int{1, 2, 3, 4, 5},
			StepsSkipped:      []int{},
			StartedAt:         now,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	}
}

func (m *OnboardingProcessMother) WithID(id uuid.UUID) *OnboardingProcessMother {
	m.process.ID = id
	return m
}

func (m *OnboardingProcessMother) WithTenantID(id uuid.UUID) *OnboardingProcessMother {
	m.process.TenantID = id
	return m
}

func (m *OnboardingProcessMother) WithUserID(id uuid.UUID) *OnboardingProcessMother {
	m.process.UserID = id
	return m
}

func (m *OnboardingProcessMother) WithCurrentStep(step int) *OnboardingProcessMother {
	m.process.CurrentStepNumber = step
	return m
}

func (m *OnboardingProcessMother) WithCompleted(completed bool) *OnboardingProcessMother {
	m.process.IsCompleted = completed
	if completed {
		now := time.Now()
		m.process.CompletedAt = &now
		m.process.StepsPending = []int{}
	}
	return m
}

func (m *OnboardingProcessMother) WithStepsCompleted(steps []int) *OnboardingProcessMother {
	m.process.StepsCompleted = steps
	return m
}

func (m *OnboardingProcessMother) WithStepsPending(steps []int) *OnboardingProcessMother {
	m.process.StepsPending = steps
	return m
}

func (m *OnboardingProcessMother) WithCompanyName(name string) *OnboardingProcessMother {
	m.process.CompanyName = name
	return m
}

func (m *OnboardingProcessMother) WithBusinessType(bt string) *OnboardingProcessMother {
	m.process.BusinessType = bt
	return m
}

func (m *OnboardingProcessMother) WithStoreSize(size string) *OnboardingProcessMother {
	m.process.StoreSize = size
	return m
}

func (m *OnboardingProcessMother) WithCategories(cats []string) *OnboardingProcessMother {
	m.process.SelectedCategories = cats
	return m
}

func (m *OnboardingProcessMother) WithNilTenant() *OnboardingProcessMother {
	m.process.TenantID = uuid.Nil
	return m
}

func (m *OnboardingProcessMother) WithNilUser() *OnboardingProcessMother {
	m.process.UserID = uuid.Nil
	return m
}

// AtStep3Verification returns a process at step 3 with steps 1,2 completed.
func (m *OnboardingProcessMother) AtStep3Verification() *OnboardingProcessMother {
	m.process.CurrentStepNumber = 3
	m.process.StepsCompleted = []int{1, 2}
	m.process.StepsPending = []int{3, 4, 5}
	return m
}

// AtStep5SelectPlan returns a process at step 5 with steps 1-4 completed.
func (m *OnboardingProcessMother) AtStep5SelectPlan() *OnboardingProcessMother {
	m.process.CurrentStepNumber = 5
	m.process.StepsCompleted = []int{1, 2, 3, 4}
	m.process.StepsPending = []int{5}
	return m
}

// AtStep6Complete returns a process at step 6 with steps 1-5 completed.
func (m *OnboardingProcessMother) AtStep6Complete() *OnboardingProcessMother {
	m.process.CurrentStepNumber = 6
	m.process.StepsCompleted = []int{1, 2, 3, 4, 5}
	m.process.StepsPending = []int{}
	return m
}

func (m *OnboardingProcessMother) Build() *entity.OnboardingProcess {
	return m.process
}
