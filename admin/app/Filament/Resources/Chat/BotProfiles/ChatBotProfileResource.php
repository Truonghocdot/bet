<?php

namespace App\Filament\Resources\Chat\BotProfiles;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\Chat\BotProfiles\Pages\CreateChatBotProfile;
use App\Filament\Resources\Chat\BotProfiles\Pages\EditChatBotProfile;
use App\Filament\Resources\Chat\BotProfiles\Pages\ListChatBotProfiles;
use App\Filament\Resources\Chat\BotProfiles\Schemas\ChatBotProfileForm;
use App\Filament\Resources\Chat\BotProfiles\Tables\ChatBotProfilesTable;
use App\Models\Chat\ChatBotProfile;
use BackedEnum;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Table;
use UnitEnum;

class ChatBotProfileResource extends BaseResource
{
    protected static ?string $model = ChatBotProfile::class;

    protected static UnitEnum|string|null $navigationGroup = 'Chat';

    protected static ?string $navigationLabel = 'Bot profile';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedUserGroup;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'chat.bot-profiles';
    }

    public static function form(Schema $schema): Schema
    {
        return ChatBotProfileForm::configure($schema);
    }

    public static function table(Table $table): Table
    {
        return ChatBotProfilesTable::configure($table);
    }

    public static function getPages(): array
    {
        return ['index' => ListChatBotProfiles::route('/'), 'create' => CreateChatBotProfile::route('/create'), 'edit' => EditChatBotProfile::route('/{record}/edit')];
    }
}
