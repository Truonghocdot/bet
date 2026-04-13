<?php

namespace App\Filament\Resources\System\NewsArticles\Schemas;

use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\FileUpload;
use Filament\Forms\Components\Textarea;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Schemas\Components\Section;
use Filament\Schemas\Schema;

class NewsArticleForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Nội dung tin')
                ->schema([
                    TextInput::make('title')
                        ->label('Tiêu đề')
                        ->required()
                        ->maxLength(200),
                    TextInput::make('slug')
                        ->label('Slug')
                        ->maxLength(220)
                        ->helperText('Để trống để hệ thống tự sinh từ tiêu đề.'),
                    Textarea::make('excerpt')
                        ->label('Tóm tắt')
                        ->rows(3)
                        ->columnSpanFull(),
                    Textarea::make('content')
                        ->label('Nội dung')
                        ->rows(12)
                        ->required()
                        ->columnSpanFull(),
                    FileUpload::make('cover_image_path')
                        ->label('Ảnh đại diện')
                        ->disk('public')
                        ->directory('news')
                        ->image()
                        ->imageEditor()
                        ->helperText('Ảnh upload sẽ tự chuyển sang .webp khi lưu.'),
                ])
                ->columns(2),
            Section::make('Phát hành')
                ->schema([
                    Toggle::make('is_published')
                        ->label('Đã phát hành')
                        ->default(false),
                    DateTimePicker::make('published_at')
                        ->label('Thời gian phát hành')
                        ->seconds(false),
                ])
                ->columns(2),
        ]);
    }
}

