<?php

namespace App\Filament\Resources\Payment\PaymentReceivingAccounts\Schemas;

use App\Enum\Payment\PaymentReceivingAccountStatus;
use App\Enum\Payment\PaymentReceivingAccountType;
use App\Enum\Wallet\UnitTransaction;
use App\Models\Payment\VietQrBank;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\FileUpload;
use Filament\Forms\Components\Hidden;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\Textarea;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Schemas\Components\Section;
use Filament\Schemas\Schema;

class PaymentReceivingAccountForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema
            ->components([
                Section::make('Thông tin tài khoản nhận tiền')
                    ->schema([
                        Hidden::make('type')
                            ->default(PaymentReceivingAccountType::BANK->value),
                        Hidden::make('unit')
                            ->default(UnitTransaction::VND->value),
                        Select::make('provider_code')
                            ->label('Ngân hàng (VietQR)')
                            ->options(fn (): array => VietQrBank::query()
                                ->orderBy('short_name')
                                ->get()
                                ->mapWithKeys(fn (VietQrBank $bank): array => [$bank->code => "{$bank->short_name} ({$bank->code}) - {$bank->name}"])
                                ->all())
                            ->searchable()
                            ->preload()
                            ->required(),
                        TextInput::make('account_name')
                            ->label('Chủ tài khoản')
                            ->required()
                            ->maxLength(255),
                        TextInput::make('account_number')
                            ->label('Số tài khoản')
                            ->required()
                            ->maxLength(255),
                        Select::make('status')
                            ->label('Trạng thái')
                            ->options(EnumPresenter::options(PaymentReceivingAccountStatus::class))
                            ->required(),
                        Toggle::make('is_default')
                            ->label('Mặc định'),
                        TextInput::make('sort_order')
                            ->label('Sắp xếp')
                            ->numeric()
                            ->default(0),
                    ])
                    ->columns(2)
                    ->columnSpanFull(),
            ]);
    }
}
