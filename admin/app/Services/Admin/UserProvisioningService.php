<?php

namespace App\Services\Admin;

use App\Enum\Affiliate\AffiliateProfileStatus;
use App\Enum\Wallet\UnitTransaction;
use App\Enum\Wallet\WalletStatus;
use App\Models\Affiliate\AffiliateProfile;
use App\Models\Transaction\AccountWithdrawalInfo;
use App\Models\User;
use App\Models\Wallet\Wallet;
use Illuminate\Support\Arr;
use Illuminate\Support\Facades\DB;

class UserProvisioningService
{
    public function createFromErp(array $data, ?User $actor = null): User
    {
        return DB::transaction(function () use ($data): User {
            $user = User::query()->create($this->userPayload($data));

            $this->applyBackfillFields($user, $data);
            $this->provisionWallets($user, $data);
            $this->provisionAffiliateProfile($user, $data);
            $this->provisionAccountWithdrawalInfo($user, $data);

            return $user->fresh([
                'wallets',
                'affiliateProfile',
                'accountWithdrawalInfos',
            ]);
        });
    }

    private function userPayload(array $data): array
    {
        return Arr::only($data, [
            'name',
            'email',
            'phone',
            'password',
            'role',
            'status',
        ]);
    }

    private function applyBackfillFields(User $user, array $data): void
    {
        $backfill = Arr::only($data, [
            'email_verified_at',
            'phone_verified_at',
            'last_login_at',
        ]);

        $payload = [];

        foreach ($backfill as $key => $value) {
            if (filled($value)) {
                $payload[$key] = $value;
            }
        }

        if ($payload === []) {
            return;
        }

        $user->forceFill($payload)->save();
    }

    private function provisionWallets(User $user, array $data): void
    {
        if (! $this->bool($data, 'provision_wallets', true)) {
            return;
        }

        $walletDefinitions = [
            UnitTransaction::VND->value => $this->bool($data, 'provision_vnd_wallet', true),
            UnitTransaction::USDT->value => $this->bool($data, 'provision_usdt_wallet', true),
        ];

        foreach ($walletDefinitions as $unit => $enabled) {
            if (! $enabled) {
                continue;
            }

            Wallet::query()->firstOrCreate(
                [
                    'user_id' => $user->id,
                    'unit' => $unit,
                ],
                [
                    'balance' => 0,
                    'locked_balance' => 0,
                    'status' => WalletStatus::ACTIVE->value,
                ],
            );
        }
    }

    private function provisionAffiliateProfile(User $user, array $data): void
    {
        if (! $this->bool($data, 'provision_affiliate_profile', false)) {
            return;
        }

        if ($user->affiliateProfile()->exists()) {
            return;
        }

        $status = $this->affiliateStatus($data);
        $identity = AffiliateProfile::generateReferralIdentity();

        AffiliateProfile::query()->create([
            'user_id' => $user->id,
            'ref_code' => $identity['ref_code'],
            'ref_link' => $identity['ref_link'],
            'status' => $status,
        ]);
    }

    private function provisionAccountWithdrawalInfo(User $user, array $data): void
    {
        if (! $this->bool($data, 'provision_account_withdrawal_info', false)) {
            return;
        }

        AccountWithdrawalInfo::query()->create([
            'user_id' => $user->id,
            'unit' => $this->accountWithdrawalUnit($data),
            'provider_code' => trim((string) ($data['withdrawal_provider_code'] ?? '')),
            'account_name' => trim((string) ($data['withdrawal_account_name'] ?? '')),
            'account_number' => trim((string) ($data['withdrawal_account_number'] ?? '')),
            'is_default' => $this->bool($data, 'withdrawal_is_default', false),
        ]);
    }

    private function affiliateStatus(array $data): int
    {
        $status = $data['affiliate_status'] ?? AffiliateProfileStatus::PENDING->value;

        if ($status instanceof AffiliateProfileStatus) {
            return $status->value;
        }

        return (int) $status;
    }

    private function accountWithdrawalUnit(array $data): int
    {
        $unit = $data['withdrawal_unit'] ?? UnitTransaction::VND->value;

        if ($unit instanceof UnitTransaction) {
            return $unit->value;
        }

        return (int) $unit;
    }
    private function bool(array $data, string $key, bool $default = false): bool
    {
        if (! array_key_exists($key, $data)) {
            return $default;
        }

        return filter_var($data[$key], FILTER_VALIDATE_BOOL, FILTER_NULL_ON_FAILURE) ?? $default;
    }
}
