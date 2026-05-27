export type AuthUser = {
  id: number
  name: string
  email?: string | null
  phone?: string | null
  role: number
  status: number
  created_at: string
  updated_at: string
}

export type AffiliateProfile = {
  id: number
  ref_code: string
  ref_link: string
  status: number
}

export type AuthResponse = {
  user: AuthUser
  affiliate_profile?: AffiliateProfile | null
  access_token: string
  refresh_token?: string
  token_type: string
  expires_in: number
  refresh_expires_in?: number
}

export type ManagedAffiliateUser = {
  user_id: number
  name: string
  phone: string
  created_at: string
  referral_status: number
  first_deposit_amount: string
  first_deposit_transaction_id: number
  first_deposit_transaction_no: string
}

export type ManagedAffiliateUserTransaction = {
  id: number
  unit: number
  direction: number
  amount: string
  balance_before: string
  balance_after: string
  reference_type: string
  reference_id?: number | null
  note: string
  created_at: string
}

export type AgencyManagedUserDeposit = {
  id: number
  user_id: number
  user_name: string
  user_phone: string
  client_ref: string
  provider: string
  provider_txn_id?: string | null
  unit: number
  type: number
  amount: string
  net_amount: string
  status: number
  meta?: Record<string, unknown> | null
  created_at: string
  updated_at: string
  approved_at?: string | null
  receiving_account?: {
    id: number
    type: number
    unit: number
    provider_code?: string | null
    account_name?: string | null
    account_number?: string | null
    status: number
    is_default: boolean
    sort_order: number
  } | null
}

export type AgencyManagedUserDepositHistoryResponse = {
  page: number
  page_size: number
  total: number
  total_pages: number
  data: AgencyManagedUserDeposit[]
}

export type AgencyManagedUserWithdrawal = {
  id: number
  user_id: number
  user_name: string
  user_phone: string
  unit: number
  amount: string
  fee: string
  net_amount: string
  status: number
  reason_rejected?: string
  account_withdrawal_info_id: number
  account_name: string
  account_number: string
  provider_code: string
  created_at: string
}

export type AgencyManagedUserWithdrawalHistoryResponse = {
  page: number
  page_size: number
  total: number
  total_pages: number
  data: AgencyManagedUserWithdrawal[]
}
