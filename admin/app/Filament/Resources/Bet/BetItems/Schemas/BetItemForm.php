<?php

namespace App\Filament\Resources\Bet\BetItems\Schemas;

use App\Enum\Bet\BetItemResult;
use App\Enum\Bet\BetOptionType;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Schemas\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Textarea;
use Filament\Schemas\Schema;

class BetItemForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Chi tiết cược')
                ->schema([
                    Select::make('ticket_id')->label('Vé')->relationship('ticket', 'ticket_no')->searchable()->preload()->required(),
                    Select::make('period_id')->label('Kỳ')->relationship('period', 'period_no')->searchable()->preload()->required(),
                    Select::make('option_type')->label('Kiểu')->options(EnumPresenter::options(BetOptionType::class))->required(),
                    TextInput::make('option_key')->label('Mã')->required()->maxLength(100),
                    TextInput::make('option_label')->label('Nhãn')->required()->maxLength(150),
                    TextInput::make('odds_at_placement')->label('Tỷ lệ')->numeric()->required(),
                    TextInput::make('stake')->label('Tiền cược')->numeric()->required(),
                    Select::make('result')->label('Kết quả')->options(EnumPresenter::options(BetItemResult::class))->required(),
                    TextInput::make('payout_amount')->label('Tiền trả')->numeric(),
                    Textarea::make('result_payload')->label('Payload kết quả')->rows(5)->columnSpanFull(),
                    DateTimePicker::make('settled_at')->label('Chốt lúc'),
                ])
                ->columns(2),
        ]);
    }
}
