<?php

namespace App\Filament\Resources\Users;

use App\Enum\User\RoleUser;
use App\Filament\Resources\Users\Pages\Admins\CreateAdminUser;
use App\Filament\Resources\Users\Pages\Admins\EditAdminUser;
use App\Filament\Resources\Users\Pages\Admins\ListAdminUsers;
use BackedEnum;
use Filament\Support\Icons\Heroicon;
use UnitEnum;

class AdminUserResource extends UserResource
{
    protected static UnitEnum|string|null $navigationGroup = 'Người dùng';
    protected static ?string $navigationLabel = 'Admin';
    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedShieldCheck;
    protected static ?int $navigationSort = 0;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'system.users.admins';
    }

    protected static function resourceRole(): ?RoleUser
    {
        return RoleUser::ADMIN;
    }

    public static function getPages(): array
    {
        return [
            'index' => ListAdminUsers::route('/'),
            'create' => CreateAdminUser::route('/create'),
            'edit' => EditAdminUser::route('/{record}/edit'),
        ];
    }
}
