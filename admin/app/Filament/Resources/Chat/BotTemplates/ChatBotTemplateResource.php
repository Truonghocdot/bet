<?php

namespace App\Filament\Resources\Chat\BotTemplates;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\Chat\BotTemplates\Pages\CreateChatBotTemplate;
use App\Filament\Resources\Chat\BotTemplates\Pages\EditChatBotTemplate;
use App\Filament\Resources\Chat\BotTemplates\Pages\ListChatBotTemplates;
use App\Filament\Resources\Chat\BotTemplates\Schemas\ChatBotTemplateForm;
use App\Filament\Resources\Chat\BotTemplates\Tables\ChatBotTemplatesTable;
use App\Models\Chat\ChatBotTemplate;
use BackedEnum;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;
use UnitEnum;

class ChatBotTemplateResource extends BaseResource
{
    protected static ?string $model = ChatBotTemplate::class;

    protected static UnitEnum|string|null $navigationGroup = 'Chat';

    protected static ?string $navigationLabel = 'Câu mẫu bot';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedDocumentText;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'chat.bot-templates';
    }

    public static function form(Schema $schema): Schema
    {
        return ChatBotTemplateForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return ChatBotTemplatesTable::configure($table);
    }

    public static function getPages(): array
    {
        return ['index' => ListChatBotTemplates::route('/'), 'create' => CreateChatBotTemplate::route('/create'), 'edit' => EditChatBotTemplate::route('/{record}/edit')];
    }
}
