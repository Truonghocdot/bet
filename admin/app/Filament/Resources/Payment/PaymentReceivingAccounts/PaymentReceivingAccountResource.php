<?php

namespace App\Filament\Resources\Payment\PaymentReceivingAccounts;

use App\Filament\Resources\Payment\PaymentReceivingAccounts\Pages\CreatePaymentReceivingAccount;
use App\Filament\Resources\Payment\PaymentReceivingAccounts\Pages\EditPaymentReceivingAccount;
use App\Filament\Resources\Payment\PaymentReceivingAccounts\Pages\ListPaymentReceivingAccounts;
use App\Filament\Resources\Payment\PaymentReceivingAccounts\Schemas\PaymentReceivingAccountForm;
use App\Filament\Resources\Payment\PaymentReceivingAccounts\Tables\PaymentReceivingAccountsTable;
use App\Models\Payment\PaymentReceivingAccount;
use BackedEnum;
use UnitEnum;
use App\Filament\Resources\BaseResource;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\SoftDeletingScope;

class PaymentReceivingAccountResource extends BaseResource
{
    protected static ?string $model = PaymentReceivingAccount::class;
    protected static UnitEnum|string|null $navigationGroup = 'Thanh toán';
    protected static ?string $navigationLabel = 'Tài khoản nhận tiền';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedRectangleStack;

    protected static ?string $recordTitleAttribute = 'code';

    protected static function abilityPrefix(): string
    {
        return 'payment.payment-receiving-accounts';
    }

    public static function form(Schema $schema): Schema
    {
        return PaymentReceivingAccountForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return PaymentReceivingAccountsTable::configure($table);
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
            'index' => ListPaymentReceivingAccounts::route('/'),
            'create' => CreatePaymentReceivingAccount::route('/create'),
            'edit' => EditPaymentReceivingAccount::route('/{record}/edit'),
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
