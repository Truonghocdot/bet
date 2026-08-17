<?php

namespace App\Filament\Resources\Chat\ModerationActions\Pages;

use App\Filament\Resources\Chat\ModerationActions\ChatModerationActionResource;
use Filament\Resources\Pages\ListRecords;

class ListChatModerationActions extends ListRecords
{
    protected static string $resource = ChatModerationActionResource::class;
}
