<?php

namespace App\Filament\Resources\Users\Pages;

use App\Enum\User\RoleUser;
use Filament\Actions\DeleteAction;
use Filament\Actions\ForceDeleteAction;
use Filament\Actions\RestoreAction;
use Filament\Resources\Pages\EditRecord;
use Illuminate\Validation\ValidationException;

abstract class EditUser extends EditRecord
{
    protected static ?string $title = 'Hồ sơ người dùng';

    protected function mutateFormDataBeforeSave(array $data): array
    {
        if (! array_key_exists('role', $data)) {
            return $data;
        }

        $requestedRole = RoleUser::tryFrom((int) $data['role']);
        $actorRole = auth()->user()?->role;
        $normalizedActorRole = $actorRole instanceof RoleUser ? $actorRole : null;

        if (! $requestedRole || ! RoleUser::canAssign($normalizedActorRole, $requestedRole)) {
            throw ValidationException::withMessages([
                'role' => 'Bạn không được phép gán vai trò này từ Filament.',
            ]);
        }

        return $data;
    }

    protected function getHeaderActions(): array
    {
        return [
            DeleteAction::make(),
            ForceDeleteAction::make(),
            RestoreAction::make(),
        ];
    }
}
