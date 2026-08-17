<?php

namespace App\Filament\Resources\Chat\BotProfiles\Schemas;

use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Schemas\Schema;

class ChatBotProfileForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            TextInput::make('display_name')->label('Tên hiển thị')->required()->maxLength(120),
            TextInput::make('avatar_path')->label('Avatar')->maxLength(255),
            TextInput::make('sort_order')->label('Thứ tự')->numeric()->default(0)->required(),
            Toggle::make('active')->label('Đang bật')->default(true),
        ]);
    }
}
