package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PlanAssignmentRepository struct {
	db *gorm.DB
}

func NewPlanAssignmentRepository(db *gorm.DB) *PlanAssignmentRepository {
	return &PlanAssignmentRepository{db: db}
}

func (r *PlanAssignmentRepository) ChangePlan(
	ctx context.Context,
	workspaceID string,
	planID string,
	actorID string,
	idempotencyKey string,
) error {
	if workspaceID == "" || planID == "" || actorID == "" {
		return errors.New("invalid assignment request")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock workspace to serialize plan changes
		var workspace struct {
			ID string `gorm:"column:id"`
		}

		err := tx.Table("workspaces").
			Clauses(clause.Locking{
				Strength: "UPDATE",
			}).
			Select("id").
			Where("id = ?", workspaceID).
			Take(&workspace).Error

		if err != nil {
			return err
		}

		// Validate target plan
		var plan PlanModel
		err = tx.Where("id = ? AND is_active = TRUE", planID).Take(&plan).Error
		if err != nil {
			return fmt.Errorf("target plan unavailable: %w", err)
		}

		now := time.Now().UTC()

		// End current assignment
		err = tx.Model(&AssignmentModel{}).
			Where("workspace_id = ? AND ends_at IS NULL", workspaceID).
			Update("ends_at", now).Error

		if err != nil {
			return err
		}

		// Create new assignment
		assignment := AssignmentModel{
			WorkspaceID: workspaceID,
			PlanID:      planID,
			StartsAt:    now,
		}

		if err := tx.Create(&assignment).Error; err != nil {
			return err
		}

		return nil
	})
}

