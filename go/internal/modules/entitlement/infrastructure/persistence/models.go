package persistence

import "time"

type PlanModel struct {
	ID       string `gorm:"primaryKey"`
	Code     string
	Name     string
	IsActive bool
}

func (PlanModel) TableName() string {
	return "plans"
}

type PlanLimitModel struct {
	ID         string `gorm:"primaryKey"`
	PlanID     string
	FeatureKey string
	LimitValue int64
}

func (PlanLimitModel) TableName() string {
	return "plan_limits"
}

type PlanFeatureModel struct {
	ID         string `gorm:"primaryKey"`
	PlanID     string
	FeatureKey string
	ValueType  string
	Value      []byte `gorm:"type:jsonb"`
}

func (PlanFeatureModel) TableName() string {
	return "plan_features"
}

type AssignmentModel struct {
	ID          string `gorm:"primaryKey"`
	WorkspaceID string
	PlanID      string
	StartsAt    time.Time
	EndsAt      *time.Time
}

func (AssignmentModel) TableName() string {
	return "workspace_plan_assignments"
}

type WorkspaceUsageModel struct {
	WorkspaceID string `gorm:"primaryKey"`
	PagesCount  int64
}

func (WorkspaceUsageModel) TableName() string {
	return "workspace_usage"
}

