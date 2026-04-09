<?php

namespace App\Filament\Resources\Affiliate\AffiliateReferrals\Schemas;

use App\Enum\Affiliate\AffiliateReferralStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Schemas\Components\Section;
use Filament\Schemas\Schema;

class AffiliateReferralForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema
            ->components([
                Section::make('Người được mời')
                    ->schema([
                        Select::make('affiliate_profile_id')
                            ->label('Hồ sơ affiliate')
                            ->relationship('affiliateProfile', 'ref_code')
                            ->searchable()
                            ->preload()
                            ->required(),
                        Select::make('referrer_user_id')
                            ->label('Người mời')
                            ->relationship('referrerUser', 'name')
                            ->searchable()
                            ->preload()
                            ->required(),
                        Select::make('referred_user_id')
                            ->label('Người được mời')
                            ->relationship('referredUser', 'name')
                            ->searchable()
                            ->preload()
                            ->required(),
                        Select::make('affiliate_link_id')
                            ->label('Liên kết theo dõi')
                            ->relationship('affiliateLink', 'tracking_code')
                            ->searchable()
                            ->preload(),
                        Select::make('first_deposit_transaction_id')
                            ->label('Giao dịch nạp đầu')
                            ->relationship('firstDepositTransaction', 'id')
                            ->searchable()
                            ->preload(),
                        TextInput::make('first_deposit_amount')
                            ->label('Số tiền nap dau')
                            ->numeric(),
                        DateTimePicker::make('qualified_at')
                            ->label('Đạt điều kiện lúc'),
                        Select::make('status')
                            ->label('Trạng thái')
                            ->options(EnumPresenter::options(AffiliateReferralStatus::class))
                            ->required(),
                    ])
                    ->columns(2),
            ]);
    }
}
