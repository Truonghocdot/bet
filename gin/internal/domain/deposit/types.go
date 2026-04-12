package deposit

import "time"

type ReceivingAccountType int

const (
	ReceivingAccountTypeBank ReceivingAccountType = 1
)

type DepositMethod string

const (
	DepositMethodVietQR DepositMethod = "vietqr"
	DepositMethodUSDT   DepositMethod = "usdt"
)

type DepositProvider string

const (
	DepositProviderSepayVietQR DepositProvider = "sepay_vietqr"
	DepositProviderNowPayments DepositProvider = "nowpayments_usdt"
)

type ReceivingAccount struct {
	ID            int64   `json:"id"`
	Type          int     `json:"type"`
	Unit          int     `json:"unit"`
	ProviderCode  *string `json:"provider_code,omitempty"`
	AccountName   *string `json:"account_name,omitempty"`
	AccountNumber *string `json:"account_number,omitempty"`
	Status        int     `json:"status"`
	IsDefault     bool    `json:"is_default"`
	SortOrder     int     `json:"sort_order"`
}

type DepositInitRequest struct {
	Amount       string `json:"amount"`
	Note         string `json:"note,omitempty"`
	ProviderCode string `json:"provider_code,omitempty"`
}

type DepositTransaction struct {
	ID               int64             `json:"id"`
	ClientRef        string            `json:"client_ref"`
	Provider         string            `json:"provider"`
	ProviderTxnID    *string           `json:"provider_txn_id,omitempty"`
	ReceivingAccount *ReceivingAccount `json:"receiving_account,omitempty"`
	Unit             int               `json:"unit"`
	Type             int               `json:"type"`
	Amount           string            `json:"amount"`
	NetAmount        string            `json:"net_amount"`
	Status           int               `json:"status"`
	Meta             map[string]any    `json:"meta,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
	ApprovedAt       *time.Time        `json:"approved_at,omitempty"`
}

type DepositInitResponse struct {
	Message          string             `json:"message"`
	Provider         string             `json:"provider"`
	Method           DepositMethod      `json:"method"`
	ClientRef        string             `json:"client_ref"`
	Amount           string             `json:"amount"`
	Transaction      DepositTransaction `json:"transaction"`
	Instructions     string             `json:"instructions,omitempty"`
	QRContent        string             `json:"qr_content,omitempty"`
	QRCodeURL        string             `json:"qr_code_url,omitempty"`
	PayURL           string             `json:"pay_url,omitempty"`
	ExpiresAt        time.Time          `json:"expires_at"`
	ReceivingAccount *ReceivingAccount  `json:"receiving_account,omitempty"`
}

type DepositStatusResponse struct {
	Message          string             `json:"message"`
	Transaction      DepositTransaction `json:"transaction"`
	ReceivingAccount *ReceivingAccount  `json:"receiving_account,omitempty"`
}

type DepositBankOption struct {
	ProviderCode string `json:"provider_code"`
	ShortName    string `json:"short_name"`
	Name         string `json:"name"`
	Bin          string `json:"bin"`
	Logo         string `json:"logo,omitempty"`
	AccountCount int    `json:"account_count"`
	IsDefault    bool   `json:"is_default"`
}

type DepositBankListResponse struct {
	Message string              `json:"message"`
	Banks   []DepositBankOption `json:"banks"`
}

type ApplyDepositRequest struct {
	Provider       string         `json:"provider"`
	ProviderStatus string         `json:"provider_status"`
	ClientRef      string         `json:"client_ref"`
	ProviderTxnID  string         `json:"provider_txn_id"`
	Amount         string         `json:"amount"`
	Currency       string         `json:"currency"`
	PaidAt         time.Time      `json:"paid_at"`
	Raw            map[string]any `json:"raw"`
}

type ApplyDepositResponse struct {
	Message   string    `json:"message"`
	ClientRef string    `json:"client_ref"`
	Status    string    `json:"status"`
	AppliedAt time.Time `json:"applied_at"`
}
