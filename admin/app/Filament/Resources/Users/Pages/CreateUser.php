<?php

namespace App\Filament\Resources\Users\Pages;

use App\Enum\User\RoleUser;
use App\Enum\Wallet\UnitTransaction;
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

        return $this->stripWalletBalanceFields($data);
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
