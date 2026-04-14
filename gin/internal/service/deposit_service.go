package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net/url"
	"strings"
	"time"

	"gin/internal/domain/deposit"
	"gin/internal/domain/user"
	gateclient "gin/internal/integration/gate"
	repopg "gin/internal/repository/postgres"
	"gin/internal/support/clock"
	"gin/internal/support/id"
	"gin/internal/support/message"
	goredis "github.com/redis/go-redis/v9"
)

var ErrDepositUSDTNotAvailable = errors.New(message.DepositUSDTNotAvailable)
var ErrDepositUSDTTemporarilyClosed = errors.New(message.DepositUSDTTemporarilyClosed)

const localUSDTMinAmount = "20"

type DepositConfig struct {
	ReceivingAccountsRedisKey string
}

type DepositService struct {
	repository *repopg.DepositRepository
	redis      *goredis.Client 
	config     DepositConfig
	wallets    *WalletService
	gate       *gateclient.DepositClient
}

func NewDepositService(
	repository *repopg.DepositRepository,
	redis *goredis.Client,
	wallets *WalletService,
	gate *gateclient.DepositClient,
	config DepositConfig,
) *DepositService {
	return &DepositService{
		repository: repository,
		redis:      redis,
		config:     config,
		wallets:    wallets,
		gate:       gate,
	}
}

func (s *DepositService) InitVietQRDeposit(ctx context.Context, userID int64, request deposit.DepositInitRequest) (deposit.DepositInitResponse, error) {
	return s.initDeposit(ctx, userID, request, deposit.DepositMethodVietQR, deposit.DepositProviderSepayVietQR, deposit.ReceivingAccountTypeBank, user.WalletUnitVND)
}

func (s *DepositService) InitUSDTDeposit(ctx context.Context, userID int64, request deposit.DepositInitRequest) (deposit.DepositInitResponse, error) {
	traceID := strings.TrimSpace(id.New())
	if s.gate == nil {
		log.Printf("[deposit][usdt.init.error] trace_id=%s user_id=%d reason=gate_client_nil", traceID, userID)
		return deposit.DepositInitResponse{}, ErrDepositUSDTNotAvailable
	}

	if cached, err := s.loadPendingDepositCache(ctx, userID, deposit.DepositMethodUSDT); err == nil && cached.ClientRef != "" {
		log.Printf("[deposit][usdt.init.cache_hit] trace_id=%s user_id=%d client_ref=%s status=%d", traceID, userID, cached.ClientRef, cached.Transaction.Status)
		return cached, nil
	}

	amount := normalizeDepositAmount(request.Amount)
	log.Printf("[deposit][usdt.init.input] trace_id=%s user_id=%d raw_amount=%q normalized_amount=%q", traceID, userID, request.Amount, amount)
	if amount == "" {
		log.Printf("[deposit][usdt.init.error] trace_id=%s user_id=%d reason=amount_required raw_amount=%q", traceID, userID, request.Amount)
		return deposit.DepositInitResponse{}, fmt.Errorf(message.DepositAmountRequired)
	}

	amountRat, _ := new(big.Rat).SetString(amount)
	minUSDT, _ := new(big.Rat).SetString(localUSDTMinAmount)
	log.Printf("[deposit][usdt.init.validate] trace_id=%s user_id=%d amount=%s local_min=%s", traceID, userID, amount, localUSDTMinAmount)
	if amountRat == nil || amountRat.Cmp(minUSDT) < 0 {
		log.Printf("[deposit][usdt.init.error] trace_id=%s user_id=%d reason=amount_invalid amount=%s local_min=%s", traceID, userID, amount, localUSDTMinAmount)
		return deposit.DepositInitResponse{}, fmt.Errorf(message.DepositAmountInvalid)
	}

	walletID, _, err := s.repository.FindWalletByUserAndUnit(ctx, userID, user.WalletUnitUSDT)
	if err != nil {
		log.Printf("[deposit][usdt.init.error] trace_id=%s user_id=%d reason=find_wallet_failed err=%v", traceID, userID, err)
		return deposit.DepositInitResponse{}, err
	}

	clientRef := "DEP-" + id.New()
	log.Printf("[deposit][usdt.init.gate.start] trace_id=%s user_id=%d client_ref=%s amount=%s", traceID, userID, clientRef, amount)
	created, err := s.gate.CreateNowPaymentsDeposit(ctx, gateclient.CreateNowPaymentsDepositRequest{
		ClientRef: clientRef,
		Amount:    amount,
	})
	if err != nil {
		log.Printf("[deposit][usdt.init.error] trace_id=%s user_id=%d client_ref=%s reason=gate_create_nowpayments_failed err=%v", traceID, userID, clientRef, err)
		errText := strings.ToUpper(strings.TrimSpace(err.Error()))
		if strings.Contains(errText, "CURRENCY_UNAVAILABLE") {
			return deposit.DepositInitResponse{}, fmt.Errorf("%w: %v", ErrDepositUSDTTemporarilyClosed, err)
		}
		return deposit.DepositInitResponse{}, fmt.Errorf("%w: %v", ErrDepositUSDTNotAvailable, err)
	}

	meta := map[string]any{
		"method":         deposit.DepositMethodUSDT,
		"provider":       deposit.DepositProviderNowPayments,
		"payment_id":     strings.TrimSpace(created.PaymentID),
		"payment_status": strings.TrimSpace(created.PaymentStatus),
		"pay_currency":   strings.TrimSpace(created.PayCurrency),
		"pay_amount":     strings.TrimSpace(created.PayAmount),
		"pay_address":    strings.TrimSpace(created.PayAddress),
		"payin_extra_id": strings.TrimSpace(created.PayinExtraID),
		"invoice_url":    strings.TrimSpace(created.InvoiceURL),
		"raw":            created.Raw,
	}

	var providerTxnID *string
	if trimmed := strings.TrimSpace(created.PaymentID); trimmed != "" {
		providerTxnID = &trimmed
	}

	record, err := s.repository.CreateDepositIntent(ctx, repopg.CreateDepositIntentParams{
		UserID:             userID,
		WalletID:           walletID,
		ClientRef:          clientRef,
		Unit:               user.WalletUnitUSDT,
		Type:               1,
		Amount:             amount,
		Status:             1,
		Provider:           string(deposit.DepositProviderNowPayments),
		ProviderTxnID:      providerTxnID,
		ReceivingAccountID: nil,
		Meta:               meta,
	})
	if err != nil {
		log.Printf("[deposit][usdt.init.error] trace_id=%s user_id=%d client_ref=%s reason=create_deposit_intent_failed err=%v", traceID, userID, clientRef, err)
		return deposit.DepositInitResponse{}, err
	}

	expiresAt := clock.Now().Add(10 * time.Minute)
	transaction := s.toDomainTransaction(record)

	network := strings.ToUpper(strings.TrimSpace(created.PayCurrency))
	address := strings.TrimSpace(created.PayAddress)
	memo := strings.TrimSpace(created.PayinExtraID)

	var qrContent string
	if memo != "" {
		qrContent = fmt.Sprintf("%s|memo:%s", address, memo)
	} else {
		qrContent = address
	}

	receiving := &deposit.ReceivingAccount{
		Type:          0,
		Unit:          user.WalletUnitUSDT,
		ProviderCode:  stringPtrOrNil(network),
		AccountName:   stringPtrOrNil("NOWPayments"),
		AccountNumber: stringPtrOrNil(address),
		Status:        1,
	}
	transaction.ReceivingAccount = receiving

	response := deposit.DepositInitResponse{
		Message:          message.DepositCreated,
		Provider:         string(deposit.DepositProviderNowPayments),
		Method:           deposit.DepositMethodUSDT,
		ClientRef:        clientRef,
		Amount:           amount,
		Transaction:      transaction,
		Instructions:     "Chuyen dung dia chi vi va memo/tag (neu co) de he thong doi soat tu dong.",
		QRContent:        qrContent,
		QRCodeURL:        "",
		PayURL:           strings.TrimSpace(created.InvoiceURL),
		ExpiresAt:        expiresAt,
		ReceivingAccount: receiving,
	}

	if err := s.savePendingDepositCache(ctx, userID, deposit.DepositMethodUSDT, response); err != nil {
		log.Printf("[deposit][cache.save.error] trace_id=%s user_id=%d method=%s client_ref=%s err=%v", traceID, userID, deposit.DepositMethodUSDT, clientRef, err)
	}
	log.Printf(
		"[deposit][usdt.init.ok] trace_id=%s user_id=%d client_ref=%s payment_id=%s pay_currency=%s pay_amount=%s pay_address=%s",
		traceID,
		userID,
		clientRef,
		strings.TrimSpace(created.PaymentID),
		strings.TrimSpace(created.PayCurrency),
		strings.TrimSpace(created.PayAmount),
		strings.TrimSpace(created.PayAddress),
	)

	return response, nil
}

func (s *DepositService) GetDepositStatus(ctx context.Context, userID int64, clientRef string) (deposit.DepositStatusResponse, error) {
	record, err := s.repository.FindDepositIntentByClientRef(ctx, clientRef)
	if err != nil {
		return deposit.DepositStatusResponse{}, err
	}

	if record.UserID != userID {
		return deposit.DepositStatusResponse{}, fmt.Errorf(message.Unauthorized)
	}

	response := deposit.DepositStatusResponse{
		Message:          message.DepositAccepted,
		Transaction:      s.toDomainTransaction(record),
		ReceivingAccount: s.toDomainReceivingAccount(record.ReceivingAccount),
	}

	if response.Transaction.Status == 2 || response.Transaction.Status == 3 || response.Transaction.Status == 4 {
		_ = s.clearPendingDepositCache(ctx, userID, depositProviderToMethod(record.Provider))
	}

	return response, nil
}

func (s *DepositService) ListVietQrBanks(ctx context.Context) (deposit.DepositBankListResponse, error) {
	records, err := s.repository.ListVietQrBanks(ctx)
	if err != nil {
		return deposit.DepositBankListResponse{}, err
	}

	items := make([]deposit.DepositBankOption, 0, len(records))
	for _, record := range records {
		items = append(items, deposit.DepositBankOption{
			ProviderCode: record.ProviderCode,
			ShortName:    record.ShortName,
			Name:         record.Name,
			Bin:          record.Bin,
			Logo:         firstNonNilString(record.Logo),
			AccountCount: record.AccountCount,
			IsDefault:    record.IsDefault,
		})
	}

	return deposit.DepositBankListResponse{
		Message: message.DepositAccepted,
		Banks:   items,
	}, nil
}

func (s *DepositService) ApplyDeposit(ctx context.Context, request deposit.ApplyDepositRequest) (deposit.ApplyDepositResponse, error) {
	amount := normalizeDepositAmount(request.Amount)
	// Trạng thái 'waiting' có thể đi kèm amount = 0, ta vẫn cho phép đi tiếp 
	// để cập nhật trạng thái đơn hàng trong DB.
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

	if result.Applied && s.wallets != nil {
		if err := s.wallets.PublishSummary(ctx, result.Transaction.UserID); err != nil {
			log.Printf("[realtime][wallet.publish.error] user_id=%d source=deposit.apply err=%v", result.Transaction.UserID, err)
		}
	}

	if result.Applied || result.AlreadyDone || result.Transaction.Status == 4 {
		_ = s.clearPendingDepositCache(ctx, result.Transaction.UserID, depositProviderToMethod(result.Transaction.Provider))
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
	if cached, err := s.loadPendingDepositCache(ctx, userID, method); err == nil && cached.ClientRef != "" {
		return cached, nil
	}

	amount := normalizeDepositAmount(request.Amount)
	if amount == "" {
		return deposit.DepositInitResponse{}, fmt.Errorf(message.DepositAmountRequired)
	}

	candidates, err := s.receivingAccounts(ctx, unit, int(accountType), strings.TrimSpace(request.ProviderCode))
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

	expiresAt := clock.Now().Add(10 * time.Minute)
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
		response.Instructions = "Quét QR hoặc chuyển khoản đúng nội dung để hệ thống tự động đối soát."
		response.QRContent = buildQRContent(selected, amount, clientRef)
		response.QRCodeURL = buildVietQrImageURL(selected, amount, clientRef)
		response.PayURL = ""
	}

	if err := s.savePendingDepositCache(ctx, userID, method, response); err != nil {
		log.Printf("[deposit][cache.save.error] user_id=%d method=%s client_ref=%s err=%v", userID, method, clientRef, err)
	}

	return response, nil
}

func (s *DepositService) receivingAccounts(ctx context.Context, unit int, accountType int, providerCode string) ([]repopg.ReceivingAccountRecord, error) {
	snapshot, err := s.loadReceivingAccountsSnapshot(ctx)
	if err == nil && len(snapshot) > 0 {
		filtered := filterReceivingAccounts(snapshot, unit, accountType, providerCode)
		if len(filtered) > 0 {
			return filtered, nil
		}
	}

	accounts, err := s.repository.ListActiveReceivingAccounts(ctx)
	if err != nil {
		return nil, err
	}

	return filterReceivingAccounts(accounts, unit, accountType, providerCode), nil
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
		Type:          record.Type,
		Unit:          record.Unit,
		ProviderCode:  record.ProviderCode,
		AccountName:   record.AccountName,
		AccountNumber: record.AccountNumber,
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

func filterReceivingAccounts(accounts []repopg.ReceivingAccountRecord, unit int, accountType int, providerCode string) []repopg.ReceivingAccountRecord {
	filtered := make([]repopg.ReceivingAccountRecord, 0, len(accounts))
	for _, account := range accounts {
		if account.Unit != unit || account.Type != accountType {
			continue
		}

		if providerCode != "" && (account.ProviderCode == nil || !strings.EqualFold(strings.TrimSpace(*account.ProviderCode), providerCode)) {
			continue
		}

		filtered = append(filtered, account)
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
	provider := firstNonEmptyString(account.ProviderCode, "BANK")
	return strings.Join([]string{
		"VIETQR",
		provider,
		firstNonEmptyString(account.AccountNumber, ""),
		amount,
		clientRef,
	}, "|")
}

func buildVietQrImageURL(account repopg.ReceivingAccountRecord, amount string, clientRef string) string {
	bankCode := strings.ToLower(strings.TrimSpace(firstNonEmptyString(account.ProviderCode, "")))
	accountNumber := strings.TrimSpace(firstNonEmptyString(account.AccountNumber, ""))
	if bankCode == "" || accountNumber == "" {
		return ""
	}

	baseURL := fmt.Sprintf(
		"https://img.vietqr.io/image/%s-%s-compact.jpg",
		url.PathEscape(bankCode),
		url.PathEscape(accountNumber),
	)

	query := url.Values{}
	if trimmedAmount := strings.TrimSpace(amount); trimmedAmount != "" {
		query.Set("amount", trimmedAmount)
	}
	if trimmedClientRef := strings.TrimSpace(clientRef); trimmedClientRef != "" {
		query.Set("addInfo", trimmedClientRef)
	}
	if accountName := strings.TrimSpace(firstNonEmptyString(account.AccountName, "")); accountName != "" {
		query.Set("accountName", accountName)
	}

	encoded := query.Encode()
	if encoded == "" {
		return baseURL
	}

	return baseURL + "?" + encoded
}

func (s *DepositService) loadPendingDepositCache(ctx context.Context, userID int64, method deposit.DepositMethod) (deposit.DepositInitResponse, error) {
	if s.redis == nil {
		return deposit.DepositInitResponse{}, fmt.Errorf("redis disabled")
	}

	raw, err := s.redis.Get(ctx, pendingDepositCacheKey(userID, method)).Result()
	if err != nil {
		return deposit.DepositInitResponse{}, err
	}

	var response deposit.DepositInitResponse
	if err := json.Unmarshal([]byte(raw), &response); err != nil {
		return deposit.DepositInitResponse{}, err
	}

	if response.ClientRef == "" {
		return deposit.DepositInitResponse{}, fmt.Errorf("pending deposit cache empty")
	}

	if response.Transaction.Status == 2 || response.Transaction.Status == 3 || response.Transaction.Status == 4 {
		return deposit.DepositInitResponse{}, fmt.Errorf("pending deposit already finished")
	}

	return response, nil
}

func (s *DepositService) savePendingDepositCache(ctx context.Context, userID int64, method deposit.DepositMethod, response deposit.DepositInitResponse) error {
	if s.redis == nil {
		return fmt.Errorf("redis disabled")
	}

	payload, err := json.Marshal(response)
	if err != nil {
		return err
	}

	return s.redis.Set(ctx, pendingDepositCacheKey(userID, method), payload, 10*time.Minute).Err()
}

func (s *DepositService) clearPendingDepositCache(ctx context.Context, userID int64, method deposit.DepositMethod) error {
	if s.redis == nil {
		return nil
	}

	return s.redis.Del(ctx, pendingDepositCacheKey(userID, method)).Err()
}

func pendingDepositCacheKey(userID int64, method deposit.DepositMethod) string {
	return fmt.Sprintf("deposit:pending:user:%d:method:%s", userID, method)
}

func depositProviderToMethod(provider string) deposit.DepositMethod {
	switch strings.TrimSpace(strings.ToLower(provider)) {
	case string(deposit.DepositProviderSepayVietQR):
		return deposit.DepositMethodVietQR
	case string(deposit.DepositProviderNowPayments):
		return deposit.DepositMethodUSDT
	default:
		return deposit.DepositMethodVietQR
	}
}

func firstNonNilString(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(*value)
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

func stringPtrOrNil(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
