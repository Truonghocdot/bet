<?php

namespace App\Filament\Resources\System\NewsArticles\Schemas;

use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\FileUpload;
use Filament\Forms\Components\RichEditor;
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
                        ->helperText('Đoạn mô tả ngắn hiển thị ở danh sách tin. Có thể để trống nếu chỉ muốn dùng nội dung bài viết.')
                        ->columnSpanFull(),
                    RichEditor::make('content')
                        ->label('Nội dung')
                        ->required()
                        ->toolbarButtons([
                            'blockquote',
                            'bold',
                            'bulletList',
                            'codeBlock',
                            'h2',
                            'h3',
                            'italic',
                            'link',
                            'orderedList',
                            'redo',
                            'strike',
                            'underline',
                            'undo',
                        ])
                        ->fileAttachmentsDisk('public')
                        ->fileAttachmentsDirectory('news/editor')
                        ->helperText('Nhập nội dung bằng trình soạn thảo trực quan. Hệ thống sẽ hiển thị đúng định dạng trên app.')
                        ->columnSpanFull(),
                    FileUpload::make('cover_image_path')
                        ->label('Thumbnail tin tức')
                        ->disk('public')
                        ->directory('news')
                        ->image()
                        ->imageEditor()
                        ->helperText('Ảnh upload sẽ tự chuyển sang .webp khi lưu. Đây là ảnh thumbnail hiển thị ở danh sách và chi tiết tin.'),
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
