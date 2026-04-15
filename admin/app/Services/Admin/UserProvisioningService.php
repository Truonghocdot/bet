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
use Illuminate\Database\UniqueConstraintViolationException;
use Illuminate\Validation\ValidationException;

class UserProvisioningService
{
    public function createFromErp(array $data, ?User $actor = null): User
    {
        $normalizedPhone = $this->normalizePhone($data['phone'] ?? null);
        if ($normalizedPhone !== null && User::query()->where('phone', $normalizedPhone)->exists()) {
            throw ValidationException::withMessages([
                'phone' => 'Số điện thoại này đã được sử dụng.',
            ]);
        }

        return DB::transaction(function () use ($data): User {
            try {
                $user = User::query()->create($this->userPayload($data));
            } catch (UniqueConstraintViolationException) {
                throw ValidationException::withMessages([
                    'phone' => 'Số điện thoại này đã được sử dụng.',
                ]);
            }

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
        $payload = Arr::only($data, [
            'name',
            'phone',
            'password',
            'role',
            'status',
        ]);

        if (array_key_exists('phone', $payload)) {
            $payload['phone'] = $this->normalizePhone($payload['phone']);
        }

        return $payload;
    }

    private function applyBackfillFields(User $user, array $data): void
    {
        $backfill = Arr::only($data, [
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
        $status = $data['affiliate_status'] ?? AffiliateProfileStatus::ACTIVE->value;

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

    private function normalizePhone(mixed $value): ?string
    {
        $raw = trim((string) $value);
        if ($raw === '') {
            return null;
        }

        $raw = preg_replace('/[\s-]+/', '', $raw) ?? $raw;

        if (str_starts_with($raw, '+')) {
            $digits = preg_replace('/\D+/', '', substr($raw, 1)) ?: '';
            if ($digits === '') {
                return null;
            }
            if (str_starts_with($digits, '84')) {
                return '+84' . ltrim(substr($digits, 2), '0');
            }

            return '+' . $digits;
        }

        $digits = preg_replace('/\D+/', '', $raw) ?: '';
        if ($digits === '') {
            return null;
        }

        if (str_starts_with($digits, '84')) {
            return '+84' . ltrim(substr($digits, 2), '0');
        }

        if (str_starts_with($digits, '0')) {
            return '+84' . ltrim(substr($digits, 1), '0');
        }

        return '+84' . ltrim($digits, '0');
    }
}
