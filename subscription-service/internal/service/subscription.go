package service

import (
	"subscription/internal/dto"
	"subscription/internal/exception"
	"subscription/internal/models"
	"subscription/internal/repository"
	"time"

	"github.com/google/uuid"
)

type SubscriptionService struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionService(repo *repository.SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{
		repo: repo,
	}
}

func (s *SubscriptionService) CreateSub(subDTO *dto.CreateSubscriptionDTO) (*dto.SubscriptionDTO, error) {
	sub := &models.Subscription{
		UserID:    subDTO.UserID,
		StartsAt:  subDTO.StartsAt,
		ExpiresAt: subDTO.ExpiresAt,
	}
	sub, err := s.repo.CreateSub(sub)
	if err != nil {
		return nil, err
	}
	return &dto.SubscriptionDTO{
		Uuid:      sub.Uuid,
		UserID:    sub.UserID,
		StartsAt:  sub.StartsAt,
		ExpiresAt: sub.ExpiresAt,
	}, nil
}

func (s *SubscriptionService) GetSubByUuid(id uuid.UUID) (*dto.SubscriptionDTO, error) {
	sub, err := s.repo.GetSubByUuid(id)
	if err != nil {
		return nil, err
	}
	return &dto.SubscriptionDTO{
		Uuid:      sub.Uuid,
		UserID:    sub.UserID,
		StartsAt:  sub.StartsAt,
		ExpiresAt: sub.ExpiresAt,
	}, nil
}

func (s *SubscriptionService) GetOrCreateSub(subDTO *dto.CreateSubscriptionDTO) (*dto.SubscriptionDTO, error) {
	sub, err := s.GetSubByUserId(subDTO.UserID)
	if err != nil {
		if err == exception.ErrSubscriptionNotFound {
			sub, err = s.CreateSub(subDTO)
			if err != nil {
				return nil, err
			}
			return sub, nil
		}
		return nil, err
	}
	return sub, nil
}

func (s *SubscriptionService) EnsureSubByUserId(userID uuid.UUID) (*dto.SubscriptionDTO, error) {
	sub, err := s.GetSubByUserId(userID)
	if err == nil {
		return sub, nil
	}
	if err != exception.ErrSubscriptionNotFound {
		return nil, err
	}

	now := time.Now()
	return s.CreateSub(&dto.CreateSubscriptionDTO{
		UserID:    userID,
		StartsAt:  now,
		ExpiresAt: now.AddDate(0, 1, 0),
	})
}

func (s *SubscriptionService) GetSubByUserId(userId uuid.UUID) (*dto.SubscriptionDTO, error) {
	sub, err := s.repo.GetSubByUserId(userId)
	if err != nil {
		return nil, err
	}
	return &dto.SubscriptionDTO{
		Uuid:      sub.Uuid,
		UserID:    sub.UserID,
		StartsAt:  sub.StartsAt,
		ExpiresAt: sub.ExpiresAt,
	}, nil
}
