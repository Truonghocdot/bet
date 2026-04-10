<?php

namespace App\Filament\Resources\Wallet\Wallets;

use App\Filament\Resources\Wallet\Wallets\Pages\CreateWallet;
use App\Filament\Resources\Wallet\Wallets\Pages\EditWallet;
use App\Filament\Resources\Wallet\Wallets\Pages\ListWallets;
use App\Filament\Resources\Wallet\Wallets\Schemas\WalletForm;
use App\Filament\Resources\Wallet\Wallets\Tables\WalletsTable;
use App\Models\Wallet\Wallet;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;

class WalletResource extends BaseResource
{
    protected static ?string $model = Wallet::class;
    protected static bool $canCreateRecords = false;
    protected static bool $canUpdateRecords = false;
    protected static bool $canDeleteRecords = false;
    protected static bool $canDeleteAnyRecords = false;
    protected static bool $canForceDeleteRecords = false;
    protected static bool $canForceDeleteAnyRecords = false;
    protected static bool $canRestoreRecords = false;
    protected static bool $canRestoreAnyRecords = false;
    protected static UnitEnum|string|null $navigationGroup = 'Tài chính';
    protected static ?string $navigationLabel = 'Ví';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'user_id';

    protected static function abilityPrefix(): string
    {
        return 'finance.wallets';
    }

    public static function form(Schema $schema): Schema
    {
        return WalletForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return WalletsTable::configure($table);
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
            'index' => ListWallets::route('/'),
            'create' => CreateWallet::route('/create'),
            'edit' => EditWallet::route('/{record}/edit'),
        ];
    }
}
