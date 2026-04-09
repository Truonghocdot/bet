<?php

namespace App\Filament\Resources\Affiliate\AffiliateLinks\Schemas;

use App\Enum\Affiliate\AffiliateLinkStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Schemas\Components\Section;
use Filament\Schemas\Schema;

class AffiliateLinkForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema
            ->components([
                Section::make('Liên kết theo dõi')
                    ->schema([
                        Select::make('affiliate_profile_id')
                            ->label('Hồ sơ affiliate')
                            ->relationship('affiliateProfile', 'ref_code')
                            ->searchable()
                            ->preload()
                            ->required(),
                        TextInput::make('campaign_name')
                            ->label('Tên chiến dịch')
                            ->required()
                            ->maxLength(100),
                        TextInput::make('tracking_code')
                            ->label('Mã tracking')
                            ->required()
                            ->maxLength(100),
                        TextInput::make('landing_url')
                            ->label('URL đích')
                            ->required()
                            ->maxLength(255),
                        Select::make('status')
                            ->label('Trạng thái')
                            ->options(EnumPresenter::options(AffiliateLinkStatus::class))
                            ->required(),
                    ])
                    ->columns(2),
            ]);
    }
}
