<?php

namespace App\Filament\Resources\System\Notifications;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\System\Notifications\Pages\CreateNotification;
use App\Filament\Resources\System\Notifications\Pages\EditNotification;
use App\Filament\Resources\System\Notifications\Pages\ListNotifications;
use App\Filament\Resources\System\Notifications\Schemas\NotificationForm;
use App\Filament\Resources\System\Notifications\Tables\NotificationsTable;
use App\Models\Notification\Notification;
use BackedEnum;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;
use UnitEnum;

class NotificationResource extends BaseResource
{
    protected static ?string $model = Notification::class;
    protected static UnitEnum|string|null $navigationGroup = 'Thiết lập';
    protected static ?string $navigationLabel = 'Thông báo hệ thống';
    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedBellAlert;
    protected static ?string $recordTitleAttribute = 'title';

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'system.notifications';
    }

    public static function form(Schema $schema): Schema
    {
        return NotificationForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return NotificationsTable::configure($table);
    }

    public static function getRelations(): array
    {
        return [];
    }

    public static function getPages(): array
    {
        return [
            'index' => ListNotifications::route('/'),
            'create' => CreateNotification::route('/create'),
            'edit' => EditNotification::route('/{record}/edit'),
        ];
    }
}

