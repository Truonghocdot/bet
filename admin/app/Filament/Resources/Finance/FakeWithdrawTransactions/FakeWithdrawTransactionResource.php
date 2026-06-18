<?php

namespace App\Filament\Resources\Finance\FakeWithdrawTransactions;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\Finance\FakeWithdrawTransactions\Pages\ListFakeWithdrawTransactions;
use App\Filament\Resources\Finance\FakeWithdrawTransactions\Tables\FakeWithdrawTransactionsTable;
use App\Models\Finance\FakeWithdrawTransaction;
use BackedEnum;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;
use UnitEnum;

class FakeWithdrawTransactionResource extends BaseResource
{
    protected static ?string $model = FakeWithdrawTransaction::class;

    protected static bool $canCreateRecords = false;
    protected static bool $canUpdateRecords = false;
    protected static bool $canDeleteRecords = false;
    protected static bool $canDeleteAnyRecords = false;
    protected static bool $canForceDeleteRecords = false;
    protected static bool $canForceDeleteAnyRecords = false;
    protected static bool $canRestoreRecords = false;
    protected static bool $canRestoreAnyRecords = false;

    protected static UnitEnum|string|null $navigationGroup = 'Tài chính';

    protected static ?string $navigationLabel = 'Giao dịch rút';

    protected static ?int $navigationSort = 31;

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'finance.fake-withdraw-transactions';
    }

    public static function form(Schema $schema): Schema
    {
        return $schema->components([]);
    }

    public static function table(Table $table): Table
    {
        return FakeWithdrawTransactionsTable::configure($table);
    }

    public static function getPages(): array
    {
        return [
            'index' => ListFakeWithdrawTransactions::route('/'),
        ];
    }

    public static function getEloquentQuery(): Builder
    {
        return parent::getEloquentQuery()->latest('id');
    }
}
