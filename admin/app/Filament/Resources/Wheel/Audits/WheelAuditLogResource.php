<?php

namespace App\Filament\Resources\Wheel\Audits;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\Wheel\Audits\Pages\ListWheelAuditLogs;
use App\Models\Wheel\WheelAuditLog;
use BackedEnum;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;
use UnitEnum;

class WheelAuditLogResource extends BaseResource
{
    protected static ?string $model = WheelAuditLog::class;

    protected static UnitEnum|string|null $navigationGroup = 'Sự kiện vòng quay';

    protected static ?string $navigationLabel = 'Nhật ký sự kiện';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedClipboardDocumentList;

    protected static bool $canCreateRecords = false;

    protected static bool $canUpdateRecords = false;

    protected static bool $canDeleteRecords = false;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'wheel.audits';
    }

    public static function form(Schema $schema): Schema
    {
        return $schema->components([]);
    }

    public static function table(Table $table): Table
    {
        return $table->columns([
            TextColumn::make('id')->label('ID')->sortable(),
            TextColumn::make('action')->label('Hành động')->badge()->searchable(),
            TextColumn::make('campaign_id')->label('Campaign'),
            TextColumn::make('invitation_id')->label('Invitation'),
            TextColumn::make('session_id')->label('Session'),
            TextColumn::make('actor_user_id')->label('Người thao tác'),
            TextColumn::make('ip_address')->label('IP'),
            TextColumn::make('created_at')->label('Thời gian')->dateTime('d/m/Y H:i:s', timezone: config('app.timezone'))->sortable(),
        ])->defaultSort('id', 'desc');
    }

    public static function getPages(): array
    {
        return ['index' => ListWheelAuditLogs::route('/')];
    }
}
