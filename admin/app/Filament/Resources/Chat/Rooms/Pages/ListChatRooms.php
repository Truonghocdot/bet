<?php

namespace App\Filament\Resources\Chat\Rooms\Pages;

use App\Filament\Resources\Chat\Rooms\ChatRoomResource;
use Filament\Resources\Pages\ListRecords;

class ListChatRooms extends ListRecords
{
    protected static string $resource = ChatRoomResource::class;
}
