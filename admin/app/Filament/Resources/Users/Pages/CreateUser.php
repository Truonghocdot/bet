<?php

namespace App\Filament\Resources\Users\Pages;

use App\Enum\User\RoleUser;
use App\Services\Admin\UserProvisioningService;
use Filament\Resources\Pages\CreateRecord;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Validation\ValidationException;

abstract class CreateUser extends CreateRecord
{
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

        return $data;
    }

    protected function handleRecordCreation(array $data): Model
    {
        return app(UserProvisioningService::class)->createFromErp($data, auth()->user());
    }
}
