<?php

namespace App\Filament\Pages\System\Schemas;

use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\Section;
use Filament\Forms\Components\Textarea;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Schemas\Schema;

class ExchangeRatePageForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Tỷ giá USDT/VND')
                ->description('Dữ liệu này được lưu vào Redis để service Gin và Laravel cùng tái sử dụng.')
                ->schema([
                    TextInput::make('code')
                        ->label('Mã cấu hình')
                        ->disabled(),
                    TextInput::make('base_currency')
                        ->label('Tiền gốc')
                        ->disabled(),
                    TextInput::make('quote_currency')
                        ->label('Tiền quy đổi')
                        ->disabled(),
                    TextInput::make('rate')
                        ->label('Tỷ giá áp dụng')
                        ->numeric()
                        ->required()
                        ->step('0.000001')
                        ->helperText('Giá đang dùng để quy đổi trong hệ thống.'),
                    TextInput::make('source_rate')
                        ->label('Tỷ giá nguồn')
                        ->disabled(),
                    Toggle::make('auto_sync')
                        ->label('Tự động đồng bộ từ nguồn'),
                    TextInput::make('source_name')
                        ->label('Nguồn cập nhật')
                        ->disabled(),
                    TextInput::make('redis_connection')
                        ->label('Kết nối Redis chia sẻ')
                        ->disabled()
                        ->helperText('Gin và Laravel đọc cùng một connection này.'),
                    TextInput::make('redis_key')
                        ->label('Redis key raw JSON')
                        ->disabled()
                        ->helperText('Gin đọc trực tiếp key này, không qua serializer Laravel.'),
                    TextInput::make('cache_store')
                        ->label('Cache store Laravel')
                        ->disabled(),
                    TextInput::make('cache_key')
                        ->label('Cache key Laravel')
                        ->disabled(),
                    DateTimePicker::make('last_synced_at')
                        ->label('Đồng bộ lần cuối')
                        ->disabled(),
                    Textarea::make('note')
                        ->label('Ghi chú')
                        ->rows(4)
                        ->columnSpanFull(),
                ])
                ->columns(2),
        ]);
    }
}
