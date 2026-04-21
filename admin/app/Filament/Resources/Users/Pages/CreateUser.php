<?php

namespace App\Filament\Resources\Users\Pages;

use App\Enum\User\RoleUser;
use App\Enum\Wallet\UnitTransaction;
use App\Models\Transaction\AccountWithdrawalInfo;
use App\Services\Admin\UserWalletBalanceService;
use App\Services\Admin\UserProvisioningService;
use Filament\Resources\Pages\CreateRecord;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Validation\ValidationException;

abstract class CreateUser extends CreateRecord
{
    /**
     * @var array<int, mixed>
     */
    protected array $walletBalancePayload = [];

    /**
     * @var array<int, array<string, mixed>>
     */
    protected array $withdrawalAccountsPayload = [];

    protected static function fixedRole(): ?RoleUser
    {
        return null;
    }

    protected static function forceAffiliateProfile(): bool
    {
        return false;
    }

    protected function mutateFormDataBeforeCreate(array $data): array
    {
        $this->walletBalancePayload = $this->extractWalletBalancePayload($data);
        $this->withdrawalAccountsPayload = $this->extractWithdrawalAccountsPayload($data);

        if ($role = static::fixedRole()) {
            $data['role'] = $role->value;
        }

        $requestedRole = RoleUser::tryFrom((int) ($data['role'] ?? 0));
        $actorRole = auth()->user()?->role;
        $normalizedActorRole = $actorRole instanceof RoleUser ? $actorRole : null;

        if (! $requestedRole || ! RoleUser::canAssign($normalizedActorRole, $requestedRole)) {
            throw ValidationException::withMessages([
                'role' => 'Bạn không được phép tạo tài khoản với vai trò này.',
            ]);
        }

        if (static::forceAffiliateProfile()) {
            $data['provision_affiliate_profile'] = true;
        }

        $data['provision_account_withdrawal_info'] = false;

        return $this->stripFormRuntimeFields($data);
    }

    protected function handleRecordCreation(array $data): Model
    {
        return app(UserProvisioningService::class)->createFromErp($data, auth()->user());
    }

    protected function afterCreate(): void
    {
        app(UserWalletBalanceService::class)->syncAvailableBalances(
            $this->record,
            $this->walletBalancePayload,
            auth()->user(),
        );

        $this->syncWithdrawalAccounts();
    }

    /**
     * @param  array<string, mixed>  $data
     * @return array<int, mixed>
     */
    protected function extractWalletBalancePayload(array $data): array
    {
        return [
            UnitTransaction::VND->value => $data['wallet_vnd_balance'] ?? 0,
            UnitTransaction::USDT->value => $data['wallet_usdt_balance'] ?? 0,
        ];
    }

    /**
     * @param  array<string, mixed>  $data
     * @return array<string, mixed>
     */
    protected function stripFormRuntimeFields(array $data): array
    {
        unset(
            $data['wallet_vnd_balance'],
            $data['wallet_usdt_balance'],
            $data['withdrawal_vnd_provider_code'],
            $data['withdrawal_vnd_account_name'],
            $data['withdrawal_vnd_account_number'],
            $data['withdrawal_usdt_provider_code'],
            $data['withdrawal_usdt_account_name'],
            $data['withdrawal_usdt_account_number'],
        );

        return $data;
    }

    /**
     * @param  array<string, mixed>  $data
     * @return array<int, array<string, mixed>>
     */
    protected function extractWithdrawalAccountsPayload(array $data): array
    {
        return [
            UnitTransaction::VND->value => $this->normalizeWithdrawalAccountPayload(
                $data,
                UnitTransaction::VND,
                'withdrawal_vnd_provider_code',
                'withdrawal_vnd_account_name',
                'withdrawal_vnd_account_number',
            ),
            UnitTransaction::USDT->value => $this->normalizeWithdrawalAccountPayload(
                $data,
                UnitTransaction::USDT,
                'withdrawal_usdt_provider_code',
                'withdrawal_usdt_account_name',
                'withdrawal_usdt_account_number',
            ),
        ];
    }

    /**
     * @param  array<string, mixed>  $data
     * @return array<string, mixed>
     */
    protected function normalizeWithdrawalAccountPayload(
        array $data,
        UnitTransaction $unit,
        string $providerField,
        string $accountNameField,
        string $accountNumberField,
    ): array {
        $providerCode = trim((string) ($data[$providerField] ?? ''));
        $accountName = trim((string) ($data[$accountNameField] ?? ''));
        $accountNumber = trim((string) ($data[$accountNumberField] ?? ''));
        $shouldSync = $providerCode !== '' || $accountName !== '' || $accountNumber !== '';

        if ($shouldSync && ($accountName === '' || $accountNumber === '')) {
            throw ValidationException::withMessages([
                $accountNameField => 'Cần nhập đầy đủ tên/chủ sở hữu và số tài khoản hoặc địa chỉ ví.',
                $accountNumberField => 'Cần nhập đầy đủ tên/chủ sở hữu và số tài khoản hoặc địa chỉ ví.',
            ]);
        }

        return [
            'should_sync' => $shouldSync,
            'unit' => $unit->value,
            'provider_code' => $providerCode !== '' ? $providerCode : null,
            'account_name' => $accountName,
            'account_number' => $accountNumber,
        ];
    }

    protected function syncWithdrawalAccounts(): void
    {
        foreach ($this->withdrawalAccountsPayload as $payload) {
            if (($payload['should_sync'] ?? false) !== true) {
                continue;
            }

            $existingRecord = $this->record->accountWithdrawalInfos()
                ->where('unit', $payload['unit'])
                ->orderByDesc('is_default')
                ->orderByDesc('id')
                ->first();

            $attributes = [
                'unit' => $payload['unit'],
                'provider_code' => $payload['provider_code'],
                'account_name' => $payload['account_name'],
                'account_number' => $payload['account_number'],
                'is_default' => true,
            ];

            $this->record->accountWithdrawalInfos()
                ->where('unit', $payload['unit'])
                ->where('is_default', true)
                ->when(
                    $existingRecord instanceof AccountWithdrawalInfo,
                    fn ($query) => $query->whereKeyNot($existingRecord->getKey()),
                )
                ->update(['is_default' => false]);

            if ($existingRecord instanceof AccountWithdrawalInfo) {
                $existingRecord->fill($attributes)->save();

                continue;
            }

            $this->record->accountWithdrawalInfos()->create($attributes);
        }
    }
}
