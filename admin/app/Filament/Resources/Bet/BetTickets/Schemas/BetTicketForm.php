<?php

namespace App\Filament\Resources\Bet\BetTickets\Schemas;

use App\Enum\Bet\BetStatus;
use App\Enum\Bet\BetTicketType;
use App\Enum\Bet\GameType;
use App\Enum\Wallet\UnitTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Schemas\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Schemas\Schema;

class BetTicketForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Thông tin vé cược')
                ->schema([
                    TextInput::make('ticket_no')->label('Mã vé')->required()->maxLength(40),
                    Select::make('user_id')->label('Người dùng')->relationship('user', 'name')->searchable()->preload()->required(),
                    Select::make('wallet_id')->label('Ví')->relationship('wallet', 'id')->searchable()->preload()->required(),
                    Select::make('unit')->label('Đơn vị')->options(EnumPresenter::options(UnitTransaction::class))->required(),
                    Select::make('game_type')->label('Trò chơi')->options(EnumPresenter::options(GameType::class))->required(),
                    Select::make('period_id')->label('Kỳ')->relationship('period', 'period_no')->searchable()->preload()->required(),
                    Select::make('bet_type')->label('Loại vé')->options(EnumPresenter::options(BetTicketType::class))->required(),
                    TextInput::make('stake')->label('Tiền cược')->numeric()->required(),
                    TextInput::make('total_odds')->label('Tổng tỷ lệ')->numeric()->required(),
                    TextInput::make('potential_payout')->label('Tiền có thể thắng')->numeric()->required(),
                    TextInput::make('actual_payout')->label('Tiền thực nhận')->numeric(),
                    Select::make('status')->label('Trạng thái')->options(EnumPresenter::options(BetStatus::class))->required(),
                    TextInput::make('placed_ip')->label('IP đặt cược')->maxLength(45),
                    TextInput::make('placed_device')->label('Thiết bị')->maxLength(100),
                    DateTimePicker::make('settled_at')->label('Chốt lúc'),
                ])
                ->columns(2),
        ]);
    }
}
