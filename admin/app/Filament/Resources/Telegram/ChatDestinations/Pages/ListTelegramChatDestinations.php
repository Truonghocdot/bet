<?php

namespace App\Filament\Resources\Telegram\ChatDestinations\Pages;

use App\Filament\Resources\Telegram\ChatDestinations\TelegramChatDestinationResource;
use Filament\Resources\Pages\ListRecords;

class ListTelegramChatDestinations extends ListRecords
{
    protected static string $resource = TelegramChatDestinationResource::class;
}
