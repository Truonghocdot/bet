<?php

namespace App\Filament\Resources\Users;

use App\Filament\Resources\Users\Schemas\UserForm;
use App\Filament\Resources\Users\Tables\UsersTable;
use App\Enum\User\RoleUser;
use App\Models\User;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\SoftDeletingScope;

abstract class UserResource extends BaseResource
{
    protected static ?string $model = User::class;
    protected static ?string $recordTitleAttribute = 'name';

    protected static function abilityPrefix(): string
    {
        return 'system.users';
    }

    protected static function resourceRole(): ?RoleUser
    {
        return null;
    }

    public static function shouldRegisterNavigation(): bool
    {
        return false;
    }

    public static function form(Schema $schema): Schema
    {
        return UserForm::configure($schema, static::resourceRole());
    }

    public static function table(Table $table): Table
    {
        return UsersTable::configure($table, static::resourceRole());
    }

    public static function getRelations(): array
    {
        return [
            \Filament\Resources\RelationManagers\RelationGroup::make('Tài chính', [
                RelationManagers\WalletsRelationManager::class,
                RelationManagers\TransactionsRelationManager::class,
                RelationManagers\WithdrawalRequestsRelationManager::class,
                RelationManagers\AccountWithdrawalInfosRelationManager::class,
            ]),
            RelationManagers\BetTicketsRelationManager::class,
            RelationManagers\AffiliateProfileRelationManager::class,
        ];
    }

    protected static function applyRoleConstraint(Builder $query): Builder
    {
        if ($role = static::resourceRole()) {
            $query->where('role', $role->value);
        }

        return $query;
    }

    public static function getEloquentQuery(): Builder
    {
        return static::applyRoleConstraint(parent::getEloquentQuery());
    }

    public static function getRecordRouteBindingEloquentQuery(): Builder
    {
        return static::applyRoleConstraint(parent::getRecordRouteBindingEloquentQuery())
            ->withoutGlobalScopes([
                SoftDeletingScope::class,
            ]);
    }
}
