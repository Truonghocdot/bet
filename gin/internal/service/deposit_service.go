package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"time"

	"gin/internal/domain/deposit"
	"gin/internal/domain/user"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/clock"
	"gin/internal/support/id"
	"gin/internal/support/message"
	goredis "github.com/redis/go-redis/v9"
)

type DepositConfig struct {
	ReceivingAccountsRedisKey string
}

type DepositService struct {
	repository *repopg.DepositRepository
	redis      *goredis.Client
	config     DepositConfig
}

func NewDepositService(repository *repopg.DepositRepository, redis *goredis.Client, config DepositConfig) *DepositService {
	return &DepositService{
		repository: repository,
		redis:      redis,
		config:     config,
	}
}

func (s *DepositService) InitVietQRDeposit(ctx context.Context, userID int64, request deposit.DepositInitRequest) (deposit.DepositInitResponse, error) {
	return s.initDeposit(ctx, userID, request, deposit.DepositMethodVietQR, deposit.DepositProviderSepayVietQR, deposit.ReceivingAccountTypeBank, user.WalletUnitVND)
}

func (s *DepositService) InitUSDTDeposit(ctx context.Context, userID int64, request deposit.DepositInitRequest) (deposit.DepositInitResponse, error) {
	return s.initDeposit(ctx, userID, request, deposit.DepositMethodUSDT, deposit.DepositProviderUSDTGateway, deposit.ReceivingAccountTypeCrypto, user.WalletUnitUSDT)
}

func (s *DepositService) GetDepositStatus(ctx context.Context, userID int64, clientRef string) (deposit.DepositStatusResponse, error) {
	record, err := s.repository.FindDepositIntentByClientRef(ctx, clientRef)
	if err != nil {
		return deposit.DepositStatusResponse{}, err
	}

	if record.UserID != userID {
		return deposit.DepositStatusResponse{}, fmt.Errorf(message.Unauthorized)
	}

	return deposit.DepositStatusResponse{
		Message:          message.DepositAccepted,
		Transaction:      s.toDomainTransaction(record),
		ReceivingAccount: s.toDomainReceivingAccount(record.ReceivingAccount),
	}, nil
}

func (s *DepositService) ApplyDeposit(ctx context.Context, request deposit.ApplyDepositRequest) (deposit.ApplyDepositResponse, error) {
	amount := normalizeDepositAmount(request.Amount)
	if amount == "" {
		return deposit.ApplyDepositResponse{}, fmt.Errorf(message.DepositAmountInvalid)
	}

	if strings.TrimSpace(request.Provider) == "" {
		return deposit.ApplyDepositResponse{}, fmt.Errorf(message.DepositProviderInvalid)
	}

	result, err := s.repository.ApplyDeposit(ctx, repopg.ApplyDepositParams{
		Provider:       request.Provider,
		ProviderStatus: request.ProviderStatus,
		ClientRef:      request.ClientRef,
		ProviderTxnID:  request.ProviderTxnID,
		Amount:         amount,
		Currency:       request.Currency,
		PaidAt:         request.PaidAt,
		Raw:            request.Raw,
	})
	if err != nil {
		return deposit.ApplyDepositResponse{}, err
	}

	status := "pending"
	if result.AlreadyDone {
		status = "completed"
	} else if result.Applied {
		status = "completed"
	} else if result.Transaction.Status == 4 {
		status = "failed"
	}

	messageText := message.DepositAccepted
	if result.Applied {
		messageText = "Giao dịch nạp đã được xác nhận và cộng ví"
	} else if result.AlreadyDone {
		messageText = "Giao dịch nạp đã được xử lý trước đó"
	} else if result.Transaction.Status == 4 {
		messageText = "Giao dịch nạp không thành công"
	}

	return deposit.ApplyDepositResponse{
		Message:   messageText,
		ClientRef: request.ClientRef,
		Status:    status,
		AppliedAt: request.PaidAt,
	}, nil
}

func (s *DepositService) initDeposit(
	ctx context.Context,
	userID int64,
	request deposit.DepositInitRequest,
	method deposit.DepositMethod,
	provider deposit.DepositProvider,
	accountType deposit.ReceivingAccountType,
	unit int,
) (deposit.DepositInitResponse, error) {
	amount := normalizeDepositAmount(request.Amount)
	if amount == "" {
		return deposit.DepositInitResponse{}, fmt.Errorf(message.DepositAmountRequired)
	}

	candidates, err := s.receivingAccounts(ctx, unit, int(accountType))
	if err != nil {
		return deposit.DepositInitResponse{}, err
	}

	selected, err := chooseReceivingAccount(candidates)
	if err != nil {
		return deposit.DepositInitResponse{}, err
	}

	walletID, _, err := s.repository.FindWalletByUserAndUnit(ctx, userID, unit)
	if err != nil {
		return deposit.DepositInitResponse{}, err
	}

	clientRef := "DEP-" + id.New()
	meta := map[string]any{
		"method":              method,
		"provider":            provider,
		"selected_account":    selected.Code,
		"selected_account_id": selected.ID,
		"note":                strings.TrimSpace(request.Note),
	}

	record, err := s.repository.CreateDepositIntent(ctx, repopg.CreateDepositIntentParams{
		UserID:             userID,
		WalletID:           walletID,
		ClientRef:          clientRef,
		Unit:               unit,
		Type:               1,
		Amount:             amount,
		Status:             1,
		Provider:           string(provider),
		ReceivingAccountID: &selected.ID,
		Meta:               meta,
	})
	if err != nil {
		return deposit.DepositInitResponse{}, err
	}

	expiresAt := clock.Now().Add(15 * time.Minute)
	transaction := s.toDomainTransaction(record)
	transaction.ReceivingAccount = s.toDomainReceivingAccount(&selected)

	response := deposit.DepositInitResponse{
		Message:          message.DepositCreated,
		Provider:         string(provider),
		Method:           method,
		ClientRef:        clientRef,
		Amount:           amount,
		Transaction:      transaction,
		ExpiresAt:        expiresAt,
		ReceivingAccount: s.toDomainReceivingAccount(&selected),
	}

	switch method {
	case deposit.DepositMethodVietQR:
		response.Instructions = firstNonEmptyString(selected.Instructions, "Quét QR hoặc chuyển khoản đúng nội dung để hệ thống tự động đối soát.")
		response.QRContent = buildQRContent(selected, amount, clientRef)
		response.QRCodeURL = firstNonEmptyString(selected.QRCodePath, "")
		response.PayURL = ""
	case deposit.DepositMethodUSDT:
		response.Instructions = firstNonEmptyString(selected.Instructions, "Chuyển đúng mạng lưới và đúng số tiền để hệ thống tự động đối soát.")
		response.QRContent = buildUSDTContent(selected, amount, clientRef)
	}

	return response, nil
}

func (s *DepositService) receivingAccounts(ctx context.Context, unit int, accountType int) ([]repopg.ReceivingAccountRecord, error) {
	snapshot, err := s.loadReceivingAccountsSnapshot(ctx)
	if err == nil && len(snapshot) > 0 {
		filtered := filterReceivingAccounts(snapshot, unit, accountType)
		if len(filtered) > 0 {
			return filtered, nil
		}
	}

	accounts, err := s.repository.ListActiveReceivingAccounts(ctx)
	if err != nil {
		return nil, err
	}

	return filterReceivingAccounts(accounts, unit, accountType), nil
}

func (s *DepositService) loadReceivingAccountsSnapshot(ctx context.Context) ([]repopg.ReceivingAccountRecord, error) {
	if s.redis == nil || strings.TrimSpace(s.config.ReceivingAccountsRedisKey) == "" {
		return nil, fmt.Errorf("redis disabled")
	}

	raw, err := s.redis.Get(ctx, s.config.ReceivingAccountsRedisKey).Result()
	if err != nil {
		return nil, err
	}

	var payload struct {
		Accounts []repopg.ReceivingAccountRecord `json:"accounts"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, err
	}

	return payload.Accounts, nil
}

func (s *DepositService) toDomainReceivingAccount(record *repopg.ReceivingAccountRecord) *deposit.ReceivingAccount {
	if record == nil {
		return nil
	}

	return &deposit.ReceivingAccount{
		ID:            record.ID,
		Code:          record.Code,
		Name:          record.Name,
		Type:          record.Type,
		Unit:          record.Unit,
		ProviderCode:  record.ProviderCode,
		AccountName:   record.AccountName,
		AccountNumber: record.AccountNumber,
		WalletAddress: record.WalletAddress,
		Network:       record.Network,
		QRCodePath:    record.QRCodePath,
		Instructions:  record.Instructions,
		Status:        record.Status,
		IsDefault:     record.IsDefault,
		SortOrder:     record.SortOrder,
	}
}

func (s *DepositService) toDomainTransaction(record repopg.DepositTransactionRecord) deposit.DepositTransaction {
	var meta map[string]any
	if len(record.Meta) > 0 {
		_ = json.Unmarshal(record.Meta, &meta)
	}

	return deposit.DepositTransaction{
		ID:               record.ID,
		ClientRef:        record.ClientRef,
		Provider:         record.Provider,
		ProviderTxnID:    record.ProviderTxnID,
		ReceivingAccount: s.toDomainReceivingAccount(record.ReceivingAccount),
		Unit:             record.Unit,
		Type:             record.Type,
		Amount:           record.Amount,
		NetAmount:        record.NetAmount,
		Status:           record.Status,
		Meta:             meta,
		CreatedAt:        record.CreatedAt,
		UpdatedAt:        record.UpdatedAt,
		ApprovedAt:       record.ApprovedAt,
	}
}

func filterReceivingAccounts(accounts []repopg.ReceivingAccountRecord, unit int, accountType int) []repopg.ReceivingAccountRecord {
	filtered := make([]repopg.ReceivingAccountRecord, 0, len(accounts))
	for _, account := range accounts {
		if account.Unit == unit && account.Type == accountType {
			filtered = append(filtered, account)
		}
	}

	return filtered
}

func chooseReceivingAccount(accounts []repopg.ReceivingAccountRecord) (repopg.ReceivingAccountRecord, error) {
	if len(accounts) == 0 {
		return repopg.ReceivingAccountRecord{}, repopg.ErrDepositReceivingAccount
	}

	defaults := make([]repopg.ReceivingAccountRecord, 0, len(accounts))
	for _, account := range accounts {
		if account.IsDefault {
			defaults = append(defaults, account)
		}
	}

	candidates := accounts
	if len(defaults) > 0 {
		candidates = defaults
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}

	index, err := randomIndex(len(candidates))
	if err != nil {
		return repopg.ReceivingAccountRecord{}, err
	}

	return candidates[index], nil
}

func randomIndex(length int) (int, error) {
	if length <= 0 {
		return 0, fmt.Errorf(message.DepositReceivingAccountMissing)
	}

	seed := clock.Now().UnixNano()
	randSrc := rand.New(rand.NewSource(seed))
	return randSrc.Intn(length), nil
}

func normalizeDepositAmount(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.ReplaceAll(trimmed, ",", "")
	trimmed = strings.ReplaceAll(trimmed, " ", "")
	if trimmed == "" {
		return ""
	}

	if _, ok := new(big.Float).SetString(trimmed); !ok {
		return ""
	}

	return trimmed
}

func buildQRContent(account repopg.ReceivingAccountRecord, amount string, clientRef string) string {
	return strings.Join([]string{
		"VIETQR",
		account.Code,
		firstNonEmptyString(account.AccountNumber, ""),
		amount,
		clientRef,
	}, "|")
}

func buildUSDTContent(account repopg.ReceivingAccountRecord, amount string, clientRef string) string {
	return strings.Join([]string{
		"USDT",
		firstNonEmptyString(account.WalletAddress, ""),
		firstNonEmptyString(account.Network, ""),
		amount,
		clientRef,
	}, "|")
}

func firstNonEmptyString(value *string, fallback string) string {
	if value == nil {
		return fallback
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return fallback
	}

	return trimmed
}
