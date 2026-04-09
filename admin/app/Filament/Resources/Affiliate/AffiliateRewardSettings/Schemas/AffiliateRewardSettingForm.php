<?php

namespace App\Filament\Resources\Affiliate\AffiliateRewardSettings\Schemas;

use App\Enum\Wallet\UnitTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Schemas\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Textarea;
use Filament\Forms\Components\Toggle;
use Filament\Schemas\Schema;

class AffiliateRewardSettingForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema
            ->components([
                Section::make('Cấu hình thưởng')
                    ->schema([
                        TextInput::make('name')
                            ->label('Tên')
                            ->required()
                            ->maxLength(100),
                        TextInput::make('required_qualified_referrals')
                            ->label('Số người hợp lệ')
                            ->numeric()
                            ->required(),
                        TextInput::make('reward_amount')
                            ->label('Tiền thưởng')
                            ->numeric()
                            ->required(),
                        Select::make('unit')
                            ->label('Đơn vị')
                            ->options(EnumPresenter::options(UnitTransaction::class))
                            ->required(),
                        Toggle::make('is_active')
                            ->label('Kích hoạt'),
                        DateTimePicker::make('effective_from')
                            ->label('Hiệu lực từ'),
                        DateTimePicker::make('effective_to')
                            ->label('Hiệu lực đến'),
                        Textarea::make('note')
                            ->label('Ghi chú')
                            ->rows(3)
                            ->columnSpanFull(),
                    ])
                    ->columns(2),
            ]);
    }
}
