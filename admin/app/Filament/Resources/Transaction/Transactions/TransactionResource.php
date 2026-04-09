<?php

namespace App\Filament\Resources\Transaction\Transactions;

use App\Filament\Resources\Transaction\Transactions\Pages\CreateTransaction;
use App\Filament\Resources\Transaction\Transactions\Pages\EditTransaction;
use App\Filament\Resources\Transaction\Transactions\Pages\ListTransactions;
use App\Filament\Resources\Transaction\Transactions\Schemas\TransactionForm;
use App\Filament\Resources\Transaction\Transactions\Tables\TransactionsTable;
use App\Models\Transaction\Transaction;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\SoftDeletingScope;

class TransactionResource extends BaseResource
{
    protected static ?string $model = Transaction::class;
    protected static UnitEnum|string|null $navigationGroup = 'Tài chính';
    protected static ?string $navigationLabel = 'Giao dịch';
    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'id';

    protected static function abilityPrefix(): string
    {
        return 'finance.transactions';
    }

    public static function form(Schema $schema): Schema
    {
        return TransactionForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return TransactionsTable::configure($table);
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
            'index' => ListTransactions::route('/'),
            'create' => CreateTransaction::route('/create'),
            'edit' => EditTransaction::route('/{record}/edit'),
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
