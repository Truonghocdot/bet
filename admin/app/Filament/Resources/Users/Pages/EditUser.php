<?php

namespace App\Filament\Resources\Users\Pages;

use App\Enum\User\RoleUser;
use App\Enum\Wallet\UnitTransaction;
use App\Models\Transaction\AccountWithdrawalInfo;
use App\Services\Admin\UserWalletBalanceService;
use Filament\Actions\DeleteAction;
use Filament\Actions\ForceDeleteAction;
use Filament\Actions\RestoreAction;
use Filament\Resources\Pages\EditRecord;
use Illuminate\Support\Facades\DB;
use Illuminate\Validation\ValidationException;

abstract class EditUser extends EditRecord
{
    protected static ?string $title = 'Hồ sơ người dùng';

    /**
     * @var array<int, mixed>
     */
    protected array $walletBalancePayload = [];

    /**
     * @var array<int, array<string, mixed>>
     */
    protected array $withdrawalAccountsPayload = [];

    protected function mutateFormDataBeforeFill(array $data): array
    {
        $wallets = $this->record->wallets()
            ->get()
            ->keyBy(fn ($wallet): int => (int) ($wallet->unit?->value ?? $wallet->unit));
        $vndWithdrawalInfo = $this->record->accountWithdrawalInfos()
            ->where('unit', UnitTransaction::VND->value)
            ->orderByDesc('is_default')
            ->orderByDesc('id')
            ->first();
        $usdtWithdrawalInfo = $this->record->accountWithdrawalInfos()
            ->where('unit', UnitTransaction::USDT->value)
            ->orderByDesc('is_default')
            ->orderByDesc('id')
            ->first();

        $data['wallet_vnd_balance'] = (string) ($wallets->get(UnitTransaction::VND->value)?->balance ?? 0);
        $data['wallet_usdt_balance'] = (string) ($wallets->get(UnitTransaction::USDT->value)?->balance ?? 0);
        $data['withdrawal_vnd_provider_code'] = (string) ($vndWithdrawalInfo?->provider_code ?? '');
        $data['withdrawal_vnd_account_name'] = (string) ($vndWithdrawalInfo?->account_name ?? '');
        $data['withdrawal_vnd_account_number'] = (string) ($vndWithdrawalInfo?->account_number ?? '');
        $data['withdrawal_usdt_provider_code'] = (string) ($usdtWithdrawalInfo?->provider_code ?? '');
        $data['withdrawal_usdt_account_name'] = (string) ($usdtWithdrawalInfo?->account_name ?? '');
        $data['withdrawal_usdt_account_number'] = (string) ($usdtWithdrawalInfo?->account_number ?? '');

        return $data;
    }

    protected function mutateFormDataBeforeSave(array $data): array
    {
        $this->walletBalancePayload = $this->extractWalletBalancePayload($data);
        $this->withdrawalAccountsPayload = $this->extractWithdrawalAccountsPayload($data);

        if (! array_key_exists('role', $data)) {
            return $this->stripFormRuntimeFields($data);
        }

        $requestedRole = RoleUser::tryFrom((int) $data['role']);
        $actorRole = auth()->user()?->role;
        $normalizedActorRole = $actorRole instanceof RoleUser ? $actorRole : null;

        if (! $requestedRole || ! RoleUser::canAssign($normalizedActorRole, $requestedRole)) {
            throw ValidationException::withMessages([
                'role' => 'Bạn không được phép gán vai trò này từ Filament.',
            ]);
        }

        return $this->stripFormRuntimeFields($data);
    }

    protected function afterSave(): void
    {
        app(UserWalletBalanceService::class)->syncAvailableBalances(
            $this->record,
            $this->walletBalancePayload,
            auth()->user(),
        );

        $this->syncWithdrawalAccounts();
    }

    protected function getHeaderActions(): array
    {
        return [
            DeleteAction::make(),
            ForceDeleteAction::make(),
            RestoreAction::make(),
        ];
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
            $data['change_password'],
            $data['password_confirmation'],
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
        DB::transaction(function (): void {
            foreach ($this->withdrawalAccountsPayload as $payload) {
                if (($payload['should_sync'] ?? false) !== true) {
                    continue;
                }

                $relationship = $this->record->accountWithdrawalInfos();
                $defaultRecord = $relationship
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

                $relationship
                    ->where('unit', $payload['unit'])
                    ->where('is_default', true)
                    ->when(
                        $defaultRecord instanceof AccountWithdrawalInfo,
                        fn ($query) => $query->whereKeyNot($defaultRecord->getKey()),
                    )
                    ->update(['is_default' => false]);

                if ($defaultRecord instanceof AccountWithdrawalInfo) {
                    $defaultRecord->fill($attributes)->save();

                    continue;
                }

                $relationship->create($attributes);
            }
        });
    }
}
