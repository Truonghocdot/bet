<?php

namespace App\Filament\Resources\Chat\BotTemplates\Schemas;

use Filament\Forms\Components\Select;
use Filament\Forms\Components\Textarea;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Schemas\Schema;

class ChatBotTemplateForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Select::make('bot_profile_id')->label('Bot cố định')->relationship('botProfile', 'display_name')->searchable(),
            Textarea::make('body')->label('Nội dung')->required()->maxLength(280)->rows(4)->columnSpanFull(),
            TextInput::make('category')->label('Nhóm')->default('general')->required()->maxLength(60),
            TextInput::make('language')->label('Ngôn ngữ')->default('vi')->required()->maxLength(12),
            Toggle::make('active')->label('Đang bật')->default(true),
        ]);
    }
}
