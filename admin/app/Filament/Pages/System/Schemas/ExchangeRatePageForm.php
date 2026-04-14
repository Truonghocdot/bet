<?php

namespace App\Filament\Pages\System\Schemas;

use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\Textarea;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Schemas\Components\Section;
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

            Section::make('Cấu hình NOWPayments')
                ->description('Thông số kết nối cổng thanh toán Crypto.')
                ->schema([
                    TextInput::make('nowpayments_api_key')
                        ->label('NOWPayments API Key')
                        ->password()
                        ->revealable()
                        ->helperText('Lấy từ dashboard nowpayments.io'),
                    TextInput::make('nowpayments_ipn_secret')
                        ->label('IPN Secret Key')
                        ->password()
                        ->revealable()
                        ->helperText('Dùng để xác thực Webhook gửi từ NOWPayments.'),
                    TextInput::make('nowpayments_payout_wallet')
                        ->label('Ví nhận tiền (Payout Wallet)')
                        ->helperText('Địa chỉ ví USDT TRC20 để nhận tiền rút.'),
                    Toggle::make('nowpayments_sandbox')
                        ->label('Chế độ Sandbox (Test)')
                        ->helperText('Bật nếu đang dùng môi trường thử nghiệm.'),
                ])
                ->columns(2),

            Section::make('Cấu hình Hỗ trợ & Mạng xã hội')
                ->description('Thông tin liên hệ CSKH hiển thị trên ứng dụng.')
                ->schema([
                    TextInput::make('telegram_cskh_link')
                        ->label('Link Telegram CSKH')
                        ->url()
                        ->placeholder('https://t.me/your_cskh')
                        ->helperText('Đường dẫn trực tiếp đến tài khoản hoặc group Telegram hỗ trợ.'),
                ])
                ->columns(2),
        ]);
    }
}
