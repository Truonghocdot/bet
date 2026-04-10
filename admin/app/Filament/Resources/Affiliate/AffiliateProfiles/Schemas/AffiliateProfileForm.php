<?php

namespace App\Filament\Resources\Affiliate\AffiliateProfiles\Schemas;

use App\Enum\Affiliate\AffiliateProfileStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Schemas\Components\Section;
use Filament\Schemas\Schema;

class AffiliateProfileForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema
            ->components([
                Section::make('Hồ sơ affiliate')
                    ->schema([
                        Select::make('user_id')
                            ->label('Người dùng')
                            ->relationship('user', 'name')
                            ->searchable()
                            ->preload()
                            ->required(),
                        TextInput::make('ref_code')
                            ->label('Mã giới thiệu')
                            ->disabled()
                            ->dehydrated(false)
                            ->helperText('Mã giới thiệu được hệ thống tự sinh và không nhập tay.')
                            ->maxLength(50),
                        TextInput::make('ref_link')
                            ->label('Link giới thiệu')
                            ->disabled()
                            ->dehydrated(false)
                            ->helperText('Link giới thiệu được tạo tự động từ mã giới thiệu.')
                            ->maxLength(255),
                        Select::make('status')
                            ->label('Trạng thái')
                            ->options(EnumPresenter::options(AffiliateProfileStatus::class))
                            ->required(),
                        Select::make('approved_by')
                            ->label('Duyệt boi')
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
