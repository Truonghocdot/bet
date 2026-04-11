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

export type PlayRoomItem = {
  code: string
  game_type: string
  duration_seconds: number
  bet_cutoff_seconds: number
  status: string
  sort_order: number
}

export type PlayRoomPeriod = {
  id: number
  period_no: string
  status: string
  open_at: string
  bet_lock_at: string
  draw_at: string
}

export type PlayHistoryItem = {
  period_no: string
  result: string
  big_small: string
  color: string
  draw_at: string
  status: string
  created_at: string
  updated_at: string
}

export type PlayRoomStateResponse = {
  message: string
  server_time: string
  room: PlayRoomItem
  current_period: PlayRoomPeriod
  recent_results: PlayHistoryItem[]
}

export type PlayRoomHistoryResponse = {
  message: string
  page: number
  page_size: number
  total: number
  total_pages: number
  items: PlayHistoryItem[]
}

export type PlayRoomBetHistoryItem = {
  id: number
  period_no: string
  result: string
  big_small: string
  color: string
  stake: string
  original_amount?: string
  tax_amount?: string
  net_amount?: string
  actual_payout: string
  profit_loss: string
  settled_at?: string | null
  status: string
  items_count: number
  created_at: string
}

export type PlayRoomBetHistoryResponse = {
  message: string
  page: number
  page_size: number
  total: number
  total_pages: number
  items: PlayRoomBetHistoryItem[]
}

export type PlayRoomBetRequest = {
  request_id: string
  period_id: string
  items: Array<{
    option_type: string
    option_key: string
    stake: string
  }>
}

export type PlayRoomBetResponse = {
  request_id: string
  room_code: string
  status: string
  accepted_at: string
  message: string
}

export type GameJoinResponse = {
  connection_id: string
  game_type: string
  joined_at: string
  message: string
}

export type GamePlaceBetResponse = {
  request_id: string
  connection_id: string
  game_type: string
  status: string
  accepted_at: string
  message: string
}

export type GameHistoryItem = {
  period_no: string
  result: string
  big_small: string
  color: string
  draw_at: string
  status: string
  created_at: string
  updated_at: string
}

export type GameHistoryResponse = {
  message: string
  page: number
  page_size: number
  total: number
  total_pages: number
  items: GameHistoryItem[]
}

export type GameBetHistoryItem = {
  id: number
  period_no: string
  result: string
  big_small: string
  color: string
  stake: string
  actual_payout: string
  profit_loss: string
  settled_at?: string | null
  status: string
  items_count: number
  created_at: string
}

export type GameBetHistoryResponse = {
  message: string
  page: number
  page_size: number
  total: number
  total_pages: number
  items: GameBetHistoryItem[]
}

export type WalletSummaryItem = {
  id: number
  unit: number
  unit_code: string
  unit_label: string
  balance: string
  locked_balance: string
  status: number
  created_at: string
  updated_at: string
}

export type WalletSummaryResponse = {
  message: string
  wallets: WalletSummaryItem[]
}

export type NotificationListItem = {
  id: number
  title: string
  body: string
  status: number
  audience: number
  publish_at?: string | null
  expires_at?: string | null
  created_at: string
  is_read: boolean
  read_at?: string | null
}

export type NotificationListResponse = {
  message: string
  page: number
  page_size: number
  total: number
  total_pages: number
  items: NotificationListItem[]
}

export type NotificationReadResponse = {
  message: string
  id: number
  read_at: string
}
