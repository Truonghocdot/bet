<?php

namespace App\Filament\Resources\Affiliate\AffiliateRewardLogs\Schemas;

use App\Enum\Affiliate\AffiliateRewardStatus;
use App\Enum\Wallet\UnitTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Schemas\Schema;

class AffiliateRewardLogForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema
            ->components([
                Section::make('Lịch sử thưởng')
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
                        Select::make('setting_id')
                            ->label('Cấu hình')
                            ->relationship('setting', 'name')
                            ->searchable()
                            ->preload()
                            ->required(),
                        TextInput::make('required_qualified_referrals')
                            ->label('Số người yêu cầu')
                            ->numeric()
                            ->required(),
                        TextInput::make('actual_qualified_referrals')
                            ->label('Số người thực tế')
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
                        Select::make('status')
                            ->label('Trạng thái')
                            ->options(EnumPresenter::options(AffiliateRewardStatus::class))
                            ->required(),
                        Select::make('wallet_ledger_entry_id')
                            ->label('Sổ cái')
                            ->relationship('walletLedgerEntry', 'id')
                            ->searchable()
                            ->preload(),
                        DateTimePicker::make('granted_at')
                            ->label('Trả thưởng lúc'),
                    ])
                    ->columns(2),
            ]);
    }
}
