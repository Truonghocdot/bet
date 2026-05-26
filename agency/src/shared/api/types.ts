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
