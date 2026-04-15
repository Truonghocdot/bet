<?php

namespace App\Filament\Resources\Users;

use App\Enum\User\RoleUser;
use App\Filament\Resources\Users\Pages\Clients\CreateClientUser;
use App\Filament\Resources\Users\Pages\Clients\EditClientUser;
use App\Filament\Resources\Users\Pages\Clients\ListClientUsers;
use BackedEnum;
use UnitEnum;
use Filament\Support\Icons\Heroicon;

class ClientUserResource extends UserResource
{
    protected static UnitEnum|string|null $navigationGroup = 'Người dùng';
    protected static ?string $navigationLabel = 'Người chơi';
    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedUsers;
    protected static ?int $navigationSort = 1;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'system.users.clients';
    }

    protected static function resourceRole(): ?RoleUser
    {
        return RoleUser::CLIENT;
    }

    public static function getPages(): array
    {
        return [
            'index' => ListClientUsers::route('/'),
            'create' => CreateClientUser::route('/create'),
            'edit' => EditClientUser::route('/{record}/edit'),
        ];
    }
}
