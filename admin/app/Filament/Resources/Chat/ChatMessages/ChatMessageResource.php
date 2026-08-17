<?php

namespace App\Filament\Resources\Chat\ChatMessages;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\Chat\ChatMessages\Pages\ListChatMessages;
use App\Filament\Resources\Chat\ChatMessages\Schemas\ChatMessageForm;
use App\Filament\Resources\Chat\ChatMessages\Tables\ChatMessagesTable;
use App\Models\Chat\ChatMessage;
use BackedEnum;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;
use UnitEnum;

class ChatMessageResource extends BaseResource
{
    protected static ?string $model = ChatMessage::class;

    protected static bool $canCreateRecords = false;

    protected static bool $canUpdateRecords = false;

    protected static bool $canDeleteRecords = false;

    protected static bool $canDeleteAnyRecords = false;

    protected static UnitEnum|string|null $navigationGroup = 'Chat';

    protected static ?string $navigationLabel = 'Tin nhắn';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedChatBubbleLeftRight;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'chat.messages';
    }

    public static function form(Schema $schema): Schema
    {
        return ChatMessageForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return ChatMessagesTable::configure($table);
    }

    public static function getPages(): array
    {
        return ['index' => ListChatMessages::route('/')];
    }
}
