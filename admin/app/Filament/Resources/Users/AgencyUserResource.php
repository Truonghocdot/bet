<?php

namespace App\Filament\Resources\Users;

use App\Enum\User\RoleUser;
use App\Filament\Resources\Users\Pages\Agencies\CreateAgencyUser;
use App\Filament\Resources\Users\Pages\Agencies\EditAgencyUser;
use App\Filament\Resources\Users\Pages\Agencies\ListAgencyUsers;
use BackedEnum;
use UnitEnum;
use Filament\Support\Icons\Heroicon;

class AgencyUserResource extends UserResource
{
    protected static UnitEnum|string|null $navigationGroup = 'Người dùng';
    protected static ?string $navigationLabel = 'Đại lý';
    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedBriefcase;
    protected static ?int $navigationSort = 3;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'system.users.agencies';
    }

    protected static function resourceRole(): ?RoleUser
    {
        return RoleUser::AGENCY;
    }

    protected static function forceAffiliateProfile(): bool
    {
        return true;
    }

    public static function getPages(): array
    {
        return [
            'index' => ListAgencyUsers::route('/'),
            'create' => CreateAgencyUser::route('/create'),
            'edit' => EditAgencyUser::route('/{record}/edit'),
        ];
    }
}
