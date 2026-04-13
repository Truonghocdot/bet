package withdrawal

import "time"

type AccountWithdrawalInfo struct {
	ID            int64     `json:"id"`
	Unit          int       `json:"unit"`
	ProviderCode  string    `json:"provider_code"`
	AccountName   string    `json:"account_name"`
	AccountNumber string    `json:"account_number"`
	IsDefault     bool      `json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
}

type WithdrawalRequest struct {
	ID                      int64     `json:"id"`
	Unit                    int       `json:"unit"`
	Amount                  string    `json:"amount"`
	Fee                     string    `json:"fee"`
	NetAmount               string    `json:"net_amount"`
	Status                  int       `json:"status"`
	ReasonRejected          string    `json:"reason_rejected,omitempty"`
	AccountWithdrawalInfoID int64     `json:"account_withdrawal_info_id"`
	AccountName             string    `json:"account_name"`
	AccountNumber           string    `json:"account_number"`
	ProviderCode            string    `json:"provider_code"`
	CreatedAt               time.Time `json:"created_at"`
}

type SetupAccountRequest struct {
	Unit          int    `json:"unit" binding:"required,oneof=1 2"`
	ProviderCode  string `json:"provider_code" binding:"max=50"`
	AccountName   string `json:"account_name" binding:"required,max=255"`
	AccountNumber string `json:"account_number" binding:"required,max=255"`
	IsDefault     bool   `json:"is_default"`
}

type SubmitWithdrawalRequest struct {
	Amount                  string `json:"amount" binding:"required"`
	AccountWithdrawalInfoID int64  `json:"account_withdrawal_info_id" binding:"required"`
}

type DeleteAccountRequest struct {
    ID int64 `uri:"id" binding:"required,min=1"`
}
