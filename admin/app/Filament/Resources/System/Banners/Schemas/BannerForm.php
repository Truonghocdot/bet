<?php

namespace App\Filament\Resources\System\Banners\Schemas;

use Filament\Forms\Components\Hidden;
use Filament\Forms\Components\FileUpload;
use Filament\Forms\Components\TextInput;
use Filament\Schemas\Components\Section;
use Filament\Schemas\Schema;

class BannerForm
{
    public static function configure(
        Schema $schema,
        string $placement = 'home',
        string $sectionTitle = 'Thông tin banner',
        string $imageLabel = 'Ảnh banner',
    ): Schema
    {
        return $schema->components([
            Section::make($sectionTitle)
                ->schema([
                    Hidden::make('placement')
                        ->default($placement)
                        ->required(),
                    FileUpload::make('image_path')
                        ->label($imageLabel)
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
