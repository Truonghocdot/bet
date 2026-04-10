export type ApiMessage = { message: string }

export type AuthUser = {
  id: number
  name: string
  email: string
  phone?: string | null
  role: number
  status: number
  email_verified_at?: string | null
  phone_verified_at?: string | null
  last_login_at?: string | null
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
  token_type: string
  expires_in: number
}

export type RegisterRequest = {
  name: string
  email: string
  phone?: string
  password: string
  ref_code?: string
}

export type LoginRequest = {
  account: string
  password: string
}

export type ForgotPasswordRequest = {
  channel: 'email' | 'phone'
  account: string
}

export type VerifyForgotOtpRequest = {
  channel: 'email' | 'phone'
  account: string
  otp: string
}

export type VerifyForgotOtpResponse = {
  message: string
  reset_token: string
  expires_in: number
}

export type ResetPasswordRequest = {
  reset_token: string
  new_password: string
}

export type ReceivingAccount = {
  id: number
  code: string
  name: string
  type: number
  unit: number
  provider_code?: string | null
  account_name?: string | null
  account_number?: string | null
  wallet_address?: string | null
  network?: string | null
  qr_code_path?: string | null
  instructions?: string | null
  status: number
  is_default: boolean
  sort_order: number
}

export type DepositTransaction = {
  id: number
  client_ref: string
  provider: string
  provider_txn_id?: string | null
  amount: string
  unit: number
  type: number
  status: number
  receiving_account?: ReceivingAccount | null
  paid_at?: string | null
  created_at?: string | null
  updated_at?: string | null
}

export type DepositInitResponse = {
  message: string
  provider: string
  method: 'vietqr' | 'usdt'
  client_ref: string
  amount: string
  expires_at: string
  instructions?: string | null
  qr_content?: string | null
  qr_code_url?: string | null
  pay_url?: string | null
  receiving_account?: ReceivingAccount | null
  transaction?: DepositTransaction | null
}

export type DepositStatusResponse = {
  message: string
  transaction: DepositTransaction
  receiving_account?: ReceivingAccount | null
}

