package service

import (
	"context"
	"strings"
	"time"

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

type ManagedAffiliateUser struct {
	UserID                    int64     `json:"user_id"`
	Name                      string    `json:"name"`
	Phone                     string    `json:"phone"`
	CreatedAt                 time.Time `json:"created_at"`
	ReferralStatus            int       `json:"referral_status"`
	FirstDepositAmount        string    `json:"first_deposit_amount"`
	FirstDepositTransactionID int64     `json:"first_deposit_transaction_id"`
	FirstDepositTransactionNo string    `json:"first_deposit_transaction_no"`
}

type ManagedAffiliateUsersResponse struct {
	Message string                 `json:"message"`
	Items   []ManagedAffiliateUser `json:"items"`
}

func (s *AffiliateService) Summary(ctx context.Context, userID int64) (AffiliateSummary, error) {
	count, err := s.userRepo.CountInvitedUsers(ctx, userID)
	if err != nil {
		return AffiliateSummary{}, err
	}
	return AffiliateSummary{InvitedUsersCount: count}, nil
}

func (s *AffiliateService) ManagedUsers(ctx context.Context, userID int64, role int) (ManagedAffiliateUsersResponse, error) {
	if role != user.RoleAgency {
		return ManagedAffiliateUsersResponse{}, ErrUnauthorized
	}

	items, err := s.userRepo.ListManagedAffiliateUsers(ctx, userID, 200)
	if err != nil {
		return ManagedAffiliateUsersResponse{}, err
	}

	result := make([]ManagedAffiliateUser, 0, len(items))
	for _, item := range items {
		result = append(result, ManagedAffiliateUser{
			UserID:                    item.UserID,
			Name:                      item.Name,
			Phone:                     item.Phone,
			CreatedAt:                 item.CreatedAt,
			ReferralStatus:            item.ReferralStatus,
			FirstDepositAmount:        item.FirstDepositAmount,
			FirstDepositTransactionID: item.FirstDepositTransactionID,
			FirstDepositTransactionNo: item.FirstDepositTransactionNo,
		})
	}

	return ManagedAffiliateUsersResponse{
		Message: "Lấy danh sách user trực thuộc thành công",
		Items:   result,
	}, nil
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
