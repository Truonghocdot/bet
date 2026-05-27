package service

import (
	"context"
	"log"
	"strings"
	"time"

	"gin/internal/domain/deposit"
	"gin/internal/domain/withdrawal"
	"gin/internal/domain/auth"
	"gin/internal/domain/user"
	repopg "gin/internal/repository/postgres"
)

type AffiliateService struct {
	userRepo          *repopg.UserRepository
	authService       *AuthService
	depositService    *DepositService
	withdrawalService *WithdrawalService
}

func NewAffiliateService(
	userRepo *repopg.UserRepository,
	authService *AuthService,
	depositService *DepositService,
	withdrawalService *WithdrawalService,
) *AffiliateService {
	return &AffiliateService{
		userRepo:          userRepo,
		authService:       authService,
		depositService:    depositService,
		withdrawalService: withdrawalService,
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

type ManagedAffiliateUserTransaction struct {
	ID            int64     `json:"id"`
	Unit          int       `json:"unit"`
	Direction     int       `json:"direction"`
	Amount        string    `json:"amount"`
	BalanceBefore string    `json:"balance_before"`
	BalanceAfter  string    `json:"balance_after"`
	ReferenceType string    `json:"reference_type"`
	ReferenceID   *int64    `json:"reference_id,omitempty"`
	Note          string    `json:"note"`
	CreatedAt     time.Time `json:"created_at"`
}

type ManagedAffiliateUserTransactionsResponse struct {
	Message string                            `json:"message"`
	Items   []ManagedAffiliateUserTransaction `json:"items"`
}

type ManagedAffiliateDeposit struct {
	ID            int64   `json:"id"`
	UserID        int64   `json:"user_id"`
	UserName      string  `json:"user_name"`
	UserPhone     string  `json:"user_phone"`
	ClientRef     string  `json:"client_ref"`
	Provider      string  `json:"provider"`
	ProviderTxnID *string `json:"provider_txn_id,omitempty"`
	Unit          int     `json:"unit"`
	Type          int     `json:"type"`
	Amount        string  `json:"amount"`
	NetAmount     string  `json:"net_amount"`
	Status        int     `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ApprovedAt    *time.Time `json:"approved_at,omitempty"`
	ReceivingAccount *struct {
		ID            int64   `json:"id"`
		ProviderCode  *string `json:"provider_code,omitempty"`
		AccountName   *string `json:"account_name,omitempty"`
		AccountNumber *string `json:"account_number,omitempty"`
	} `json:"receiving_account,omitempty"`
}

type ManagedAffiliateDepositsResponse struct {
	Message    string                   `json:"message"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	Total      int                      `json:"total"`
	TotalPages int                      `json:"total_pages"`
	Data       []ManagedAffiliateDeposit `json:"data"`
}

type ManagedAffiliateWithdrawal struct {
	ID            int64  `json:"id"`
	UserID        int64  `json:"user_id"`
	UserName      string `json:"user_name"`
	UserPhone     string `json:"user_phone"`
	Unit          int    `json:"unit"`
	Amount        string `json:"amount"`
	Fee           string `json:"fee"`
	NetAmount     string `json:"net_amount"`
	Status        int    `json:"status"`
	ReasonRejected string `json:"reason_rejected,omitempty"`
	AccountWithdrawalInfoID int64 `json:"account_withdrawal_info_id"`
	AccountName    string `json:"account_name"`
	AccountNumber  string `json:"account_number"`
	ProviderCode   string `json:"provider_code"`
	CreatedAt      time.Time `json:"created_at"`
}

type ManagedAffiliateWithdrawalsResponse struct {
	Message    string                      `json:"message"`
	Page       int                         `json:"page"`
	PageSize   int                         `json:"page_size"`
	Total      int                         `json:"total"`
	TotalPages int                         `json:"total_pages"`
	Data       []ManagedAffiliateWithdrawal `json:"data"`
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

func (s *AffiliateService) ManagedUserTransactions(ctx context.Context, referrerUserID int64, role int, managedUserID int64) (ManagedAffiliateUserTransactionsResponse, error) {
	if role != user.RoleAgency {
		return ManagedAffiliateUserTransactionsResponse{}, ErrUnauthorized
	}

	allowed, err := s.userRepo.IsManagedAffiliateUser(ctx, referrerUserID, managedUserID)
	if err != nil {
		return ManagedAffiliateUserTransactionsResponse{}, err
	}
	if !allowed {
		return ManagedAffiliateUserTransactionsResponse{}, ErrUnauthorized
	}

	records, err := s.userRepo.ListManagedAffiliateUserTransactions(ctx, managedUserID, 100)
	if err != nil {
		return ManagedAffiliateUserTransactionsResponse{}, err
	}

	items := make([]ManagedAffiliateUserTransaction, 0, len(records))
	for _, record := range records {
		items = append(items, ManagedAffiliateUserTransaction{
			ID:            record.ID,
			Unit:          record.Unit,
			Direction:     record.Direction,
			Amount:        record.Amount,
			BalanceBefore: record.BalanceBefore,
			BalanceAfter:  record.BalanceAfter,
			ReferenceType: record.ReferenceType,
			ReferenceID:   record.ReferenceID,
			Note:          record.Note,
			CreatedAt:     record.CreatedAt,
		})
	}

	return ManagedAffiliateUserTransactionsResponse{
		Message: "Lấy giao dịch người chơi thành công",
		Items:   items,
	}, nil
}

func (s *AffiliateService) ManagedDeposits(ctx context.Context, referrerUserID int64, role int, page int, pageSize int) (ManagedAffiliateDepositsResponse, error) {
	if role != user.RoleAgency {
		return ManagedAffiliateDepositsResponse{}, ErrUnauthorized
	}
	log.Printf("[affiliate][service.managed_deposits.start] referrer_user_id=%d role=%d page=%d page_size=%d", referrerUserID, role, page, pageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	total, err := s.userRepo.CountManagedAffiliateDeposits(ctx, referrerUserID)
	if err != nil {
		log.Printf("[affiliate][service.managed_deposits.count.error] referrer_user_id=%d err=%v", referrerUserID, err)
		return ManagedAffiliateDepositsResponse{}, err
	}
	totalPages := 1
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	if page > totalPages {
		page = totalPages
	}

	records, err := s.userRepo.ListManagedAffiliateDeposits(ctx, referrerUserID, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Printf("[affiliate][service.managed_deposits.list.error] referrer_user_id=%d page=%d page_size=%d err=%v", referrerUserID, page, pageSize, err)
		return ManagedAffiliateDepositsResponse{}, err
	}
	items := make([]ManagedAffiliateDeposit, 0, len(records))
	for _, record := range records {
		item := ManagedAffiliateDeposit{
			ID: record.ID, UserID: record.UserID, UserName: record.UserName, UserPhone: record.UserPhone,
			ClientRef: record.ClientRef, Provider: record.Provider, ProviderTxnID: record.ProviderTxnID,
			Unit: record.Unit, Type: record.Type, Amount: record.Amount, NetAmount: record.NetAmount,
			Status: record.Status, CreatedAt: record.CreatedAt, UpdatedAt: record.UpdatedAt, ApprovedAt: record.ApprovedAt,
		}
		if record.ReceivingAccountID != nil {
			item.ReceivingAccount = &struct {
				ID            int64   `json:"id"`
				ProviderCode  *string `json:"provider_code,omitempty"`
				AccountName   *string `json:"account_name,omitempty"`
				AccountNumber *string `json:"account_number,omitempty"`
			}{
				ID: *record.ReceivingAccountID, ProviderCode: record.ProviderCode, AccountName: record.AccountName, AccountNumber: record.AccountNumber,
			}
		}
		items = append(items, item)
	}
	log.Printf("[affiliate][service.managed_deposits.ok] referrer_user_id=%d total=%d page=%d page_size=%d items=%d", referrerUserID, total, page, pageSize, len(items))
	return ManagedAffiliateDepositsResponse{Message: "Lấy giao dịch nạp agency thành công", Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages, Data: items}, nil
}

func (s *AffiliateService) ManagedWithdrawals(ctx context.Context, referrerUserID int64, role int, page int, pageSize int) (ManagedAffiliateWithdrawalsResponse, error) {
	if role != user.RoleAgency {
		return ManagedAffiliateWithdrawalsResponse{}, ErrUnauthorized
	}
	log.Printf("[affiliate][service.managed_withdrawals.start] referrer_user_id=%d role=%d page=%d page_size=%d", referrerUserID, role, page, pageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}
	total, err := s.userRepo.CountManagedAffiliateWithdrawals(ctx, referrerUserID)
	if err != nil {
		log.Printf("[affiliate][service.managed_withdrawals.count.error] referrer_user_id=%d err=%v", referrerUserID, err)
		return ManagedAffiliateWithdrawalsResponse{}, err
	}
	totalPages := 1
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	if page > totalPages {
		page = totalPages
	}
	records, err := s.userRepo.ListManagedAffiliateWithdrawals(ctx, referrerUserID, pageSize, (page-1)*pageSize)
	if err != nil {
		log.Printf("[affiliate][service.managed_withdrawals.list.error] referrer_user_id=%d page=%d page_size=%d err=%v", referrerUserID, page, pageSize, err)
		return ManagedAffiliateWithdrawalsResponse{}, err
	}
	items := make([]ManagedAffiliateWithdrawal, 0, len(records))
	for _, record := range records {
		items = append(items, ManagedAffiliateWithdrawal{
			ID: record.ID, UserID: record.UserID, UserName: record.UserName, UserPhone: record.UserPhone,
			Unit: record.Unit, Amount: record.Amount, Fee: record.Fee, NetAmount: record.NetAmount,
			Status: record.Status, ReasonRejected: record.ReasonRejected, AccountWithdrawalInfoID: record.AccountWithdrawalInfoID,
			AccountName: record.AccountName, AccountNumber: record.AccountNumber, ProviderCode: record.ProviderCode, CreatedAt: record.CreatedAt,
		})
	}
	log.Printf("[affiliate][service.managed_withdrawals.ok] referrer_user_id=%d total=%d page=%d page_size=%d items=%d", referrerUserID, total, page, pageSize, len(items))
	return ManagedAffiliateWithdrawalsResponse{Message: "Lấy giao dịch rút agency thành công", Page: page, PageSize: pageSize, Total: total, TotalPages: totalPages, Data: items}, nil
}

func (s *AffiliateService) ManagedUserDeposits(
	ctx context.Context,
	referrerUserID int64,
	role int,
	managedUserID int64,
	page int,
	pageSize int,
) (deposit.DepositHistoryResponse, error) {
	if role != user.RoleAgency {
		return deposit.DepositHistoryResponse{}, ErrUnauthorized
	}

	allowed, err := s.userRepo.IsManagedAffiliateUser(ctx, referrerUserID, managedUserID)
	if err != nil {
		return deposit.DepositHistoryResponse{}, err
	}
	if !allowed {
		return deposit.DepositHistoryResponse{}, ErrUnauthorized
	}

	return s.depositService.ListHistory(ctx, managedUserID, page, pageSize)
}

func (s *AffiliateService) ManagedUserWithdrawals(
	ctx context.Context,
	referrerUserID int64,
	role int,
	managedUserID int64,
	page int,
	pageSize int,
) (withdrawal.WithdrawalHistoryResponse, error) {
	if role != user.RoleAgency {
		return withdrawal.WithdrawalHistoryResponse{}, ErrUnauthorized
	}

	allowed, err := s.userRepo.IsManagedAffiliateUser(ctx, referrerUserID, managedUserID)
	if err != nil {
		return withdrawal.WithdrawalHistoryResponse{}, err
	}
	if !allowed {
		return withdrawal.WithdrawalHistoryResponse{}, ErrUnauthorized
	}

	return s.withdrawalService.ListHistory(ctx, managedUserID, page, pageSize)
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
