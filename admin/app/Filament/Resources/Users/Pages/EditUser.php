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
     * @var array<string, mixed>
     */
    protected array $withdrawalInfoPayload = [];

    protected function mutateFormDataBeforeFill(array $data): array
    {
        $wallets = $this->record->wallets()
            ->get()
            ->keyBy(fn ($wallet): int => (int) ($wallet->unit?->value ?? $wallet->unit));
        $withdrawalInfo = $this->record->accountWithdrawalInfos()
            ->orderByDesc('is_default')
            ->orderByDesc('id')
            ->first();

        $data['wallet_vnd_balance'] = (string) ($wallets->get(UnitTransaction::VND->value)?->balance ?? 0);
        $data['wallet_usdt_balance'] = (string) ($wallets->get(UnitTransaction::USDT->value)?->balance ?? 0);
        $data['withdrawal_unit'] = (int) ($withdrawalInfo?->unit?->value ?? $withdrawalInfo?->unit ?? UnitTransaction::VND->value);
        $data['withdrawal_provider_code'] = (string) ($withdrawalInfo?->provider_code ?? '');
        $data['withdrawal_account_name'] = (string) ($withdrawalInfo?->account_name ?? '');
        $data['withdrawal_account_number'] = (string) ($withdrawalInfo?->account_number ?? '');
        $data['withdrawal_is_default'] = true;

        return $data;
    }

    protected function mutateFormDataBeforeSave(array $data): array
    {
        $this->walletBalancePayload = $this->extractWalletBalancePayload($data);
        $this->withdrawalInfoPayload = $this->extractWithdrawalInfoPayload($data);

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

        $this->syncDefaultWithdrawalInfo();
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
            $data['withdrawal_unit'],
            $data['withdrawal_provider_code'],
            $data['withdrawal_account_name'],
            $data['withdrawal_account_number'],
            $data['withdrawal_is_default'],
        );

        return $data;
    }

    /**
     * @param  array<string, mixed>  $data
     * @return array<string, mixed>
     */
    protected function extractWithdrawalInfoPayload(array $data): array
    {
        $accountName = trim((string) ($data['withdrawal_account_name'] ?? ''));
        $accountNumber = trim((string) ($data['withdrawal_account_number'] ?? ''));
        $providerCode = trim((string) ($data['withdrawal_provider_code'] ?? ''));
        $shouldSync = $accountName !== '' || $accountNumber !== '' || $providerCode !== '';

        if ($shouldSync && ($accountName === '' || $accountNumber === '')) {
            throw ValidationException::withMessages([
                'withdrawal_account_name' => 'Cần nhập đầy đủ chủ tài khoản và số tài khoản rút.',
                'withdrawal_account_number' => 'Cần nhập đầy đủ chủ tài khoản và số tài khoản rút.',
            ]);
        }

        return [
            'should_sync' => $shouldSync,
            'unit' => (int) ($data['withdrawal_unit'] ?? UnitTransaction::VND->value),
            'provider_code' => $providerCode !== '' ? $providerCode : null,
            'account_name' => $accountName,
            'account_number' => $accountNumber,
        ];
    }

    protected function syncDefaultWithdrawalInfo(): void
    {
        if (($this->withdrawalInfoPayload['should_sync'] ?? false) !== true) {
            return;
        }

        DB::transaction(function (): void {
            $relationship = $this->record->accountWithdrawalInfos();
            $defaultRecord = $relationship->where('is_default', true)->orderByDesc('id')->first()
                ?? $relationship->orderByDesc('id')->first();

            $payload = [
                'unit' => $this->withdrawalInfoPayload['unit'],
                'provider_code' => $this->withdrawalInfoPayload['provider_code'],
                'account_name' => $this->withdrawalInfoPayload['account_name'],
                'account_number' => $this->withdrawalInfoPayload['account_number'],
                'is_default' => true,
            ];

            if ($defaultRecord instanceof AccountWithdrawalInfo) {
                $relationship
                    ->whereKeyNot($defaultRecord->getKey())
                    ->where('is_default', true)
                    ->update(['is_default' => false]);

                $defaultRecord->fill($payload)->save();

                return;
            }

            $relationship->create($payload);
        });
    }
}
