<?php

namespace App\Filament\Resources\Transaction\AccountWithdrawalInfos;

use App\Filament\Resources\Transaction\AccountWithdrawalInfos\Pages\CreateAccountWithdrawalInfo;
use App\Filament\Resources\Transaction\AccountWithdrawalInfos\Pages\EditAccountWithdrawalInfo;
use App\Filament\Resources\Transaction\AccountWithdrawalInfos\Pages\ListAccountWithdrawalInfos;
use App\Filament\Resources\Transaction\AccountWithdrawalInfos\Schemas\AccountWithdrawalInfoForm;
use App\Filament\Resources\Transaction\AccountWithdrawalInfos\Tables\AccountWithdrawalInfosTable;
use App\Models\Transaction\AccountWithdrawalInfo;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\SoftDeletingScope;

class AccountWithdrawalInfoResource extends BaseResource
{
    protected static ?string $model = AccountWithdrawalInfo::class;
    protected static UnitEnum|string|null $navigationGroup = 'Tài chính';
    protected static ?string $navigationLabel = 'Tài khoản rút';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'id';

    protected static function abilityPrefix(): string
    {
        return 'finance.account-withdrawal-infos';
    }

    public static function form(Schema $schema): Schema
    {
        return AccountWithdrawalInfoForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return AccountWithdrawalInfosTable::configure($table);
    }

    public static function getRelations(): array
    {
        return [
            //
        ];
    }

    public static function getPages(): array
    {
        return [
            'index' => ListAccountWithdrawalInfos::route('/'),
            'create' => CreateAccountWithdrawalInfo::route('/create'),
            'edit' => EditAccountWithdrawalInfo::route('/{record}/edit'),
        ];
    }

    public static function getRecordRouteBindingEloquentQuery(): Builder
    {
        return parent::getRecordRouteBindingEloquentQuery()
            ->withoutGlobalScopes([
                SoftDeletingScope::class,
            ]);
    }
}
