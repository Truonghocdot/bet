<?php

namespace App\Filament\Resources\Wallet\Wallets\Schemas;

use App\Enum\Wallet\UnitTransaction;
use App\Enum\Wallet\WalletStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Schemas\Schema;

class WalletForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Thông tin ví')
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
                    TextInput::make('balance')
                        ->label('Số dư')
                        ->numeric()
                        ->required()
                        ->default(0),
                    TextInput::make('locked_balance')
                        ->label('Số dư khóa')
                        ->numeric()
                        ->required()
                        ->default(0),
                    Select::make('status')
                        ->label('Trạng thái')
                        ->options(EnumPresenter::options(WalletStatus::class))
                        ->required(),
                ])
                ->columns(2),
        ]);
    }
}
