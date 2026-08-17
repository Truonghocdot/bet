<?php

namespace App\Filament\Resources\Chat\BotTemplates\Pages;

use App\Filament\Resources\Chat\BotTemplates\ChatBotTemplateResource;
use Filament\Resources\Pages\ListRecords;

class ListChatBotTemplates extends ListRecords
{
    protected static string $resource = ChatBotTemplateResource::class;
}
