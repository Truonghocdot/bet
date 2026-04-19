<?php

namespace App\Filament\Resources\Users\Pages;

use App\Enum\User\RoleUser;
use App\Enum\Wallet\UnitTransaction;
use App\Services\Admin\UserWalletBalanceService;
use Filament\Actions\DeleteAction;
use Filament\Actions\ForceDeleteAction;
use Filament\Actions\RestoreAction;
use Filament\Resources\Pages\EditRecord;
use Illuminate\Validation\ValidationException;

abstract class EditUser extends EditRecord
{
    protected static ?string $title = 'Hồ sơ người dùng';

    /**
     * @var array<int, mixed>
     */
    protected array $walletBalancePayload = [];

    protected function mutateFormDataBeforeFill(array $data): array
    {
        $wallets = $this->record->wallets()
            ->get()
            ->keyBy(fn ($wallet): int => (int) ($wallet->unit?->value ?? $wallet->unit));

        $data['wallet_vnd_balance'] = (string) ($wallets->get(UnitTransaction::VND->value)?->balance ?? 0);
        $data['wallet_usdt_balance'] = (string) ($wallets->get(UnitTransaction::USDT->value)?->balance ?? 0);

        return $data;
    }

    protected function mutateFormDataBeforeSave(array $data): array
    {
        $this->walletBalancePayload = $this->extractWalletBalancePayload($data);

        if (! array_key_exists('role', $data)) {
            return $this->stripWalletBalanceFields($data);
        }

        $requestedRole = RoleUser::tryFrom((int) $data['role']);
        $actorRole = auth()->user()?->role;
        $normalizedActorRole = $actorRole instanceof RoleUser ? $actorRole : null;

        if (! $requestedRole || ! RoleUser::canAssign($normalizedActorRole, $requestedRole)) {
            throw ValidationException::withMessages([
                'role' => 'Bạn không được phép gán vai trò này từ Filament.',
            ]);
        }

        return $this->stripWalletBalanceFields($data);
    }

    protected function afterSave(): void
    {
        app(UserWalletBalanceService::class)->syncAvailableBalances(
            $this->record,
            $this->walletBalancePayload,
            auth()->user(),
        );
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
    protected function stripWalletBalanceFields(array $data): array
    {
        unset(
            $data['wallet_vnd_balance'],
            $data['wallet_usdt_balance'],
        );

        return $data;
    }
}
