<?php

namespace App\Filament\Resources\Wallet\WalletLedgerEntries\Schemas;

use App\Enum\Wallet\LedgerDirection;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Schemas\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Textarea;
use Filament\Schemas\Schema;

class WalletLedgerEntryForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Chi tiết sổ cái')
                ->schema([
                    Select::make('wallet_id')
                        ->label('Ví')
                        ->relationship('wallet', 'id')
                        ->searchable()
                        ->preload()
                        ->required(),
                    Select::make('user_id')
                        ->label('Người dùng')
                        ->relationship('user', 'name')
                        ->searchable()
                        ->preload()
                        ->required(),
                    Select::make('direction')
                        ->label('Chiều')
                        ->options(EnumPresenter::options(LedgerDirection::class))
                        ->required(),
                    TextInput::make('amount')
                        ->label('Số tiền')
                        ->numeric()
                        ->required(),
                    TextInput::make('balance_before')
                        ->label('Số dư trước')
                        ->numeric()
                        ->required(),
                    TextInput::make('balance_after')
                        ->label('Số dư sau')
                        ->numeric()
                        ->required(),
                    TextInput::make('reference_type')
                        ->label('Loại tham chiếu')
                        ->maxLength(50)
                        ->required(),
                    TextInput::make('reference_id')
                        ->label('ID tham chiếu')
                        ->numeric(),
                    Textarea::make('note')
                        ->label('Ghi chú')
                        ->rows(3)
                        ->columnSpanFull(),
                    DateTimePicker::make('created_at')
                        ->label('Tạo lúc')
                        ->required(),
                ])
                ->columns(2),
        ]);
    }
}
