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
                    FileUpload::make('image_path')
                        ->label('Ảnh banner')
                        ->disk('public')
                        ->directory('banners')
                        ->image()
                        ->imageEditor()
                        ->required()
                        ->columnSpanFull(),
                    TextInput::make('sort_order')
                        ->label('Thứ tự hiển thị')
                        ->numeric()
                        ->default(0)
                        ->required(),
                ])
                ->columns(1)
                ->columnSpanFull(),
        ]);
    }
}

