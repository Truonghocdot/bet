<?php

namespace App\Filament\Resources\Chat\ChatMessages\Pages;

use App\Filament\Resources\Chat\ChatMessages\ChatMessageResource;
use Filament\Resources\Pages\ListRecords;

class ListChatMessages extends ListRecords
{
    protected static string $resource = ChatMessageResource::class;
}
