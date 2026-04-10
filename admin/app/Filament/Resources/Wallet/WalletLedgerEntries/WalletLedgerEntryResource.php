<?php

namespace App\Filament\Resources\Wallet\WalletLedgerEntries;

use App\Filament\Resources\Wallet\WalletLedgerEntries\Pages\CreateWalletLedgerEntry;
use App\Filament\Resources\Wallet\WalletLedgerEntries\Pages\EditWalletLedgerEntry;
use App\Filament\Resources\Wallet\WalletLedgerEntries\Pages\ListWalletLedgerEntries;
use App\Filament\Resources\Wallet\WalletLedgerEntries\Schemas\WalletLedgerEntryForm;
use App\Filament\Resources\Wallet\WalletLedgerEntries\Tables\WalletLedgerEntriesTable;
use App\Models\Wallet\WalletLedgerEntry;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;

class WalletLedgerEntryResource extends BaseResource
{
    protected static ?string $model = WalletLedgerEntry::class;
    protected static bool $canCreateRecords = false;
    protected static bool $canUpdateRecords = false;
    protected static bool $canDeleteRecords = false;
    protected static bool $canDeleteAnyRecords = false;
    protected static bool $canForceDeleteRecords = false;
    protected static bool $canForceDeleteAnyRecords = false;
    protected static bool $canRestoreRecords = false;
    protected static bool $canRestoreAnyRecords = false;
    protected static UnitEnum|string|null $navigationGroup = 'Tài chính';
    protected static ?string $navigationLabel = 'Sổ cái ví';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'id';

    protected static function abilityPrefix(): string
    {
        return 'finance.wallet-ledger-entries';
    }

    public static function form(Schema $schema): Schema
    {
        return WalletLedgerEntryForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return WalletLedgerEntriesTable::configure($table);
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
            'index' => ListWalletLedgerEntries::route('/'),
            'create' => CreateWalletLedgerEntry::route('/create'),
            'edit' => EditWalletLedgerEntry::route('/{record}/edit'),
        ];
    }
}
