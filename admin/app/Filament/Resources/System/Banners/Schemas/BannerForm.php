<?php

namespace App\Filament\Resources\System\Banners\Schemas;

use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\FileUpload;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Schemas\Components\Section;
use Filament\Schemas\Schema;

class BannerForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Thông tin banner')
                ->schema([
                    TextInput::make('title')
                        ->label('Tiêu đề')
                        ->required()
                        ->maxLength(160),
                    FileUpload::make('image_path')
                        ->label('Ảnh banner')
                        ->disk('public')
                        ->directory('banners')
                        ->image()
                        ->imageEditor()
                        ->required()
                        ->helperText('Ảnh upload sẽ được tự động chuyển sang định dạng .webp khi lưu.'),
                    TextInput::make('link_url')
                        ->label('Link điều hướng')
                        ->maxLength(255)
                        ->url(),
                    TextInput::make('sort_order')
                        ->label('Thứ tự hiển thị')
                        ->numeric()
                        ->default(0),
                    Toggle::make('is_active')
                        ->label('Đang hoạt động')
                        ->default(true),
                    DateTimePicker::make('start_at')
                        ->label('Hiệu lực từ')
                        ->seconds(false),
                    DateTimePicker::make('end_at')
                        ->label('Hiệu lực đến')
                        ->seconds(false),
                ])
                ->columns(2)
                ->columnSpanFull(),
        ]);
    }
}

