<?php

namespace App\Filament\Resources\Wheel\Rewards;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\Wheel\Rewards\Pages\ListWheelRewards;
use App\Models\Wheel\WheelReward;
use BackedEnum;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;
use UnitEnum;

class WheelRewardResource extends BaseResource
{
    protected static ?string $model = WheelReward::class;

    protected static UnitEnum|string|null $navigationGroup = 'Sự kiện vòng quay';

    protected static ?string $navigationLabel = 'Phần thưởng';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedBanknotes;

    protected static bool $canCreateRecords = false;

    protected static bool $canUpdateRecords = false;

    protected static bool $canDeleteRecords = false;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'wheel.rewards';
    }

    public static function form(Schema $schema): Schema
    {
        return $schema->components([]);
    }

    public static function table(Table $table): Table
    {
        return $table->columns([
            TextColumn::make('id')->label('ID')->sortable(),
            TextColumn::make('session.public_id')->label('Session')->limit(14)->copyable(),
            TextColumn::make('user.id')->label('User ID')->sortable(),
            TextColumn::make('user.name')->label('Người chơi')->searchable(),
            TextColumn::make('round_no')->label('Lượt')->badge(),
            TextColumn::make('amount')->label('Thưởng VND')->numeric(decimalPlaces: 0, locale: 'vi')->sortable(),
            TextColumn::make('status')->label('Trạng thái')->badge()->color(fn (string $state): string => $state === 'paid' ? 'success' : 'warning'),
            TextColumn::make('wallet_ledger_entry_id')->label('Ledger ID')->copyable(),
            TextColumn::make('paid_at')->label('Cộng lúc')->dateTime('d/m/Y H:i:s', timezone: config('app.timezone')),
        ])->defaultSort('id', 'desc')->poll(3000);
    }

    public static function getPages(): array
    {
        return ['index' => ListWheelRewards::route('/')];
    }
}
