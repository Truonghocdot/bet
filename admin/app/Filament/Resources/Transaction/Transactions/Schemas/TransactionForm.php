<?php

namespace App\Filament\Resources\Transaction\Transactions\Schemas;

use App\Enum\Transaction\TransactionStatus;
use App\Enum\Transaction\TypeTransaction;
use App\Enum\Wallet\UnitTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Textarea;
use Filament\Schemas\Schema;

class TransactionForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Thông tin giao dịch')
                ->schema([
                    Select::make('user_id')
                        ->label('Người dùng')
                        ->relationship('user', 'name')
                        ->searchable()
                        ->preload()
                        ->required(),
                    Select::make('wallet_id')
                        ->label('Ví')
                        ->relationship('wallet', 'id')
                        ->searchable()
                        ->preload()
                        ->required(),
                    Select::make('unit')
                        ->label('Đơn vị')
                        ->options(EnumPresenter::options(UnitTransaction::class))
                        ->required(),
                    Select::make('type')
                        ->label('Loại')
                        ->options(EnumPresenter::options(TypeTransaction::class))
                        ->required(),
                    TextInput::make('amount')
                        ->label('Số tiền')
                        ->numeric()
                        ->required(),
                    TextInput::make('fee')
                        ->label('Phí')
                        ->numeric()
                        ->required()
                        ->default(0),
                    TextInput::make('net_amount')
                        ->label('Số tiền thực nhận')
                        ->numeric()
                        ->required(),
                    Select::make('status')
                        ->label('Trạng thái')
                        ->options(EnumPresenter::options(TransactionStatus::class))
                        ->required(),
                    TextInput::make('provider')
                        ->label('Nhà cung cấp')
                        ->maxLength(50),
                    TextInput::make('provider_txn_id')
                        ->label('Mã giao dịch nhà cung cấp')
                        ->maxLength(100),
                    Textarea::make('reason_failed')
                        ->label('Lý do thất bại')
                        ->rows(3)
                        ->columnSpanFull(),
                    Select::make('approved_by')
                        ->label('Duyệt bởi')
                        ->relationship('approvedBy', 'name')
                        ->searchable()
                        ->preload(),
                    DateTimePicker::make('approved_at')
                        ->label('Duyệt lúc'),
                ])
                ->columns(2),
        ]);
    }
}
