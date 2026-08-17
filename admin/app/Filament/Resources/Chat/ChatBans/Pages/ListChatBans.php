<?php

namespace App\Filament\Resources\Chat\ChatBans\Pages;

use App\Filament\Resources\Chat\ChatBans\ChatBanResource;
use Filament\Resources\Pages\ListRecords;

class ListChatBans extends ListRecords
{
    protected static string $resource = ChatBanResource::class;
}
