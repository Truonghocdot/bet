<?php

namespace App\Filament\Resources\Users;

use App\Enum\User\RoleUser;
use App\Filament\Resources\Users\Pages\Staff\CreateStaffUser;
use App\Filament\Resources\Users\Pages\Staff\EditStaffUser;
use App\Filament\Resources\Users\Pages\Staff\ListStaffUsers;
use BackedEnum;
use UnitEnum;
use Filament\Support\Icons\Heroicon;

class StaffUserResource extends UserResource
{
    protected static UnitEnum|string|null $navigationGroup = 'Người dùng';
    protected static ?string $navigationLabel = 'Staff';
    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedBuildingLibrary;
    protected static ?int $navigationSort = 2;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'system.users.staffs';
    }

    protected static function resourceRole(): ?RoleUser
    {
        return RoleUser::STAFF;
    }

    protected static function forceAffiliateProfile(): bool
    {
        return true;
    }

    public static function getPages(): array
    {
        return [
            'index' => ListStaffUsers::route('/'),
            'create' => CreateStaffUser::route('/create'),
            'edit' => EditStaffUser::route('/{record}/edit'),
        ];
    }
}
