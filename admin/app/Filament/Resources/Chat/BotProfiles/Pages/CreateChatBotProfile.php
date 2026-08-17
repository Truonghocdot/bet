<?php

namespace App\Filament\Resources\Chat\BotProfiles\Pages;

use App\Filament\Resources\Chat\BotProfiles\ChatBotProfileResource;
use Filament\Resources\Pages\CreateRecord;

class CreateChatBotProfile extends CreateRecord
{
    protected static string $resource = ChatBotProfileResource::class;
}
