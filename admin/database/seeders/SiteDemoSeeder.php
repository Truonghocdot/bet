<?php

namespace Database\Seeders;

use App\Enum\Affiliate\AffiliateLinkStatus;
use App\Enum\Affiliate\AffiliateProfileStatus;
use App\Enum\Affiliate\AffiliateReferralStatus;
use App\Enum\Affiliate\AffiliateRewardStatus;
use App\Enum\Bet\BetItemResult;
use App\Enum\Bet\BetOptionType;
use App\Enum\Bet\BetStatus;
use App\Enum\Bet\BetTicketType;
use App\Enum\Bet\DrawSource;
use App\Enum\Bet\GameType;
use App\Enum\Bet\PeriodStatus;
use App\Enum\Bet\SettlementType;
use App\Enum\Payment\PaymentReceivingAccountStatus;
use App\Enum\Payment\PaymentReceivingAccountType;
use App\Enum\Transaction\TransactionStatus;
use App\Enum\Transaction\TypeTransaction;
use App\Enum\Transaction\WithdrawalStatus;
use App\Enum\User\RoleUser;
use App\Enum\User\UserStatus;
use App\Enum\Wallet\LedgerDirection;
use App\Enum\Wallet\UnitTransaction;
use App\Enum\Wallet\WalletStatus;
use App\Models\Affiliate\AffiliateLink;
use App\Models\Affiliate\AffiliateProfile;
use App\Models\Affiliate\AffiliateReferral;
use App\Models\Affiliate\AffiliateRewardLog;
use App\Models\Affiliate\AffiliateRewardSetting;
use App\Models\Bet\BetItem;
use App\Models\Bet\BetSettlement;
use App\Models\Bet\BetTicket;
use App\Models\Bet\GamePeriod;
use App\Models\Payment\PaymentReceivingAccount;
use App\Models\Payment\VietQrBank;
use App\Models\Transaction\AccountWithdrawalInfo;
use App\Models\Transaction\Transaction;
use App\Models\Transaction\WithdrawalRequest;
use App\Models\User;
use App\Models\Wallet\Wallet;
use App\Models\Wallet\WalletLedgerEntry;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;

class SiteDemoSeeder extends Seeder
{
    public function run(): void
    {
        DB::transaction(function (): void {
            $admin = $this->upsertUser([
                'email' => 'admin@ff789.club',
            ], [
                'name' => 'Administrator',
                'phone' => '0901000001',
                'password' => Hash::make('password'),
                'role' => RoleUser::ADMIN,
                'status' => UserStatus::ACTIVE,
                'email_verified_at' => now(),
                'phone_verified_at' => now(),
            ]);

            $staff = $this->upsertUser([
                'email' => 'ops@ff789.club',
            ], [
                'name' => 'Nhân sự vận hành',
                'phone' => '0901000002',
                'password' => Hash::make('password'),
                'role' => RoleUser::STAFF,
                'status' => UserStatus::ACTIVE,
                'email_verified_at' => now(),
                'phone_verified_at' => now(),
            ]);

            $agency = $this->upsertUser([
                'email' => 'agency@ff789.club',
            ], [
                'name' => 'Agency Demo',
                'phone' => '0901000003',
                'password' => Hash::make('password'),
                'role' => RoleUser::AGENCY,
                'status' => UserStatus::ACTIVE,
                'email_verified_at' => now(),
                'phone_verified_at' => now(),
            ]);

            $alpha = $this->upsertUser(['email' => 'player.alpha@ff789.club'], [
                'name' => 'Player Alpha',
                'phone' => '0901000101',
                'password' => Hash::make('password'),
                'role' => RoleUser::CLIENT,
                'status' => UserStatus::ACTIVE,
                'email_verified_at' => now(),
            ]);

            $beta = $this->upsertUser(['email' => 'player.beta@ff789.club'], [
                'name' => 'Player Beta',
                'phone' => '0901000102',
                'password' => Hash::make('password'),
                'role' => RoleUser::CLIENT,
                'status' => UserStatus::ACTIVE,
                'email_verified_at' => now(),
            ]);

            $gamma = $this->upsertUser(['email' => 'player.gamma@ff789.club'], [
                'name' => 'Player Gamma',
                'phone' => '0901000103',
                'password' => Hash::make('password'),
                'role' => RoleUser::CLIENT,
                'status' => UserStatus::ACTIVE,
                'email_verified_at' => now(),
            ]);

            $delta = $this->upsertUser(['email' => 'player.delta@ff789.club'], [
                'name' => 'Player Delta',
                'phone' => '0901000104',
                'password' => Hash::make('password'),
                'role' => RoleUser::CLIENT,
                'status' => UserStatus::ACTIVE,
                'email_verified_at' => now(),
            ]);

            $echo = $this->upsertUser(['email' => 'player.echo@ff789.club'], [
                'name' => 'Player Echo',
                'phone' => '0901000105',
                'password' => Hash::make('password'),
                'role' => RoleUser::CLIENT,
                'status' => UserStatus::SUSPENDED,
                'email_verified_at' => now(),
            ]);

            $rewardSetting3 = AffiliateRewardSetting::query()->updateOrCreate(
                ['required_qualified_referrals' => 3],
                [
                    'name' => 'Mốc 3 người hợp lệ',
                    'reward_amount' => 50000,
                    'unit' => UnitTransaction::VND,
                    'is_active' => true,
                    'effective_from' => now()->subMonths(2),
                    'note' => 'Seeder: thưởng 50.000 VND khi đạt 3 người hợp lệ.',
                ],
            );

            AffiliateRewardSetting::query()->updateOrCreate(
                ['required_qualified_referrals' => 5],
                [
                    'name' => 'Mốc 5 người hợp lệ',
                    'reward_amount' => 80000,
                    'unit' => UnitTransaction::VND,
                    'is_active' => true,
                    'effective_from' => now()->subMonths(2),
                    'note' => 'Seeder: thưởng 80.000 VND khi đạt 5 người hợp lệ.',
                ],
            );

            $bank = VietQrBank::query()
                ->where('transfer_supported', true)
                ->orderBy('short_name')
                ->first();

            PaymentReceivingAccount::query()->updateOrCreate(
                ['code' => 'SEED-BANK-VND'],
                [
                    'name' => 'Tài khoản nạp ngân hàng chính',
                    'type' => PaymentReceivingAccountType::BANK,
                    'unit' => UnitTransaction::VND,
                    'provider_code' => $bank?->code ?? 'VCB',
                    'account_name' => 'CONG TY FF789',
                    'account_number' => '1900100008888',
                    'wallet_address' => null,
                    'network' => null,
                    'qr_code_path' => null,
                    'instructions' => 'Chuyển khoản đúng nội dung để hệ thống đối soát nhanh.',
                    'status' => PaymentReceivingAccountStatus::ACTIVE,
                    'is_default' => true,
                    'sort_order' => 1,
                ],
            );

            PaymentReceivingAccount::query()->updateOrCreate(
                ['code' => 'SEED-USDT-TRC20'],
                [
                    'name' => 'Ví nạp USDT TRC20',
                    'type' => PaymentReceivingAccountType::CRYPTO,
                    'unit' => UnitTransaction::USDT,
                    'provider_code' => 'TRC20',
                    'account_name' => 'FF789 Treasury',
                    'account_number' => null,
                    'wallet_address' => 'TQ8XDemoWalletAddress000000000001',
                    'network' => 'TRC20',
                    'qr_code_path' => null,
                    'instructions' => 'Chỉ chuyển đúng mạng TRC20.',
                    'status' => PaymentReceivingAccountStatus::ACTIVE,
                    'is_default' => true,
                    'sort_order' => 2,
                ],
            );

            $agencyVndWallet = $this->upsertWallet($agency, UnitTransaction::VND, 1250000, 0);
            $this->upsertWallet($agency, UnitTransaction::USDT, 120.5, 0);

            $alphaVndWallet = $this->upsertWallet($alpha, UnitTransaction::VND, 420000, 0);
            $alphaUsdtWallet = $this->upsertWallet($alpha, UnitTransaction::USDT, 15.75, 0);
            $betaVndWallet = $this->upsertWallet($beta, UnitTransaction::VND, 850000, 250000);
            $betaUsdtWallet = $this->upsertWallet($beta, UnitTransaction::USDT, 42.5, 0);
            $gammaVndWallet = $this->upsertWallet($gamma, UnitTransaction::VND, 110000, 0);
            $deltaVndWallet = $this->upsertWallet($delta, UnitTransaction::VND, 95000, 0);
            $echoVndWallet = $this->upsertWallet($echo, UnitTransaction::VND, 50000, 0);

            $alphaVndInfo = $this->upsertAccountWithdrawalInfo($alpha, [
                'unit' => UnitTransaction::VND,
                'provider_code' => $bank?->code ?? 'VCB',
                'account_name' => 'PLAYER ALPHA',
                'account_number' => '0933000001',
                'is_default' => true,
            ]);

            $betaUsdtInfo = $this->upsertAccountWithdrawalInfo($beta, [
                'unit' => UnitTransaction::USDT,
                'provider_code' => 'TRC20',
                'account_name' => 'PLAYER BETA',
                'account_number' => 'TQ8XPlayerBetaWallet0001',
                'is_default' => true,
            ]);

            $betaVndInfo = $this->upsertAccountWithdrawalInfo($beta, [
                'unit' => UnitTransaction::VND,
                'provider_code' => $bank?->code ?? 'VCB',
                'account_name' => 'PLAYER BETA',
                'account_number' => '0933000002',
                'is_default' => false,
            ]);

            $gammaVndInfo = $this->upsertAccountWithdrawalInfo($gamma, [
                'unit' => UnitTransaction::VND,
                'provider_code' => $bank?->code ?? 'VCB',
                'account_name' => 'PLAYER GAMMA',
                'account_number' => '0933000003',
                'is_default' => true,
            ]);

            $agencyProfile = $this->upsertAffiliateProfile($agency, AffiliateProfileStatus::ACTIVE);
            $alphaProfile = $this->upsertAffiliateProfile($alpha, AffiliateProfileStatus::ACTIVE);
            $betaProfile = $this->upsertAffiliateProfile($beta, AffiliateProfileStatus::ACTIVE);
            $gammaProfile = $this->upsertAffiliateProfile($gamma, AffiliateProfileStatus::ACTIVE);
            $this->upsertAffiliateProfile($delta, AffiliateProfileStatus::PENDING);
            $this->upsertAffiliateProfile($echo, AffiliateProfileStatus::SUSPENDED);

            $campaignLink = AffiliateLink::query()->updateOrCreate(
                ['tracking_code' => 'seed-agency-main'],
                [
                    'affiliate_profile_id' => $agencyProfile->id,
                    'campaign_name' => 'Agency Main Campaign',
                    'landing_url' => rtrim((string) config('app.url'), '/').'/register?source=agency-main',
                    'status' => AffiliateLinkStatus::ACTIVE,
                ],
            );

            $alphaDeposit = $this->upsertTransaction('seed-txn-alpha-deposit', [
                'user_id' => $alpha->id,
                'wallet_id' => $alphaVndWallet->id,
                'unit' => UnitTransaction::VND,
                'type' => TypeTransaction::DEPOSIT,
                'amount' => 500000,
                'fee' => 0,
                'net_amount' => 500000,
                'status' => TransactionStatus::COMPLETED,
                'provider' => 'vietqr',
                'provider_txn_id' => 'seed-txn-alpha-deposit',
                'approved_by' => $staff->id,
                'approved_at' => now()->subDays(7),
            ]);

            $betaDeposit = $this->upsertTransaction('seed-txn-beta-deposit', [
                'user_id' => $beta->id,
                'wallet_id' => $betaVndWallet->id,
                'unit' => UnitTransaction::VND,
                'type' => TypeTransaction::DEPOSIT,
                'amount' => 1000000,
                'fee' => 0,
                'net_amount' => 1000000,
                'status' => TransactionStatus::COMPLETED,
                'provider' => 'vietqr',
                'provider_txn_id' => 'seed-txn-beta-deposit',
                'approved_by' => $staff->id,
                'approved_at' => now()->subDays(6),
            ]);

            $gammaDeposit = $this->upsertTransaction('seed-txn-gamma-deposit', [
                'user_id' => $gamma->id,
                'wallet_id' => $gammaVndWallet->id,
                'unit' => UnitTransaction::VND,
                'type' => TypeTransaction::DEPOSIT,
                'amount' => 80000,
                'fee' => 0,
                'net_amount' => 80000,
                'status' => TransactionStatus::COMPLETED,
                'provider' => 'vietqr',
                'provider_txn_id' => 'seed-txn-gamma-deposit',
                'approved_by' => $staff->id,
                'approved_at' => now()->subDays(5),
            ]);

            $deltaDeposit = $this->upsertTransaction('seed-txn-delta-deposit', [
                'user_id' => $delta->id,
                'wallet_id' => $deltaVndWallet->id,
                'unit' => UnitTransaction::VND,
                'type' => TypeTransaction::DEPOSIT,
                'amount' => 60000,
                'fee' => 0,
                'net_amount' => 60000,
                'status' => TransactionStatus::CONFIRMED,
                'provider' => 'vietqr',
                'provider_txn_id' => 'seed-txn-delta-deposit',
                'approved_by' => $staff->id,
                'approved_at' => now()->subDays(4),
            ]);

            $this->upsertTransaction('seed-txn-echo-failed', [
                'user_id' => $echo->id,
                'wallet_id' => $echoVndWallet->id,
                'unit' => UnitTransaction::VND,
                'type' => TypeTransaction::DEPOSIT,
                'amount' => 200000,
                'fee' => 0,
                'net_amount' => 200000,
                'status' => TransactionStatus::FAILED,
                'provider' => 'vietqr',
                'provider_txn_id' => 'seed-txn-echo-failed',
                'reason_failed' => 'Seeder: giao dịch bị từ chối từ cổng thanh toán.',
            ]);

            $this->upsertLedgerEntry(
                [
                    'wallet_id' => $alphaVndWallet->id,
                    'user_id' => $alpha->id,
                    'direction' => LedgerDirection::CREDIT,
                    'amount' => 500000,
                    'balance_before' => 0,
                    'balance_after' => 500000,
                    'reference_type' => Transaction::class,
                    'reference_id' => $alphaDeposit->id,
                    'note' => 'Seeder: nạp tiền Alpha',
                    'created_at' => now()->subDays(7),
                ],
            );

            $this->upsertLedgerEntry(
                [
                    'wallet_id' => $betaVndWallet->id,
                    'user_id' => $beta->id,
                    'direction' => LedgerDirection::CREDIT,
                    'amount' => 1000000,
                    'balance_before' => 0,
                    'balance_after' => 1000000,
                    'reference_type' => Transaction::class,
                    'reference_id' => $betaDeposit->id,
                    'note' => 'Seeder: nạp tiền Beta',
                    'created_at' => now()->subDays(6),
                ],
            );

            $this->upsertLedgerEntry(
                [
                    'wallet_id' => $gammaVndWallet->id,
                    'user_id' => $gamma->id,
                    'direction' => LedgerDirection::CREDIT,
                    'amount' => 80000,
                    'balance_before' => 0,
                    'balance_after' => 80000,
                    'reference_type' => Transaction::class,
                    'reference_id' => $gammaDeposit->id,
                    'note' => 'Seeder: nạp tiền Gamma',
                    'created_at' => now()->subDays(5),
                ],
            );

            AffiliateReferral::query()->updateOrCreate(
                ['referred_user_id' => $alpha->id],
                [
                    'affiliate_profile_id' => $agencyProfile->id,
                    'referrer_user_id' => $agency->id,
                    'affiliate_link_id' => $campaignLink->id,
                    'first_deposit_transaction_id' => $alphaDeposit->id,
                    'first_deposit_amount' => 500000,
                    'qualified_at' => now()->subDays(7),
                    'status' => AffiliateReferralStatus::QUALIFIED,
                ],
            );

            AffiliateReferral::query()->updateOrCreate(
                ['referred_user_id' => $beta->id],
                [
                    'affiliate_profile_id' => $agencyProfile->id,
                    'referrer_user_id' => $agency->id,
                    'affiliate_link_id' => $campaignLink->id,
                    'first_deposit_transaction_id' => $betaDeposit->id,
                    'first_deposit_amount' => 1000000,
                    'qualified_at' => now()->subDays(6),
                    'status' => AffiliateReferralStatus::QUALIFIED,
                ],
            );

            AffiliateReferral::query()->updateOrCreate(
                ['referred_user_id' => $gamma->id],
                [
                    'affiliate_profile_id' => $agencyProfile->id,
                    'referrer_user_id' => $agency->id,
                    'affiliate_link_id' => $campaignLink->id,
                    'first_deposit_transaction_id' => $gammaDeposit->id,
                    'first_deposit_amount' => 80000,
                    'qualified_at' => now()->subDays(5),
                    'status' => AffiliateReferralStatus::QUALIFIED,
                ],
            );

            AffiliateReferral::query()->updateOrCreate(
                ['referred_user_id' => $delta->id],
                [
                    'affiliate_profile_id' => $agencyProfile->id,
                    'referrer_user_id' => $agency->id,
                    'affiliate_link_id' => $campaignLink->id,
                    'first_deposit_transaction_id' => $deltaDeposit->id,
                    'first_deposit_amount' => 60000,
                    'qualified_at' => null,
                    'status' => AffiliateReferralStatus::PENDING,
                ],
            );

            AffiliateReferral::query()->updateOrCreate(
                ['referred_user_id' => $echo->id],
                [
                    'affiliate_profile_id' => $agencyProfile->id,
                    'referrer_user_id' => $agency->id,
                    'affiliate_link_id' => $campaignLink->id,
                    'first_deposit_transaction_id' => null,
                    'first_deposit_amount' => null,
                    'qualified_at' => null,
                    'status' => AffiliateReferralStatus::INVALID,
                ],
            );

            $agencyRewardLedger = $this->upsertLedgerEntry(
                [
                    'wallet_id' => $agencyVndWallet->id,
                    'user_id' => $agency->id,
                    'direction' => LedgerDirection::CREDIT,
                    'amount' => 50000,
                    'balance_before' => 1200000,
                    'balance_after' => 1250000,
                    'reference_type' => AffiliateRewardSetting::class,
                    'reference_id' => $rewardSetting3->id,
                    'note' => 'Seeder: thưởng affiliate mốc 3',
                    'created_at' => now()->subDays(3),
                ],
            );

            AffiliateRewardLog::query()->updateOrCreate(
                [
                    'affiliate_profile_id' => $agencyProfile->id,
                    'setting_id' => $rewardSetting3->id,
                ],
                [
                    'referrer_user_id' => $agency->id,
                    'required_qualified_referrals' => 3,
                    'actual_qualified_referrals' => 3,
                    'reward_amount' => 50000,
                    'unit' => UnitTransaction::VND,
                    'status' => AffiliateRewardStatus::PAID,
                    'wallet_ledger_entry_id' => $agencyRewardLedger->id,
                    'granted_at' => now()->subDays(3),
                ],
            );

            $paidWithdrawal = WithdrawalRequest::query()->updateOrCreate(
                ['admin_note' => 'seed-withdraw-alpha-paid'],
                [
                    'user_id' => $alpha->id,
                    'wallet_id' => $alphaVndWallet->id,
                    'account_withdrawal_info_id' => $alphaVndInfo->id,
                    'unit' => UnitTransaction::VND,
                    'amount' => 120000,
                    'fee' => 0,
                    'net_amount' => 120000,
                    'status' => WithdrawalStatus::PAID,
                    'reviewed_by' => $staff->id,
                    'reviewed_at' => now()->subDays(2),
                    'paid_by' => $admin->id,
                    'paid_at' => now()->subDays(2),
                    'transfer_reference' => 'seed-withdraw-alpha-paid',
                    'admin_note' => 'seed-withdraw-alpha-paid',
                ],
            );

            WithdrawalRequest::query()->updateOrCreate(
                ['admin_note' => 'seed-withdraw-beta-pending'],
                [
                    'user_id' => $beta->id,
                    'wallet_id' => $betaVndWallet->id,
                    'account_withdrawal_info_id' => $betaVndInfo->id,
                    'unit' => UnitTransaction::VND,
                    'amount' => 250000,
                    'fee' => 0,
                    'net_amount' => 250000,
                    'status' => WithdrawalStatus::PENDING,
                    'admin_note' => 'seed-withdraw-beta-pending',
                ],
            );

            WithdrawalRequest::query()->updateOrCreate(
                ['admin_note' => 'seed-withdraw-gamma-rejected'],
                [
                    'user_id' => $gamma->id,
                    'wallet_id' => $gammaVndWallet->id,
                    'account_withdrawal_info_id' => $gammaVndInfo->id,
                    'unit' => UnitTransaction::VND,
                    'amount' => 50000,
                    'fee' => 0,
                    'net_amount' => 50000,
                    'status' => WithdrawalStatus::REJECTED,
                    'reason_rejected' => 'Seeder: sai thông tin tài khoản nhận tiền.',
                    'reviewed_by' => $staff->id,
                    'reviewed_at' => now()->subDay(),
                    'admin_note' => 'seed-withdraw-gamma-rejected',
                ],
            );

            $this->upsertLedgerEntry(
                [
                    'wallet_id' => $alphaVndWallet->id,
                    'user_id' => $alpha->id,
                    'direction' => LedgerDirection::DEBIT,
                    'amount' => 120000,
                    'balance_before' => 540000,
                    'balance_after' => 420000,
                    'reference_type' => WithdrawalRequest::class,
                    'reference_id' => $paidWithdrawal->id,
                    'note' => 'Seeder: rút tiền Alpha',
                    'created_at' => now()->subDays(2),
                ],
            );

            $wingoPeriod = GamePeriod::query()->updateOrCreate(
                ['period_no' => 'WINGO-20260410-001'],
                [
                    'game_type' => GameType::WINGO,
                    'room_code' => 'A1',
                    'open_at' => now()->subHours(3),
                    'close_at' => now()->subHours(2)->subMinutes(59),
                    'draw_at' => now()->subHours(2)->subMinutes(58),
                    'settled_at' => now()->subHours(2)->subMinutes(55),
                    'status' => PeriodStatus::SETTLED,
                    'draw_source' => DrawSource::AUTO,
                    'result_payload' => [
                        'number' => 7,
                        'big_small' => 'BIG',
                        'odd_even' => 'ODD',
                        'color' => 'GREEN',
                    ],
                    'result_hash' => 'seed-wingo-result-001',
                ],
            );

            $k3Period = GamePeriod::query()->updateOrCreate(
                ['period_no' => 'K3-20260410-014'],
                [
                    'game_type' => GameType::K3,
                    'room_code' => 'K3-1',
                    'open_at' => now()->subHour(),
                    'close_at' => now()->subMinutes(55),
                    'draw_at' => now()->subMinutes(54),
                    'settled_at' => null,
                    'status' => PeriodStatus::DRAWN,
                    'draw_source' => DrawSource::AUTO,
                    'result_payload' => [
                        'dice' => [2, 4, 6],
                        'sum' => 12,
                    ],
                    'result_hash' => 'seed-k3-result-014',
                ],
            );

            $lotteryPeriod = GamePeriod::query()->updateOrCreate(
                ['period_no' => 'LOTTERY-20260410-090'],
                [
                    'game_type' => GameType::LOTTERY,
                    'room_code' => 'L1',
                    'open_at' => now()->addMinutes(15),
                    'close_at' => now()->addMinutes(20),
                    'draw_at' => now()->addMinutes(21),
                    'settled_at' => null,
                    'status' => PeriodStatus::OPEN,
                    'draw_source' => DrawSource::MANUAL,
                    'result_payload' => null,
                    'result_hash' => null,
                ],
            );

            $alphaTicket = BetTicket::query()->updateOrCreate(
                ['ticket_no' => 'SEED-TICKET-ALPHA-001'],
                [
                    'user_id' => $alpha->id,
                    'wallet_id' => $alphaVndWallet->id,
                    'unit' => UnitTransaction::VND,
                    'game_type' => GameType::WINGO,
                    'period_id' => $wingoPeriod->id,
                    'bet_type' => BetTicketType::SINGLE,
                    'stake' => 50000,
                    'total_odds' => 1.95,
                    'potential_payout' => 97500,
                    'actual_payout' => 97500,
                    'status' => BetStatus::WON,
                    'placed_ip' => '127.0.0.10',
                    'placed_device' => 'Seeder Device Alpha',
                    'settled_at' => now()->subHours(2)->subMinutes(55),
                ],
            );

            BetItem::query()->updateOrCreate(
                [
                    'ticket_id' => $alphaTicket->id,
                    'option_type' => BetOptionType::BIG_SMALL,
                    'option_key' => 'BIG',
                ],
                [
                    'period_id' => $wingoPeriod->id,
                    'option_label' => 'Lớn',
                    'odds_at_placement' => 1.95,
                    'stake' => 50000,
                    'result' => BetItemResult::WON,
                    'payout_amount' => 97500,
                    'result_payload' => ['winning_key' => 'BIG'],
                    'settled_at' => now()->subHours(2)->subMinutes(55),
                ],
            );

            BetSettlement::query()->updateOrCreate(
                [
                    'ticket_id' => $alphaTicket->id,
                    'settlement_type' => SettlementType::AUTO,
                ],
                [
                    'period_id' => $wingoPeriod->id,
                    'before_status' => BetStatus::PENDING,
                    'after_status' => BetStatus::WON,
                    'payout_amount' => 97500,
                    'profit_loss' => -47500,
                    'note' => 'Seeder: Alpha thắng cửa lớn Wingo.',
                    'settled_by' => $staff->id,
                    'created_at' => now()->subHours(2)->subMinutes(55),
                ],
            );

            $betaTicket = BetTicket::query()->updateOrCreate(
                ['ticket_no' => 'SEED-TICKET-BETA-001'],
                [
                    'user_id' => $beta->id,
                    'wallet_id' => $betaVndWallet->id,
                    'unit' => UnitTransaction::VND,
                    'game_type' => GameType::WINGO,
                    'period_id' => $wingoPeriod->id,
                    'bet_type' => BetTicketType::SINGLE,
                    'stake' => 80000,
                    'total_odds' => 1.95,
                    'potential_payout' => 156000,
                    'actual_payout' => 0,
                    'status' => BetStatus::LOST,
                    'placed_ip' => '127.0.0.11',
                    'placed_device' => 'Seeder Device Beta',
                    'settled_at' => now()->subHours(2)->subMinutes(55),
                ],
            );

            BetItem::query()->updateOrCreate(
                [
                    'ticket_id' => $betaTicket->id,
                    'option_type' => BetOptionType::ODD_EVEN,
                    'option_key' => 'EVEN',
                ],
                [
                    'period_id' => $wingoPeriod->id,
                    'option_label' => 'Chẵn',
                    'odds_at_placement' => 1.95,
                    'stake' => 80000,
                    'result' => BetItemResult::LOST,
                    'payout_amount' => 0,
                    'result_payload' => ['winning_key' => 'ODD'],
                    'settled_at' => now()->subHours(2)->subMinutes(55),
                ],
            );

            BetSettlement::query()->updateOrCreate(
                [
                    'ticket_id' => $betaTicket->id,
                    'settlement_type' => SettlementType::AUTO,
                ],
                [
                    'period_id' => $wingoPeriod->id,
                    'before_status' => BetStatus::PENDING,
                    'after_status' => BetStatus::LOST,
                    'payout_amount' => 0,
                    'profit_loss' => 80000,
                    'note' => 'Seeder: Beta thua cửa chẵn Wingo.',
                    'settled_by' => $staff->id,
                    'created_at' => now()->subHours(2)->subMinutes(55),
                ],
            );

            $gammaTicket = BetTicket::query()->updateOrCreate(
                ['ticket_no' => 'SEED-TICKET-GAMMA-001'],
                [
                    'user_id' => $gamma->id,
                    'wallet_id' => $gammaVndWallet->id,
                    'unit' => UnitTransaction::VND,
                    'game_type' => GameType::WINGO,
                    'period_id' => $wingoPeriod->id,
                    'bet_type' => BetTicketType::MULTI,
                    'stake' => 30000,
                    'total_odds' => 3.20,
                    'potential_payout' => 96000,
                    'actual_payout' => 96000,
                    'status' => BetStatus::WON,
                    'placed_ip' => '127.0.0.12',
                    'placed_device' => 'Seeder Device Gamma',
                    'settled_at' => now()->subHours(2)->subMinutes(55),
                ],
            );

            BetItem::query()->updateOrCreate(
                [
                    'ticket_id' => $gammaTicket->id,
                    'option_type' => BetOptionType::COLOR,
                    'option_key' => 'GREEN',
                ],
                [
                    'period_id' => $wingoPeriod->id,
                    'option_label' => 'Xanh',
                    'odds_at_placement' => 1.60,
                    'stake' => 15000,
                    'result' => BetItemResult::WON,
                    'payout_amount' => 48000,
                    'result_payload' => ['winning_key' => 'GREEN'],
                    'settled_at' => now()->subHours(2)->subMinutes(55),
                ],
            );

            BetItem::query()->updateOrCreate(
                [
                    'ticket_id' => $gammaTicket->id,
                    'option_type' => BetOptionType::ODD_EVEN,
                    'option_key' => 'ODD',
                ],
                [
                    'period_id' => $wingoPeriod->id,
                    'option_label' => 'Lẻ',
                    'odds_at_placement' => 1.60,
                    'stake' => 15000,
                    'result' => BetItemResult::WON,
                    'payout_amount' => 48000,
                    'result_payload' => ['winning_key' => 'ODD'],
                    'settled_at' => now()->subHours(2)->subMinutes(55),
                ],
            );

            BetSettlement::query()->updateOrCreate(
                [
                    'ticket_id' => $gammaTicket->id,
                    'settlement_type' => SettlementType::AUTO,
                ],
                [
                    'period_id' => $wingoPeriod->id,
                    'before_status' => BetStatus::PENDING,
                    'after_status' => BetStatus::WON,
                    'payout_amount' => 96000,
                    'profit_loss' => -66000,
                    'note' => 'Seeder: Gamma thắng ticket multi Wingo.',
                    'settled_by' => $staff->id,
                    'created_at' => now()->subHours(2)->subMinutes(55),
                ],
            );

            $deltaTicket = BetTicket::query()->updateOrCreate(
                ['ticket_no' => 'SEED-TICKET-DELTA-001'],
                [
                    'user_id' => $delta->id,
                    'wallet_id' => $deltaVndWallet->id,
                    'unit' => UnitTransaction::VND,
                    'game_type' => GameType::K3,
                    'period_id' => $k3Period->id,
                    'bet_type' => BetTicketType::SINGLE,
                    'stake' => 25000,
                    'total_odds' => 2.10,
                    'potential_payout' => 52500,
                    'actual_payout' => 0,
                    'status' => BetStatus::PENDING,
                    'placed_ip' => '127.0.0.13',
                    'placed_device' => 'Seeder Device Delta',
                    'settled_at' => null,
                ],
            );

            BetItem::query()->updateOrCreate(
                [
                    'ticket_id' => $deltaTicket->id,
                    'option_type' => BetOptionType::SUM,
                    'option_key' => '12',
                ],
                [
                    'period_id' => $k3Period->id,
                    'option_label' => 'Tổng 12',
                    'odds_at_placement' => 2.10,
                    'stake' => 25000,
                    'result' => BetItemResult::PENDING,
                    'payout_amount' => 0,
                    'result_payload' => null,
                    'settled_at' => null,
                ],
            );

            GamePeriod::query()->updateOrCreate(
                ['period_no' => $lotteryPeriod->period_no],
                [
                    'game_type' => $lotteryPeriod->game_type,
                    'room_code' => $lotteryPeriod->room_code,
                    'open_at' => $lotteryPeriod->open_at,
                    'close_at' => $lotteryPeriod->close_at,
                    'draw_at' => $lotteryPeriod->draw_at,
                    'settled_at' => $lotteryPeriod->settled_at,
                    'status' => $lotteryPeriod->status,
                    'draw_source' => $lotteryPeriod->draw_source,
                    'result_payload' => $lotteryPeriod->result_payload,
                    'result_hash' => $lotteryPeriod->result_hash,
                ],
            );

            $alphaUsdtWallet->forceFill([
                'balance' => 15.75,
                'locked_balance' => 0,
            ])->save();

            $betaUsdtWallet->forceFill([
                'balance' => 42.5,
                'locked_balance' => 0,
            ])->save();
        });
    }

    private function upsertUser(array $match, array $values): User
    {
        $user = User::query()->updateOrCreate($match, $values);

        $timestamps = array_filter([
            'email_verified_at' => $values['email_verified_at'] ?? null,
            'phone_verified_at' => $values['phone_verified_at'] ?? null,
            'last_login_at' => $values['last_login_at'] ?? null,
        ], static fn ($value) => $value !== null);

        if ($timestamps !== []) {
            $user->forceFill($timestamps)->save();
        }

        return $user->refresh();
    }

    private function upsertWallet(User $user, UnitTransaction $unit, float $balance, float $lockedBalance): Wallet
    {
        return Wallet::query()->updateOrCreate(
            [
                'user_id' => $user->id,
                'unit' => $unit->value,
            ],
            [
                'balance' => $balance,
                'locked_balance' => $lockedBalance,
                'status' => WalletStatus::ACTIVE->value,
            ],
        );
    }

    private function upsertAccountWithdrawalInfo(User $user, array $values): AccountWithdrawalInfo
    {
        return AccountWithdrawalInfo::query()->updateOrCreate(
            [
                'user_id' => $user->id,
                'unit' => $values['unit'] instanceof UnitTransaction ? $values['unit']->value : $values['unit'],
                'account_number' => $values['account_number'],
            ],
            [
                'provider_code' => $values['provider_code'],
                'account_name' => $values['account_name'],
                'is_default' => $values['is_default'] ?? false,
            ],
        );
    }

    private function upsertAffiliateProfile(User $user, AffiliateProfileStatus $status): AffiliateProfile
    {
        $profile = AffiliateProfile::query()->firstOrNew([
            'user_id' => $user->id,
        ]);

        if (! $profile->exists) {
            $identity = AffiliateProfile::generateReferralIdentity();
            $profile->ref_code = $identity['ref_code'];
            $profile->ref_link = $identity['ref_link'];
        }

        $profile->status = $status;
        $profile->save();

        return $profile->refresh();
    }

    private function upsertTransaction(string $providerTxnId, array $values): Transaction
    {
        return Transaction::query()->updateOrCreate(
            ['provider_txn_id' => $providerTxnId],
            $values,
        );
    }

    private function upsertLedgerEntry(array $values): WalletLedgerEntry
    {
        return WalletLedgerEntry::query()->updateOrCreate(
            [
                'reference_type' => $values['reference_type'],
                'reference_id' => $values['reference_id'],
                'note' => $values['note'],
            ],
            $values,
        );
    }
}
