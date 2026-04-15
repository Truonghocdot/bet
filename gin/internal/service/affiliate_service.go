package service

import (
	"context"
	"strings"

	"gin/internal/domain/auth"
	"gin/internal/domain/user"
	repopg "gin/internal/repository/postgres"
)

type AffiliateService struct {
	userRepo    *repopg.UserRepository
	authService *AuthService
}

func NewAffiliateService(userRepo *repopg.UserRepository, authService *AuthService) *AffiliateService {
	return &AffiliateService{
		userRepo:    userRepo,
		authService: authService,
	}
}

type AffiliateSummary struct {
	InvitedUsersCount int64 `json:"invited_users_count"`
}

func (s *AffiliateService) Summary(ctx context.Context, userID int64) (AffiliateSummary, error) {
	count, err := s.userRepo.CountInvitedUsers(ctx, userID)
	if err != nil {
		return AffiliateSummary{}, err
	}
	return AffiliateSummary{InvitedUsersCount: count}, nil
}

func (s *AffiliateService) BecomeAgency(ctx context.Context, userID int64, role int, staffRefCode string) (auth.AuthResponse, error) {
	if role != user.RoleClient {
		return auth.AuthResponse{}, ErrUnauthorized
	}

	if strings.TrimSpace(staffRefCode) == "" {
		return auth.AuthResponse{}, repopg.ErrStaffInviteInvalid
	}

	if err := s.userRepo.PromoteToAgencyByStaffRefCode(ctx, userID, staffRefCode); err != nil {
		return auth.AuthResponse{}, err
	}

	// Issue new token with updated role so FE updates immediately.
	return s.authService.LoginByUserID(ctx, userID)
}
