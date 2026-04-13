<?php

namespace App\Filament\Resources\System\Notifications\Schemas;

use App\Enum\Notification\NotificationAudience;
use App\Enum\Notification\NotificationStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\Textarea;
use Filament\Forms\Components\TextInput;
use Filament\Schemas\Components\Section;
use Filament\Schemas\Components\Utilities\Get;
use Filament\Schemas\Schema;

class NotificationForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Nội dung thông báo')
                ->schema([
                    TextInput::make('title')
                        ->label('Tiêu đề')
                        ->required()
                        ->maxLength(200),
                    Textarea::make('body')
                        ->label('Nội dung')
                        ->required()
                        ->rows(6)
                        ->columnSpanFull(),
                ])
                ->columns(2),

            Section::make('Thiết lập phát hành')
                ->schema([
                    Select::make('status')
                        ->label('Trạng thái')
                        ->options(EnumPresenter::options(NotificationStatus::class))
                        ->default(NotificationStatus::DRAFT->value)
                        ->required(),
                    Select::make('audience')
                        ->label('Đối tượng nhận')
                        ->options(EnumPresenter::options(NotificationAudience::class))
                        ->default(NotificationAudience::ALL->value)
                        ->required()
                        ->live(),
                    DateTimePicker::make('publish_at')
                        ->label('Thời gian phát hành')
                        ->seconds(false),
                    DateTimePicker::make('expires_at')
                        ->label('Hết hạn lúc')
                        ->seconds(false),
                    Select::make('targetUsers')
                        ->label('Người dùng chỉ định')
                        ->relationship('targetUsers', 'name')
                        ->multiple()
                        ->searchable()
                        ->preload()
                        ->helperText('Chỉ áp dụng khi đối tượng nhận là "Người dùng chỉ định".')
                        ->visible(fn (Get $get): bool => (int) $get('audience') === NotificationAudience::USERS->value)
                        ->required(fn (Get $get): bool => (int) $get('audience') === NotificationAudience::USERS->value),
                ])
                ->columns(2),
        ]);
    }
}

