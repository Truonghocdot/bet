<?php

namespace App\Filament\Resources\Bet\GamePeriods\Schemas;

use App\Enum\Bet\DrawSource;
use App\Enum\Bet\GameType;
use App\Enum\Bet\PeriodStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Textarea;
use Filament\Schemas\Schema;

class GamePeriodForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Thông tin kỳ game')
                ->schema([
                    Select::make('game_type')
                        ->label('Trò chơi')
                        ->options(EnumPresenter::options(GameType::class))
                        ->required(),
                    TextInput::make('period_no')
                        ->label('Kỳ số')
                        ->required()
                        ->maxLength(50),
                    TextInput::make('room_code')
                        ->label('Phòng')
                        ->maxLength(30),
                    DateTimePicker::make('open_at')->label('Mở lúc')->required(),
                    DateTimePicker::make('close_at')->label('Đóng lúc')->required(),
                    DateTimePicker::make('draw_at')->label('Quay lúc')->required(),
                    DateTimePicker::make('settled_at')->label('Chốt lúc'),
                    Select::make('status')
                        ->label('Trạng thái')
                        ->options(EnumPresenter::options(PeriodStatus::class))
                        ->required(),
                    Select::make('draw_source')
                        ->label('Nguồn ket qua')
                        ->options(EnumPresenter::options(DrawSource::class)),
                    Textarea::make('result_payload')
                        ->label('Payload kết quả')
                        ->rows(5)
                        ->columnSpanFull(),
                    TextInput::make('result_hash')
                        ->label('Hash kết quả')
                        ->maxLength(255),
                ])
                ->columns(2),
        ]);
    }
}
