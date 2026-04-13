<?php

namespace App\Filament\Resources\System\NewsArticles\Tables;

use Filament\Actions\BulkActionGroup;
use Filament\Actions\CreateAction;
use Filament\Actions\DeleteBulkAction;
use Filament\Actions\EditAction;
use Filament\Tables\Columns\IconColumn;
use Filament\Tables\Columns\ImageColumn;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\TernaryFilter;
use Filament\Tables\Table;

class NewsArticlesTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                ImageColumn::make('cover_image_path')
                    ->label('Ảnh')
                    ->disk('public')
                    ->square()
                    ->height(54),
                TextColumn::make('title')->label('Tiêu đề')->searchable()->limit(80),
                TextColumn::make('slug')->label('Slug')->toggleable(),
                IconColumn::make('is_published')->label('Phát hành')->boolean(),
                TextColumn::make('published_at')->label('Phát hành lúc')->dateTime()->sortable(),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
            ])
            ->filters([
                TernaryFilter::make('is_published')->label('Đã phát hành'),
            ])
            ->recordActions([
                EditAction::make(),
            ])
            ->headerActions([
                CreateAction::make()->label('Tạo tin tức'),
            ])
            ->defaultSort('id', 'desc')
            ->toolbarActions([
                BulkActionGroup::make([
                    DeleteBulkAction::make(),
                ]),
            ]);
    }
}

