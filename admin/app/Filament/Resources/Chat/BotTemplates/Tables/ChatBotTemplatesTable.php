<?php

namespace App\Filament\Resources\Chat\BotTemplates\Tables;

use Filament\Tables\Columns\IconColumn;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;

class ChatBotTemplatesTable
{
    public static function configure(Table $table): Table
    {
        return $table->columns([
            TextColumn::make('id')->label('ID')->sortable(),
            TextColumn::make('body')->label('Nội dung')->limit(100)->searchable(),
            TextColumn::make('category')->label('Nhóm')->sortable(),
            TextColumn::make('botProfile.display_name')->label('Bot'),
            IconColumn::make('active')->label('Bật')->boolean(),
            TextColumn::make('usage_count')->label('Đã dùng')->sortable(),
            TextColumn::make('last_used_at')->label('Lần cuối')->dateTime('d/m/Y H:i'),
        ])->defaultSort('id', 'desc');
    }
}
