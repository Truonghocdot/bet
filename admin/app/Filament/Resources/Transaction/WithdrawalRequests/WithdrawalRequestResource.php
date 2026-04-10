<?php

namespace App\Filament\Resources\Transaction\WithdrawalRequests;

use App\Filament\Resources\Transaction\WithdrawalRequests\Pages\CreateWithdrawalRequest;
use App\Filament\Resources\Transaction\WithdrawalRequests\Pages\EditWithdrawalRequest;
use App\Filament\Resources\Transaction\WithdrawalRequests\Pages\ListWithdrawalRequests;
use App\Filament\Resources\Transaction\WithdrawalRequests\Schemas\WithdrawalRequestForm;
use App\Filament\Resources\Transaction\WithdrawalRequests\Tables\WithdrawalRequestsTable;
use App\Models\Transaction\WithdrawalRequest;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\SoftDeletingScope;

class WithdrawalRequestResource extends BaseResource
{
    protected static ?string $model = WithdrawalRequest::class;
    protected static bool $canCreateRecords = false;
    protected static bool $canDeleteRecords = false;
    protected static bool $canDeleteAnyRecords = false;
    protected static bool $canForceDeleteRecords = false;
    protected static bool $canForceDeleteAnyRecords = false;
    protected static bool $canRestoreRecords = false;
    protected static bool $canRestoreAnyRecords = false;
    protected static UnitEnum|string|null $navigationGroup = 'Tài chính';
    protected static ?string $navigationLabel = 'Yêu cầu rút';
    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'id';

    protected static function abilityPrefix(): string
    {
        return 'finance.withdrawal-requests';
    }

    public static function form(Schema $schema): Schema
    {
        return WithdrawalRequestForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return WithdrawalRequestsTable::configure($table);
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
            'index' => ListWithdrawalRequests::route('/'),
            'create' => CreateWithdrawalRequest::route('/create'),
            'edit' => EditWithdrawalRequest::route('/{record}/edit'),
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
