<?php

namespace App\Filament\Resources\Payment\PaymentReceivingAccounts\Schemas;

use App\Enum\Payment\PaymentReceivingAccountStatus;
use App\Enum\Payment\PaymentReceivingAccountType;
use App\Enum\Wallet\UnitTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\FileUpload;
use Filament\Forms\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\Textarea;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Schemas\Schema;

class PaymentReceivingAccountForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema
            ->components([
                Section::make('Thông tin tài khoản nhận tiền')
                    ->schema([
                        TextInput::make('code')
                            ->label('Mã')
                            ->required()
                            ->maxLength(50),
                        TextInput::make('name')
                            ->label('Tên')
                            ->required()
                            ->maxLength(100),
                        Select::make('type')
                            ->label('Loại')
                            ->options(EnumPresenter::options(PaymentReceivingAccountType::class))
                            ->required(),
                        Select::make('unit')
                            ->label('Đơn vị')
                            ->options(EnumPresenter::options(UnitTransaction::class))
                            ->required(),
                        TextInput::make('provider_code')
                            ->label('Mã nhà cung cấp')
                            ->maxLength(50),
                        TextInput::make('account_name')
                            ->label('Chủ tài khoản')
                            ->maxLength(255),
                        TextInput::make('account_number')
                            ->label('Số tài khoản')
                            ->maxLength(255),
                        TextInput::make('wallet_address')
                            ->label('Địa chỉ ví')
                            ->maxLength(255),
                        TextInput::make('network')
                            ->label('Mạng')
                            ->maxLength(50),
                        FileUpload::make('qr_code_path')
                            ->label('Ảnh QR')
                            ->disk('public')
                            ->directory('payment-receiving-accounts')
                            ->preserveFilenames(),
                        Textarea::make('instructions')
                            ->label('Hướng dẫn')
                            ->rows(3)
                            ->columnSpanFull(),
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
                    ->columns(2),
            ]);
    }
}
