<?php

namespace App\Filament\Resources\Chat\Rooms;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\Chat\Rooms\Pages\EditChatRoom;
use App\Filament\Resources\Chat\Rooms\Pages\ListChatRooms;
use App\Models\Chat\ChatRoom;
use BackedEnum;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Columns\IconColumn;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;
use UnitEnum;

class ChatRoomResource extends BaseResource
{
    protected static ?string $model = ChatRoom::class;

    protected static bool $canCreateRecords = false;

    protected static bool $canDeleteRecords = false;

    protected static bool $canDeleteAnyRecords = false;

    protected static UnitEnum|string|null $navigationGroup = 'Chat';

    protected static ?string $navigationLabel = 'Cấu hình phòng';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedCog6Tooth;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'chat.rooms';
    }

    public static function form(Schema $schema): Schema
    {
        return $schema->components([
            TextInput::make('code')->label('Mã phòng')->disabled()->dehydrated(false),
            TextInput::make('name')->label('Tên phòng')->required()->maxLength(120),
            Toggle::make('enabled')->label('Cho phép chat')->required(),
        ]);
    }

    public static function table(Table $table): Table
    {
        return $table->columns([
            TextColumn::make('code')->label('Mã')->sortable(),
            TextColumn::make('name')->label('Tên phòng'),
            IconColumn::make('enabled')->label('Mở')->boolean(),
            TextColumn::make('bot_active_until')->label('Bot chạy đến')->dateTime('d/m/Y H:i:s', timezone: config('app.timezone')),
            TextColumn::make('updated_at')->label('Cập nhật')->dateTime('d/m/Y H:i'),
        ]);
    }

    public static function getPages(): array
    {
        return [
            'index' => ListChatRooms::route('/'),
            'edit' => EditChatRoom::route('/{record}/edit'),
        ];
    }
}
