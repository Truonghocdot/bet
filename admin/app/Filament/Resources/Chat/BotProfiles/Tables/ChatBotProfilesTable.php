<?php

namespace App\Filament\Resources\Chat\BotProfiles\Tables;

use Filament\Tables\Columns\IconColumn;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;

class ChatBotProfilesTable
{
    public static function configure(Table $table): Table
    {
        return $table->columns([
            TextColumn::make('id')->label('ID')->sortable(),
            TextColumn::make('display_name')->label('Tên')->searchable(),
            IconColumn::make('active')->label('Bật')->boolean(),
            TextColumn::make('sort_order')->label('Thứ tự')->sortable(),
        ])->defaultSort('sort_order');
    }
}
