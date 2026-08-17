<?php

namespace App\Filament\Resources\Chat\BotProfiles\Pages;

use App\Filament\Resources\Chat\BotProfiles\ChatBotProfileResource;
use Filament\Resources\Pages\ListRecords;

class ListChatBotProfiles extends ListRecords
{
    protected static string $resource = ChatBotProfileResource::class;
}
