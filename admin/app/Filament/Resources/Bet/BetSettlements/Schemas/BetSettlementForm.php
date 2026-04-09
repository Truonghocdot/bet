<?php

namespace App\Filament\Resources\Bet\BetSettlements\Schemas;

use App\Enum\Bet\BetStatus;
use App\Enum\Bet\SettlementType;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Schemas\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Textarea;
use Filament\Schemas\Schema;

class BetSettlementForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Chốt vé')
                ->schema([
                    Select::make('ticket_id')->label('Vé')->relationship('ticket', 'ticket_no')->searchable()->preload()->required(),
                    Select::make('period_id')->label('Kỳ')->relationship('period', 'period_no')->searchable()->preload()->required(),
                    Select::make('settlement_type')->label('Kiểu chốt')->options(EnumPresenter::options(SettlementType::class))->required(),
                    Select::make('before_status')->label('Trạng thái trước')->options(EnumPresenter::options(BetStatus::class))->required(),
                    Select::make('after_status')->label('Trạng thái sau')->options(EnumPresenter::options(BetStatus::class))->required(),
                    TextInput::make('payout_amount')->label('Tiền trả')->numeric()->required(),
                    TextInput::make('profit_loss')->label('Lãi lỗ')->numeric()->required(),
                    Textarea::make('note')->label('Ghi chú')->rows(3)->columnSpanFull(),
                    Select::make('settled_by')->label('Chốt bởi')->relationship('settledBy', 'name')->searchable()->preload(),
                    DateTimePicker::make('created_at')->label('Tạo lúc')->required(),
                ])
                ->columns(2),
        ]);
    }
}
