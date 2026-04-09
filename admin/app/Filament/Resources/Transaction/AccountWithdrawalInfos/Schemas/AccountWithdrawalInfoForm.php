<?php

namespace App\Filament\Resources\Transaction\AccountWithdrawalInfos\Schemas;

use App\Enum\Wallet\UnitTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Schemas\Schema;

class AccountWithdrawalInfoForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Thông tin tài khoản rút')
                ->schema([
                    Select::make('user_id')
                        ->label('Người dùng')
                        ->relationship('user', 'name')
                        ->searchable()
                        ->preload()
                        ->required(),
                    Select::make('unit')
                        ->label('Đơn vị')
                        ->options(EnumPresenter::options(UnitTransaction::class))
                        ->required(),
                    TextInput::make('provider_code')
                        ->label('Mã nhà cung cấp')
                        ->maxLength(50)
                        ->required(),
                    TextInput::make('account_name')
                        ->label('Chủ tài khoản')
                        ->maxLength(255)
                        ->required(),
                    TextInput::make('account_number')
                        ->label('Số tài khoản')
                        ->maxLength(255)
                        ->required(),
                    Toggle::make('is_default')
                        ->label('Mặc định'),
                ])
                ->columns(2),
        ]);
    }
}
