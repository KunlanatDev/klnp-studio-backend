package application

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/klnp-studio/backend/internal/modules/auth/domain"
	appErrors "github.com/klnp-studio/backend/internal/shared/errors"
)

type Service struct {
	repo domain.ProfileRepository
}

func NewService(repo domain.ProfileRepository) *Service {
	return &Service{repo: repo}
}

type UpdateProfileInput struct {
	DisplayName *string
	AvatarURL   *string
}

func (s *Service) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	profile, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			return &domain.Profile{ID: userID}, nil
		}
		return nil, err
	}
	return profile, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, input UpdateProfileInput) (*domain.Profile, error) {
	profile, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, appErrors.ErrNotFound) {
			profile = &domain.Profile{ID: userID}
		} else {
			return nil, err
		}
	}

	if input.DisplayName != nil {
		trimmed := strings.TrimSpace(*input.DisplayName)
		if len(trimmed) > 80 {
			return nil, appErrors.ErrValidation
		}
		profile.DisplayName = trimmed
	}

	if input.AvatarURL != nil {
		profile.AvatarURL = strings.TrimSpace(*input.AvatarURL)
	}

	if err := s.repo.Update(ctx, profile); err != nil {
		return nil, err
	}
	return profile, nil
}
