package telegram

import "time"

type GroupEvent struct {
	SiteCode   string    `json:"site_code"`
	UpdateID   int64     `json:"update_id"`
	ChatID     int64     `json:"chat_id"`
	ChatType   string    `json:"chat_type"`
	Title      string    `json:"title"`
	Username   string    `json:"username,omitempty"`
	BotStatus  string    `json:"bot_status"`
	OccurredAt time.Time `json:"occurred_at"`
}

type Target struct {
	ID       int64  `json:"id"`
	ChatID   int64  `json:"chat_id"`
	ChatType string `json:"chat_type"`
	Title    string `json:"title"`
	Username string `json:"username,omitempty"`
}

type TargetError struct {
	SiteCode string `json:"site_code"`
	ChatID   int64  `json:"chat_id"`
	Error    string `json:"error"`
}

type LookupRequest struct {
	Provider      string `json:"provider"`
	ProviderTxnID string `json:"provider_txn_id,omitempty"`
	ClientRef     string `json:"client_ref,omitempty"`
}

type LookupResponse struct {
	Matched              bool      `json:"matched"`
	TransactionID        int64     `json:"transaction_id,omitempty"`
	UserID               int64     `json:"user_id,omitempty"`
	UserName             string    `json:"user_name,omitempty"`
	UserPhone            string    `json:"user_phone,omitempty"`
	ClientRef            string    `json:"client_ref,omitempty"`
	ProviderTxnID        string    `json:"provider_txn_id,omitempty"`
	Provider             string    `json:"provider,omitempty"`
	Amount               string    `json:"amount,omitempty"`
	Status               int       `json:"status,omitempty"`
	CreatedAt            time.Time `json:"created_at,omitempty"`
	ReceivingBank        string    `json:"receiving_bank,omitempty"`
	ReceivingAccountName string    `json:"receiving_account_name,omitempty"`
	ReceivingAccount     string    `json:"receiving_account,omitempty"`
}
